package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
)

func AddToEtcd(c *clientv3.Client, leaseId clientv3.LeaseID, service string, addr string) error {
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseId))
}

// 向etcd注册一个服务
func RegisterToEtcd(service string, addr string, iscancel chan error) error {
	cli := DefaultClient
	defer cli.Close()
	//ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	resp, err := cli.Grant(context.Background(), 10)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}
	leaseId := resp.ID //租约id

	err = AddToEtcd(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("add to etcd failed: %v", err)
	}
	//心跳检测

	aliveCh, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	log.Printf("[%s] register service ok\n", addr)

	for {
		select {
		case _, ok := <-aliveCh:
			if !ok {
				log.Println("keep alive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
			log.Printf("[%s] keepalive ok\n", addr)
		case err := <-iscancel:
			log.Printf("[%s] service closed\n", addr)
			return err
		case <-cli.Ctx().Done():
			log.Printf("[%s] service closed\n", addr)
			return nil
		}
	}
}
