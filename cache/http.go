package cache

import (
	"sync"
	"time"
)

var httpCacheLock = &sync.RWMutex{}

var httpCacheStore = make(map[string]*ResponseCacheItem)

type ResponseCacheItem struct {
	Data      []byte
	UsageInfo string
	expire    time.Time
}

func GetResponseCache(key string) *ResponseCacheItem {
	httpCacheLock.RLock()
	defer httpCacheLock.RUnlock()

	item, exists := httpCacheStore[key]
	if !exists {
		return nil
	}
	return item
}

func SetResponseCache(key string, item *ResponseCacheItem) {
	httpCacheLock.Lock()
	defer httpCacheLock.Unlock()

	item.expire = time.Now().Add(10 * time.Minute)
	httpCacheStore[key] = item
}

func purgeExpiredHTTPCacheData() {
	httpCacheLock.Lock()
	defer httpCacheLock.Unlock()
	for key, item := range httpCacheStore {
		if time.Now().After(item.expire) {
			delete(httpCacheStore, key)
		}
	}
}

func init() {
	go func() {
		timer := time.NewTimer(5 * time.Minute)
		for {
			select {
			case <-timer.C:
				purgeExpiredHTTPCacheData()
			}
		}
	}()
}
