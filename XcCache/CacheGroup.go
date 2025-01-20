package XcCache

import (
	"errors"
	"log"
	"sync"
)

type CacheGroup struct {
	name   string
	getter Getter //回调函数
	cache  cache
	peers  PeerPicker
}

var mu sync.RWMutex
var groups = make(map[string]*CacheGroup)

func NewCacheGroup(name string, cacheBytes int64, getter Getter) *CacheGroup {
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

func (cg *CacheGroup) RegisterPeers(peers PeerPicker) {
	if cg.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	cg.peers = peers
}

func (cg *CacheGroup) load(key string) (ByteView, error) {
	if cg.peers != nil {
		if peer, ok := cg.peers.PickPeer(key); ok {
			if value, err := cg.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[CacheGroup] Failed to get from peer")
		}
	}
	return cg.loadLocally(key)
}

func (cg *CacheGroup) loadLocally(key string) (ByteView, error) {
	if cg.getter == nil {
		return ByteView{}, errors.New("no such key")
	}
	bytes, err := cg.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	cg.cache.add(key, value)
	return value, nil
}

func (cg *CacheGroup) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(cg.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
