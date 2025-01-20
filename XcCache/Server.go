package XcCache

import (
	"XcCache/XcCache/consistentHash"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

const defaultBasePath = "/_xccache/"
const defaultReplicas = 50

type Server struct {
	addr        string
	basePath    string
	stopCh      chan error
	mu          sync.Mutex
	peers       *consistentHash.Map
	httpGetters map[string]*cacheClient
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (p *Server) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.addr, fmt.Sprintf(format, v...))
}

func (p *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("Server serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetCacheGroup(groupName)
	if group == nil {
		http.Error(w, "no such group :"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (p *Server) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistentHash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*cacheClient, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &cacheClient{serviceName: "xccache/" + peer}
	}
}

func (p *Server) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.addr {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false

}
