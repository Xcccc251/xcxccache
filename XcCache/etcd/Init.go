package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var DefaultClient = ClientInit()
var defaultEtcdConfig = clientv3.Config{
	Endpoints:   []string{"http://127.0.0.1:2379"},
	DialTimeout: time.Second * 10,
}

func ClientInit() *clientv3.Client {
	client, _ := clientv3.New(defaultEtcdConfig)
	return client
}
