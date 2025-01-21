package main

import (
	"XcCache/XcCache"
	"XcCache/XcCache/etcd"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var addr string
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Before starting this service, make sure that you have configured ETCD")
	fmt.Println("Please input server port:")
	fmt.Scanln(&addr)
	addr = "localhost:" + addr
	go func(addr string) {
		cacheGroup := XcCache.NewCacheGroup("test", 2<<10, nil)
		server := XcCache.NewServer(addr)
		cli, err := etcd.ClientInit()
		defer cli.Close()
		if err != nil {
			log.Fatalf("etcd client init failed: %v", err)
		}
		peers, err := etcd.GetAllPeers(cli, "xccache")
		if err != nil {
			log.Fatalf("get other peers from etcd failed: %v", err)
		}
		fmt.Printf("other peers: %v\n", peers)
		peers = append(peers, addr)
		server.SetPeers(peers...)
		cacheGroup.RegisterPeers(server)

		var allPeers []string
		allPeers = append(allPeers, peers...)
		go func() {
			time.Sleep(5 * time.Second)
			for {
				currentPeers, _ := etcd.GetAllPeers(cli, "xccache")
				closedPeers := GetClosedPeers(allPeers, currentPeers)
				newPeers := GetNewPeers(allPeers, currentPeers)
				if len(closedPeers) != 0 {
					server.DelPeers(closedPeers...)
					cacheGroup.RegisterPeers(server)
					log.Println("Nodes have been closed :", closedPeers)
					log.Println("Current nodes :", currentPeers)
				}
				if len(newPeers) != 0 {
					server.AddPeers(newPeers...)
					cacheGroup.RegisterPeers(server)
					log.Println("New nodes have been added :", newPeers)
					log.Println("Current nodes :", currentPeers)
				}
				allPeers = currentPeers
				time.Sleep(5 * time.Second)

			}
		}()

		if err := server.StartServer(); err != nil {
			log.Fatalf("start server error: %v", err)
		}
	}(addr)
	<-exitCh
	log.Println("Shutdown signal received, exiting...")
}

func GetClosedPeers(peers, newpeers []string) []string {
	var closedPeers []string
	for _, peer := range peers {
		if !IsContain(newpeers, peer) {
			closedPeers = append(closedPeers, peer)
		}
	}
	return closedPeers
}
func GetNewPeers(peers, newpeers []string) []string {
	var newPeers []string
	for _, peer := range newpeers {
		if !IsContain(peers, peer) {
			newPeers = append(newPeers, peer)
		}
	}
	return newPeers
}

func IsContain(peers []string, peer string) bool {
	for _, p := range peers {
		if p == peer {
			return true
		}
	}
	return false
}
