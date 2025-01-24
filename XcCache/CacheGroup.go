package XcCache

import (
	"XcCache/XcCache/singleflight"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const snapShotPrefix = "snapShot_"

type CacheGroup struct {
	name   string
	getter Getter //回调函数
	cache  cache
	peers  PeerPicker
	flight *singleflight.FlightGroup
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
		flight: &singleflight.FlightGroup{},
	}

	go func() {
		cg.loadSnapShot()
		cg.StartSaveSnapShot(time.Minute * 2)
	}()

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

func (cg *CacheGroup) Set(key string, value []byte) {
	cg.cache.add(key, ByteView{b: value})

}

func (cg *CacheGroup) RegisterPeers(peers PeerPicker) {
	//if cg.peers != nil {
	//	panic("RegisterPeerPicker called more than once")
	//}
	cg.peers = peers
}

func (cg *CacheGroup) load(key string) (ByteView, error) {
	view, err := cg.flight.Do(key, func() (interface{}, error) {
		if cg.peers != nil {
			if peer, ok := cg.peers.PickPeer(key); ok {
				if value, err := cg.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[CacheGroup] Failed to get from peer")
			}
		}
		return cg.loadLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return view.(ByteView), nil
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

// DB快照(参考Rdb快照redis database)
func (cg *CacheGroup) StartSaveSnapShot(duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func(ticker *time.Ticker) {
		defer ticker.Stop()
		filepath := snapShotPrefix + cg.name + ".json"
		for range ticker.C {
			cacheList := cg.cache.getAllList()
			fmt.Println("cacheList:", cacheList)

			// 打开文件并使用os.O_TRUNC标志位来截断文件内容
			localJson, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Println("[CacheGroup] SaveSnapShot error:", err)
				continue
			}
			defer localJson.Close()

			encoder := json.NewEncoder(localJson)
			if err := encoder.Encode(cacheList); err != nil {
				log.Println("[CacheGroup] SaveSnapShot error:", err)
				continue
			}

			log.Println("[CacheGroup] SaveSnapShot success")
		}
	}(ticker)
}

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
}

func (cg *CacheGroup) loadSnapShot() {
	bytes, err := os.ReadFile(snapShotPrefix + cg.name + ".json")
	if err != nil {
		log.Println("[CacheGroup] loadSnapShot error:", err)
		return
	}
	var cacheList []Entry
	if err := json.Unmarshal(bytes, &cacheList); err != nil {
		log.Println("[CacheGroup] loadSnapShot error:", err)
		return
	}

	for _, entry := range cacheList {
		cg.cache.add(entry.Key, ByteView{b: entry.Value})
	}
	log.Println("[CacheGroup] loadSnapShot success")
}
