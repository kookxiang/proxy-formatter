package action

import (
	"errors"
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
	core.RegisterAction("fetch-once", func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid action rule format, expected 'action <url>'")
		}
		return &FetchAction{URL: params[1], Once: true}, nil
	})
}

var fetchLock = &sync.Mutex{}

type FetchAction struct {
	URL  string
	Once bool
}

func (action *FetchAction) Execute(ctx *core.ExecuteContext) error {
	fetchLock.Lock()
	defer fetchLock.Unlock()

	urlObj, err := url.Parse(action.URL)
	if err != nil {
		return err
	}

	request := &http.Request{
		Method: http.MethodGet,
		URL:    urlObj,
		Header: ctx.ReqHeader,
	}

	body, header, err := cache.RequestWithCache(ctx, request, action.Once)
	if err != nil {
		return err
	}

	ctx.ResHeader = header.Clone()
	if action.Once {
		ctx.ResHeader.Del(core.HeaderSubscriptionUserinfo)
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
				Type:   proxy["type"].(string),
				Option: option,
			})
		}
	}

	return nil
}
