package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"proxy-provider/cache"
	"proxy-provider/core"
	"proxy-provider/geosite"

	_ "proxy-provider/action"
	_ "proxy-provider/formatter"
)

func formulaHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
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

var dir string
var cacheDir string
var port int

func main() {
	cwd, _ := os.Getwd()
	flag.StringVar(&dir, "dir", cwd, "Directory to use as home directory for configuration files")
	flag.StringVar(&cacheDir, "cache-dir", filepath.Join(os.TempDir(), "cache"), "Directory to use as cache directory for fetched data")
	flag.IntVar(&port, "port", 15725, "Port to run the HTTP server on")
	flag.Parse()

	cache.HTTPCacheFolder = cacheDir
	_ = os.MkdirAll(cache.HTTPCacheFolder, 0755)

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
