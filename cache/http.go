package cache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var HTTPCacheFolder = "cache"

func LoadResponse(url string) ([]byte, http.Header, bool) {
	urlHash := sha256.Sum256([]byte(url))
	localFilePath := filepath.Join(HTTPCacheFolder, hex.EncodeToString(urlHash[:]))
	needRefresh := false
	if stat, err := os.Stat(localFilePath); err != nil {
		return nil, nil, false
	} else if stat.ModTime().Add(10 * time.Minute).Before(time.Now()) {
		needRefresh = true
	}
	cacheFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Println("failed to open cache file", localFilePath, ":", err)
		return nil, nil, false
	}
	defer cacheFile.Close()
	header := http.Header{}
	var body []byte
	scanner := bufio.NewScanner(cacheFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			header.Add(key, value)
		}
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		body = append(body, line...)
		body = append(body, '\n')
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Println("error reading cache file", localFilePath, ":", err)
		return nil, nil, false
	}

	return body, header, needRefresh
}

func SaveResponse(url string, body []byte, header http.Header) {
	urlHash := sha256.Sum256([]byte(url))
	localFilePath := filepath.Join(HTTPCacheFolder, hex.EncodeToString(urlHash[:]))
	cacheFile, err := os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("failed to open cache file", localFilePath, ":", err)
		return
	}
	defer cacheFile.Close()
	for key, values := range header {
		for _, value := range values {
			cacheFile.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	cacheFile.WriteString("\n")
	cacheFile.Write(body)
}
