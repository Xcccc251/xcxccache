package etcd

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

// 服务发现,获取指定地址
func DiscoverFromEtcd(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdBuilder, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(
		service,
		grpc.WithResolvers(etcdBuilder),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
}

func GetAllPeers(c *clientv3.Client, prefix string) ([]string, error) {
	resp, err := c.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var addresses []string

	for _, kv := range resp.Kvs {
		result := kv.Value
		var endPoint endpoints.Endpoint
		if err := json.Unmarshal(result, &endPoint); err != nil {
			return nil, err
		}
		addresses = append(addresses, endPoint.Addr)
	}
	return addresses, nil

}
