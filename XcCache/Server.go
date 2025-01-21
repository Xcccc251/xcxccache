package XcCache

import (
	"XcCache/XcCache/consistentHash"
	"XcCache/XcCache/etcd"
	"XcCache/XcCache/xccache"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

const defaultBasePath = "/_xccache/"
const defaultReplicas = 50

type Server struct {
	xccache.UnimplementedXcCacheServer

	addr        string
	status      bool
	basePath    string
	stopCh      chan struct{}
	mu          sync.Mutex
	peers       *consistentHash.Map
	httpGetters map[string]*cacheClient
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", s.addr, fmt.Sprintf(format, v...))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, s.basePath) {
		panic("Server serving unexpected path: " + r.URL.Path)
	}

	s.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(s.basePath):], "/", 2)
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

func (s *Server) SetPeers(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers = consistentHash.New(defaultReplicas, nil)
	s.peers.Add(peers...)
	s.httpGetters = make(map[string]*cacheClient, len(peers))
	for _, peer := range peers {
		s.httpGetters[peer] = &cacheClient{serviceName: "xccache/" + peer}
	}
}

func (s *Server) PickPeer(key string) (PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if peer := s.peers.Get(key); peer != "" && peer != s.addr {
		s.Log("Pick peer %s", peer)
		return s.httpGetters[peer], true
	}
	return nil, false

}

func (s *Server) Get(ctx context.Context, in *xccache.GetRequest) (*xccache.GetResponse, error) {
	group, key := in.Group, in.Key
	log.Printf("[xccache %s] Recv RPC Request->Get - Group: %s,Key: %s", s.addr, group, key)
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}

	g := GetCacheGroup(group)
	if g == nil {
		return nil, fmt.Errorf("group %s not found", group)
	}

	view, err := g.Get(key)
	if err != nil {
		return nil, err
	}
	return &xccache.GetResponse{Value: view.ByteSlice()}, nil
}

func (s *Server) Set(ctx context.Context, in *xccache.SetRequest) (*xccache.SetResponse, error) {
	group, key, value := in.Group, in.Key, in.Value
	log.Printf("[xccache %s] Recv RPC Request->Set - Group: %s,Key: %s,Value: %s", s.addr, group, key, string(in.Value))
	if key == "" {
		return &xccache.SetResponse{Success: false}, fmt.Errorf("key is empty")
	}
	peer := s.peers.Get(key)
	if peer == s.addr {
		g := GetCacheGroup(group)
		if g == nil {
			log.Printf("[xccache %s] group %s not found", s.addr, group)
			NewCacheGroup(group, 0, nil)
			g = GetCacheGroup(group)
		}
		g.Set(key, value)
		return &xccache.SetResponse{Success: true}, nil
	}

	//isSuccess, err := s.httpGetters[peer].Set(group, key, value)
	//if err != nil {
	//	return &xccache.SetResponse{Success: false}, err
	//}
	//
	//return &xccache.SetResponse{Success: isSuccess}, nil
	return &xccache.SetResponse{Success: true}, nil
}

func (s *Server) StartServer() error {
	s.mu.Lock()
	if s.status == true {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}

	s.status = true
	s.stopCh = make(chan struct{})

	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port) // 监听端口
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	xccache.RegisterXcCacheServer(grpcServer, s)
	//注册至etcd
	go func() {
		log.Printf("server %s", s.addr)
		err := etcd.RegisterToEtcd("xccache", s.addr, s.stopCh)
		if err != nil {
			log.Fatalf("register to etcd failed: %v", err)
		}
		close(s.stopCh)
		err = lis.Close()
		if err != nil {
			log.Fatalf("close listener failed: %v", err)
		}
		log.Printf("[%s] Revoke service and close tcp socket ok.", s.addr)

	}()
	s.mu.Unlock()
	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	return nil
}

func (s *Server) StopServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == false {
		return fmt.Errorf("server already stopped")
	}
	s.status = false
	s.stopCh <- struct{}{}
	close(s.stopCh)
	return nil
}
