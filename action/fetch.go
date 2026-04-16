package action

import (
	"errors"
	"net/http"
	"net/url"
	"proxy-provider/cache"
	"proxy-provider/core"
	"sync"
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
	ctx.ResHeader.Del("Content-Disposition")

	return ctx.RegisterProxiesFromBytes(body)
}
