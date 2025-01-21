package XcCache

import (
	"XcCache/XcCache/etcd"
	"XcCache/XcCache/xccache"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	cacheGroup := NewCacheGroup("scores", 2<<10, nil)
	addr := "localhost:9999"
	server := NewServer(addr)
	server.SetPeers(addr)
	cacheGroup.RegisterPeers(server)
	server.StartServer()

}

func TestGRPCSet(t *testing.T) {
	cli, err := etcd.ClientInit()
	if err != nil {
		fmt.Println(err)
	}
	defer cli.Close()
	conn, err := etcd.DiscoverFromEtcd(cli, "xccache/localhost:9999")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	grpcClient := xccache.NewXcCacheClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	rsp, err := grpcClient.Set(ctx, &xccache.SetRequest{
		Group: "test",
		Key:   "test",
		Value: []byte("Ok?"),
	})
	fmt.Println(rsp.Success)
}

func TestGRPCGet(t *testing.T) {
	cli, err := etcd.ClientInit()
	if err != nil {
		fmt.Println(err)
	}
	defer cli.Close()
	conn, err := etcd.DiscoverFromEtcd(cli, "xccache/localhost:9999")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	grpcClient := xccache.NewXcCacheClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	rsp, err := grpcClient.Get(ctx, &xccache.GetRequest{
		Group: "test",
		Key:   "test",
	})
	fmt.Println(string(rsp.Value))
}
