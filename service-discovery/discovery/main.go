package main

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

type ServiceDiscovery struct {
	cli        *clientv3.Client
	serverList map[string]string // 服务列表
	mu         sync.Mutex
}

func NewServiceDiscovery(endpoints []string) *ServiceDiscovery {
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return &ServiceDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

func (s *ServiceDiscovery) WatchService(prefix string) error {
	//根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		s.SetServiceList(string(kv.Key), string(kv.Value))
	}

	//监视前缀，修改变更的server
	go s.watcher(prefix)
	return nil
}

// watcher 监听前缀
func (s *ServiceDiscovery) watcher(prefix string) {
	watchChan := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)

	for wresp := range watchChan {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// DelServiceList 删除服务地址
func (s *ServiceDiscovery) DelServiceList(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serverList, key)
	log.Println("del key:", key)
}

func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serverList[key] = string(val)
	log.Println("put key :", key, "val:", val)
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetServices() map[string][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	addrs := make([]string, 0)
	serviceList := make(map[string][]string)

	for key, v := range s.serverList {
		addrs = append(addrs, v)
		serviceList[key] = addrs
	}
	return serviceList
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}

func main() {
	var endpoints = []string{"localhost:2379"}
	ser := NewServiceDiscovery(endpoints)
	defer ser.Close()
	ser.WatchService("/web/")
	ser.WatchService("/gRPC/")
	for {
		select {
		case <-time.Tick(5 * time.Second):
			for svc, ep := range ser.GetServices() {
				log.Printf("service: %s ,endpoints: %s\n", svc, ep)
			}
		}
	}

}
