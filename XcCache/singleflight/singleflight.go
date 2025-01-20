package singleflight

import "sync"

//防止缓存击穿

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}
type group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if p, ok := g.m[key]; ok {
		g.mu.Unlock()
		p.wg.Wait()
		return p.val, p.err
	}
	p := &call{}
	p.wg.Add(1)
	g.m[key] = p
	g.mu.Unlock()

	p.val, p.err = fn()
	p.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return p.val, p.err
}
