package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"proxy-provider/cache"
	"proxy-provider/core"
	"proxy-provider/geosite"
	"proxy-provider/util"
	"strings"
	"time"

	_ "proxy-provider/action"
	_ "proxy-provider/formatter"

	"github.com/gofrs/uuid/v5"
)

//go:embed public/index.html
var newPageHTML []byte

const editorComment = "# Created by WebUI Editor"

func writeJSONError(writer http.ResponseWriter, statusCode int, err error) {
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"ok":    false,
		"error": err.Error(),
	})
}

func buildEditorContent(content string) string {
	return fmt.Sprintf("%s\n# %s\n\n%s\n", editorComment, time.Now().Format(time.UnixDate), strings.TrimRight(content, "\n"))
}

func readEditorFormula(content []byte) (string, bool) {
	text := string(content)
	if !strings.HasPrefix(text, editorComment+"\n# ") {
		return "", false
	}
	bodyIndex := strings.Index(text, "\n\n")
	if bodyIndex < 0 {
		return "", false
	}
	return text[bodyIndex+2:], true
}

func formulaHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if editorEnabled && request.URL.Path == "/" && request.Method == http.MethodGet {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write(newPageHTML)
		return
	}

	filePath := filepath.Join(dir, filepath.Clean(request.URL.Path))
	if stat, err := os.Stat(filePath); err != nil {
		fmt.Println("file", filePath, "not found")
		writer.WriteHeader(http.StatusNotFound)
		return
	} else if stat.IsDir() {
		fmt.Println(filePath, "is a folder")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("read formula from", filePath, "failed:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	if request.URL.Query().Get("type") == "formula" {
		if formulaContent, ok := readEditorFormula(content); ok {
			writer.Write([]byte(formulaContent))
			return
		}
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	formula := &core.Formula{
		Name: filepath.Base(filePath),
	}
	if err := formula.Parse(content); err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	ctx := core.NewExecuteContext()

	if err := formula.Execute(ctx); err != nil {
		fmt.Println("error occurred when running formula:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	} else {
		ctx.Pipe(writer)
	}
}

type saveRequest struct {
	Password string `json:"password"`
	Content  string `json:"content"`
	Edit     string `json:"edit"`
}

func saveHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !editorEnabled {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	if password == "" {
		writeJSONError(writer, http.StatusForbidden, fmt.Errorf("save is disabled because -password is empty"))
		return
	}

	payload := &saveRequest{}
	if err := json.NewDecoder(request.Body).Decode(payload); err != nil {
		writeJSONError(writer, http.StatusBadRequest, err)
		return
	}

	if payload.Password != password {
		writeJSONError(writer, http.StatusUnauthorized, fmt.Errorf("invalid password"))
		return
	}

	formula := &core.Formula{Name: "editor"}
	if err := formula.Parse([]byte(payload.Content)); err != nil {
		writeJSONError(writer, http.StatusBadRequest, err)
		return
	}

	content := buildEditorContent(payload.Content)
	fileName := payload.Edit
	if fileName != "" {
		filePath := filepath.Join(dir, fileName)
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			writeJSONError(writer, http.StatusInternalServerError, err)
			return
		}
		if _, ok := readEditorFormula(existingContent); !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if err := util.WriteFileSafely(filePath, []byte(content), 0644); err != nil {
			writeJSONError(writer, http.StatusInternalServerError, err)
			return
		}
	} else {
		fileID, err := uuid.NewV4()
		if err != nil {
			writeJSONError(writer, http.StatusInternalServerError, err)
			return
		}

		fileName = fileID.String()
		filePath := filepath.Join(dir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			writeJSONError(writer, http.StatusInternalServerError, fmt.Errorf("file already exists: %s", fileName))
			return
		} else if !os.IsNotExist(err) {
			writeJSONError(writer, http.StatusInternalServerError, err)
			return
		}
		if err := util.WriteFileSafely(filePath, []byte(content), 0644); err != nil {
			writeJSONError(writer, http.StatusInternalServerError, err)
			return
		}
	}

	_ = json.NewEncoder(writer).Encode(map[string]any{
		"ok":       true,
		"filename": fileName,
		"path":     "/" + fileName,
	})
}

var dir string
var cacheDir string
var port int
var editorEnabled bool
var password string

func main() {
	cwd, _ := os.Getwd()
	flag.StringVar(&dir, "dir", cwd, "Directory to use as home directory for configuration files")
	flag.StringVar(&cacheDir, "cache-dir", filepath.Join(os.TempDir(), "cache"), "Directory to use as cache directory for fetched data")
	flag.BoolVar(&editorEnabled, "editor", false, "Enable the browser editor")
	flag.StringVar(&password, "password", "", "Password required to save formulas from the web editor")
	flag.IntVar(&port, "port", 15725, "Port to run the HTTP server on")
	flag.Parse()

	if editorEnabled && password == "" {
		fmt.Fprintln(os.Stderr, "-password is required when -editor is enabled")
		os.Exit(1)
	}

	cache.HTTPCacheFolder = cacheDir
	_ = os.MkdirAll(cache.HTTPCacheFolder, 0755)

	http.HandleFunc("/new", saveHandler)
	http.HandleFunc("/geosite/domains/{name}", geosite.ServeDomainSet)
	http.HandleFunc("/geosite/{name}", geosite.Serve)
	http.HandleFunc("/", formulaHandler)
	fmt.Printf("use formula from: %s\n", dir)
	fmt.Printf("use cache folder: %s\n", cacheDir)
	fmt.Printf("server started at http://0.0.0.0:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}
