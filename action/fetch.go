package action

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"proxy-provider/cache"
	"proxy-provider/core"
	"proxy-provider/util"
	"sync"

	"github.com/metacubex/mihomo/adapter/provider"
	"github.com/metacubex/mihomo/common/convert"
	"gopkg.in/yaml.v3"
)

func init() {
	core.RegisterAction("fetch", func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid action rule format, expected 'action <url>'")
		}
		return &FetchAction{URL: params[1]}, nil
	})
}

var fetchLock = &sync.Mutex{}

type FetchAction struct {
	URL string
}

func (action *FetchAction) Execute(ctx *core.ExecuteContext) error {
	fetchLock.Lock()
	defer fetchLock.Unlock()

	body, header, needRefresh := cache.LoadResponse(action.URL)
	if body == nil || len(body) == 0 || needRefresh {
		var err error
		body, header, err = action.doFetch()
		if err != nil && needRefresh {
			fmt.Println("refreshing cache for url", action.URL, "failed:", err)
		} else if err != nil {
			return err
		} else {
			cache.SaveResponse(action.URL, body, header)
		}
	}

	if header.Get("Subscription-Userinfo") != "" {
		ctx.WriteHeader("Subscription-Userinfo", header.Get("Subscription-Userinfo"))
	}

	schema := &provider.ProxySchema{}
	// try yaml first
	if err := yaml.Unmarshal(body, schema); err != nil {
		// when failed, try to parse as base64 encoded list
		proxies, err := convert.ConvertsV2Ray(body)
		if err != nil {
			return err
		}
		schema.Proxies = proxies
	}

	for _, proxy := range schema.Proxies {
		name, ok := proxy["name"].(string)
		if !ok || name == "" {
			return errors.New("proxy name is required in the schema")
		}
		if option, err := util.ParseProxyOptions(proxy); err != nil {
			return err
		} else {
			ctx.RegisterProxy(&core.ProxyItem{
				Name:   name,
				Option: option,
			})
		}
	}

	return nil
}

func (action *FetchAction) doFetch() ([]byte, http.Header, error) {
	urlObj, err := url.Parse(action.URL)
	if err != nil {
		return nil, nil, err
	}
	request := &http.Request{
		Method: http.MethodGet,
		URL:    urlObj,
		Header: http.Header{
			"User-Agent": []string{"clash.meta/1.19.11"},
		},
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, errors.New(fmt.Sprintf("failed to fetch URL %s, status code: %d", action.URL, response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, response.Header, nil
}

func (action *FetchAction) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum224([]byte(action.URL)))
}
