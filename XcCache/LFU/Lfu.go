package lfu

import "container/list"

type Cache struct {
	maxBytes int64

	nBytes int64

	minRead int

	freqToList map[int]*list.List

	keyToElement map[string]*list.Element

	OnEvicted func(key string, value Value) //淘汰不常用缓存记录时的回调函数
}

type entry struct {
	key   string
	value Value
	freq  int
}

func (e *entry) GetKey() string {
	return e.key
}
func (e *entry) GetValue() Value {
	return e.value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:     maxBytes,
		freqToList:   make(map[int]*list.List),
		keyToElement: make(map[string]*list.Element),
		OnEvicted:    onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	e := c.getEntry(key)
	if e == nil {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) getEntry(key string) *entry {
	ele, ok := c.keyToElement[key]
	if !ok {
		return nil
	}
	e := ele.Value.(*entry)
	list := c.freqToList[e.freq]
	list.Remove(ele)
	if list.Len() == 0 {
		if e.freq == c.minRead {
			c.minRead++
		}
		delete(c.freqToList, e.freq)
	}
	e.freq += 1
	c.pushFront(e)
	return e
}
func (c *Cache) pushFront(e *entry) {
	_, ok := c.freqToList[e.freq]
	if !ok {
		c.freqToList[e.freq] = list.New()
	}
	c.keyToElement[e.key] = c.freqToList[e.freq].PushFront(e)
}
func (c *Cache) Put(key string, value Value) {
	if e := c.getEntry(key); e != nil {
		c.nBytes += int64(value.Len()) - int64(e.value.Len())
		e.value = value
		return
	}
	if c.nBytes+int64(value.Len())+int64(len(key)) > c.maxBytes {
		c.RemoveLowestFreq()
	}

	e := &entry{
		key:   key,
		value: value,
		freq:  1,
	}
	c.pushFront(e)
	c.nBytes += int64(len(key)) + int64(value.Len())
	c.minRead = 1
}

func (c *Cache) RemoveLowestFreq() {
	list := c.freqToList[c.minRead]
	delete(c.keyToElement, list.Remove(list.Back()).(*entry).key)
	if list.Len() == 0 {
		delete(c.freqToList, c.minRead)
	}
}
