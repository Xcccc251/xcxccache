package main

import (
	"XcCache/XcCache"
	"log"
)

func main() {
	cacheGroup := XcCache.NewCacheGroup("scores", 2<<10, nil)
	addr := "localhost:9999"
	server := XcCache.NewServer(addr)
	server.SetPeers(addr)
	cacheGroup.RegisterPeers(server)
	err := server.StartServer()
	if err != nil {
		log.Fatalf("start server error: %v", err)
	}
}
