package XcCache

import (
	"XcCache/XcCache/LRU"
	"encoding/json"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

type entry struct {
	key   string
	value []byte
}
type Entry struct {
	Key   string
	Value []byte
}

// 实现 json.Marshaler 接口
func (e Entry) MarshalJSON() ([]byte, error) {
	// 你可以根据需求来改变 value 字段的编码方式
	return json.Marshal(map[string]interface{}{
		"Key":   e.Key,
		"Value": string(e.Value), // 把字节切片转换为字符串
	})
}

// 实现 json.Unmarshaler 接口
func (e *Entry) UnmarshalJSON(data []byte) error {
	var aux map[string]interface{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if key, ok := aux["Key"].(string); ok {
		e.Key = key
	}
	if value, ok := aux["Value"].(string); ok {
		e.Value = []byte(value) // 把字符串转回字节切片
	}
	return nil
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.AddOrUpdate(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

func (c *cache) getAllList() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	list := c.lru.GetAllList()
	nodes := make([]Entry, 0)
	for _, node := range list {
		nodes = append(nodes, Entry{node.GetKey(), node.GetValue().(ByteView).b})
	}
	return nodes
}
