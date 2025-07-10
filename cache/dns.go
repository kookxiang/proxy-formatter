package cache

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var dnsCacheLock = &sync.RWMutex{}

var dnsCacheStore = make(map[string]*DNSCacheItem)

type DNSCacheItem struct {
	IPs    []net.IP
	expire time.Time
}

func GetDNSCache(resolver, host string) *DNSCacheItem {
	dnsCacheLock.RLock()
	defer dnsCacheLock.RUnlock()

	key := fmt.Sprintf("%s@%s", resolver, host)
	item, exists := dnsCacheStore[key]
	if !exists {
		return nil
	}
	return item
}

func SetDNSCache(resolver, host string, item *DNSCacheItem) {
	dnsCacheLock.Lock()
	defer dnsCacheLock.Unlock()

	key := fmt.Sprintf("%s@%s", resolver, host)
	item.expire = time.Now().Add(5 * time.Minute)
	dnsCacheStore[key] = item
}

func purgeExpiredDNSCacheRecord() {
	dnsCacheLock.Lock()
	defer dnsCacheLock.Unlock()
	for key, item := range dnsCacheStore {
		if time.Now().After(item.expire) {
			delete(dnsCacheStore, key)
		}
	}
}

func init() {
	go func() {
		timer := time.NewTimer(5 * time.Minute)
		for {
			select {
			case <-timer.C:
				purgeExpiredDNSCacheRecord()
			}
		}
	}()
}
