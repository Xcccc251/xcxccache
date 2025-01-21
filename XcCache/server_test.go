package XcCache

import (
	"testing"
)

func TestGet(t *testing.T) {
	cacheGroup := NewCacheGroup("scores", 2<<10, nil)
	addr := "localhost:9999"
	server := NewServer(addr)
	server.SetPeers(addr)
	cacheGroup.RegisterPeers(server)
	server.StartServer()

}
