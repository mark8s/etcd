package main

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

// ServiceRegister 创建租约注册服务
type ServiceRegister struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
}

// NewServiceRegister 新建注册服务
func NewServiceRegister(endpoints []string, key, val string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	ser := &ServiceRegister{
		cli: cli,
		key: key,
		val: val,
	}

	//申请租约设置时间keepalive
	if err := ser.putKeyWithLease(lease); err != nil {
		return nil, err
	}

	return ser, nil
}

func (s *ServiceRegister) putKeyWithLease(lease int64) error {

	// 创建租约
	grant, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}

	// 绑定租约到服务
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}

	// 续租 每次续3s、4s
	alive, err := s.cli.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return err
	}

	s.leaseID = grant.ID
	s.keepAliveChan = alive
	log.Printf("Put key:%s  val:%s  success!", s.key, s.val)

	return nil
}

func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		log.Println("续约成功", leaseKeepResp)
	}
	log.Println("关闭续租")
}

// Close 注销服务
func (s *ServiceRegister) Close() error {
	// 撤销租约
	_, err := s.cli.Revoke(context.Background(), s.leaseID)
	if err != nil {
		return err
	}
	log.Println("撤销租约")
	return s.cli.Close() // 关闭 client连接
}

func main() {
	// etcd 端点
	var endpoints = []string{"localhost:2379"}

	register, err := NewServiceRegister(endpoints, "/web/node1", "localhost:8080", 10)
	if err != nil {
		return
	}

	//监听续租相应chan
	go register.ListenLeaseRespChan()

	// 三分钟后自动关闭
	select {
	case <-time.After(180 * time.Second):
		register.Close()
	}
}
