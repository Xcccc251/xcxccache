package xccache

import (
	"log"
	"sync"
)

type CacheGroup struct {
	name   string
	getter Getter //回调函数
	cache  cache
}

var mu sync.RWMutex
var groups = make(map[string]*CacheGroup)

func NewCacheGroup(name string, cacheBytes int64, getter Getter) *CacheGroup {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	cg := &CacheGroup{
		name:   name,
		getter: getter,
		cache:  cache{cacheBytes: cacheBytes},
	}
	groups[name] = cg
	return cg
}

func GetCacheGroup(name string) *CacheGroup {
	mu.RLock()
	defer mu.RUnlock()
	cg := groups[name]
	return cg
}

func (cg *CacheGroup) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, nil
	}
	if v, ok := cg.cache.get(key); ok {
		log.Println("[CacheGroup] hit")
		return v, nil
	}
	log.Println("[CacheGroup] miss, Now run getter")
	return cg.load(key)
}

func (cg *CacheGroup) load(key string) (ByteView, error) {
	bytes, err := cg.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	cg.cache.add(key, value)
	return value, nil
}
