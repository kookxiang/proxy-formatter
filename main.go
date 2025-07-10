package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"proxy-provider/core"

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
var port int

func main() {
	cwd, _ := os.Getwd()
	flag.StringVar(&dir, "dir", cwd, "Directory to use as home directory for configuration files")
	flag.IntVar(&port, "port", 8080, "Port to run the HTTP server on")
	flag.Parse()

	http.HandleFunc("/", formulaHandler)
	fmt.Printf("server started at http://0.0.0.0:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}
