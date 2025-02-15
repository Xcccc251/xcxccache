package XcCache

import (
	"XcCache/XcCache/etcd"
	"XcCache/XcCache/xccache"
	"context"
	"fmt"
	"time"
)

// 实现PickGetter接口
type cacheClient struct {
	serviceName string //服务名 xccache/x.x.x.x:port
}

func (h *cacheClient) Get(group string, key string) ([]byte, error) {
	//创建一个etcd client
	cli, err := etcd.ClientInit()
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	conn, err := etcd.DiscoverFromEtcd(cli, h.serviceName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	grpcClient := xccache.NewXcCacheClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	rsp, err := grpcClient.Get(ctx, &xccache.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, h.serviceName)
	}
	return rsp.Value, nil
}

func (h *cacheClient) Set(group string, key string, value []byte) (bool, error) {
	//创建一个etcd client
	cli, err := etcd.ClientInit()
	if err != nil {
		return false, err
	}
	defer cli.Close()
	conn, err := etcd.DiscoverFromEtcd(cli, h.serviceName)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	grpcClient := xccache.NewXcCacheClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	rsp, err := grpcClient.Set(ctx, &xccache.SetRequest{
		Group: group,
		Key:   key,
		Value: value,
	})
	if err != nil {
		return false, fmt.Errorf("could not set %s/%s to peer %s", group, key, h.serviceName)
	}
	return rsp.Success, nil
}
