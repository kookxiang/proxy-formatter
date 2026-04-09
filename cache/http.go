package cache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"proxy-provider/core"
	"strings"
	"time"
)

var HTTPCacheFolder = "cache"

func RequestWithCache(ctx *core.ExecuteContext, request *http.Request, preferCache bool) ([]byte, http.Header, error) {
	body, header, needRefresh := getResponseCache(request.URL.String())
	if len(body) == 0 || (needRefresh && !preferCache) {
		var err error
		body, header, err = requestWithoutCache(ctx, request)
		if err != nil && needRefresh {
			fmt.Println("refreshing cache for url", request.URL, "failed:", err)
		} else if err != nil {
			return nil, nil, err
		} else {
			saveResponseCache(request.URL.String(), body, header)
		}
	}

	if preferCache {
		header.Del(core.HeaderSubscriptionUserinfo)
	} else if info := header.Get(core.HeaderSubscriptionUserinfo); info != "" {
		header.Set(core.HeaderSubscriptionUserinfo, info)
	}

	return body, header, nil
}

func requestWithoutCache(ctx *core.ExecuteContext, request *http.Request) ([]byte, http.Header, error) {
	client := &http.Client{}
	if ctx.Proxy != "" {
		proxy, err := url.Parse(ctx.Proxy)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid proxy URL %s: %v", ctx.Proxy, err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	fmt.Printf("requesting %s without cache\n", request.URL.Host)
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to fetch URL %s, status code: %d", request.URL, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, response.Header, nil
}

func getResponseCache(url string) ([]byte, http.Header, bool) {
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

func saveResponseCache(url string, body []byte, header http.Header) {
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
			fmt.Fprintf(cacheFile, "%s: %s\n", key, value)
		}
	}
	cacheFile.WriteString("\n")
	cacheFile.Write(body)
}
