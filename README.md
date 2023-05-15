# etcd

[etcd](https://github.com/etcd-io/etcd) is a distributed reliable key-value store for the most critical data of a distributed system, with a focus on being:

- Simple: well-defined, user-facing API (gRPC)
- Secure: automatic TLS with optional client cert authentication
- Fast: benchmarked 10,000 writes/sec
- Reliable: properly distributed using Raft

## 安装

地址： https://github.com/etcd-io/etcd/releases/

启动etcd：

```shell
raft2023/05/15 20:13:53 INFO: 8e9e05c52164694d is starting a new election at term 1
raft2023/05/15 20:13:53 INFO: 8e9e05c52164694d became candidate at term 2
raft2023/05/15 20:13:53 INFO: 8e9e05c52164694d received MsgVoteResp from 8e9e05c52164694d at term 2
raft2023/05/15 20:13:53 INFO: 8e9e05c52164694d became leader at term 2
raft2023/05/15 20:13:53 INFO: raft.node: 8e9e05c52164694d elected leader 8e9e05c52164694d at term 2
2023-05-15 20:13:53.916380 I | etcdserver: setting up the initial cluster version to 3.4
2023-05-15 20:13:53.916380 I | etcdserver: published {Name:default ClientURLs:[http://localhost:2379]} to cluster cdf818194e3a8c32
2023-05-15 20:13:53.916380 I | embed: ready to serve client requests
2023-05-15 20:13:53.916380 N | etcdserver/membership: set the initial cluster version to 3.4
2023-05-15 20:13:53.917379 N | embed: serving insecure client requests on 127.0.0.1:2379, this is strongly discouraged!
2023-05-15 20:13:53.917379 I | etcdserver/api: enabled capabilities for version 3.4
```

etcdctl版本：

```shell
$ etcdctl version
etcdctl version: 3.4.26
API version: 3.4
```

## API

### 写入key

```shell
$ etcdctl put mark "hello mark"
OK
```

### 读取key

```shell
$ etcdctl get mark
mark
hello mark
```

### 删除key

```shell
$ etcdctl del mark
1

$ etcdctl get mark

```

### 监听key

```shell
$ etcdctl watch mark
PUT
mark      
hello mark
PUT
mark             
hello hello mark 
PUT
mark
hello hello hello mark
```

另一个终端中执行：

```shell
mark@LAPTOP-VH57ARI1 MINGW64 ~
$ etcdctl put mark "hello mark"
OK

mark@LAPTOP-VH57ARI1 MINGW64 ~
$ etcdctl put mark "hello hello mark "
OK

mark@LAPTOP-VH57ARI1 MINGW64 ~
$ etcdctl put mark "hello hello hello mark "
OK
```

### 设置租约（Grant leases）

当一个key被绑定到一个租约上时，它的生命周期与租约的生命周期绑定。

值得注意的地方，一个租约可以绑定多个key，当租约过期后，所有key值会被删除。

当一个租约只绑定了一个key时，想删除这个key，最好的办法是撤销它的租约，而不是直接删除这个key。

创建一个租约：
```shell
$ etcdctl lease grant -h

$ etcdctl lease grant 60s
Error: bad TTL (strconv.ParseInt: parsing "60s": invalid syntax)

# 设置一个60s的租约
$ etcdctl lease grant 60
lease 694d881f54d5810c granted with TTL(60s)

# 将租约与 mark这个key绑定
$ etcdctl put --lease=694d881f54d5810c mark mark
OK

$ etcdctl get mark
mark
mark

$ etcdctl get mark
mark
mark


# 60s后，获取不到mark了
$ etcdctl get mark

$ etcdctl get mark
```

### 主动撤销租约（Revoke leases）
撤销租约将删除其所有绑定的key

```shell
$ etcdctl lease grant 60
lease 694d881f54d5811f granted with TTL(60s)

$ etcdctl put mark mark --lease 694d881f54d5811f
OK

$ etcdctl lease revoke 694d881f54d5811f
lease 694d881f54d5811f revoked

$ etcdctl get mark

```

### 续租约（Keep leases alive）

通过刷新其TTL来保持租约的有效，使其不会过期。

```shell
$ etcdctl lease grant 60
lease 694d881f54d58124 granted with TTL(60s)

$ etcdctl put mark mark --lease 694d881f54d58124
OK

#续租约，自动定时执行续租约，续约成功后每次租约为60秒
$ etcdctl lease keep-alive 694d881f54d58124
lease 694d881f54d58124 keepalived with TTL(60)
lease 694d881f54d58124 keepalived with TTL(60)
lease 694d881f54d58124 keepalived with TTL(60)
```

### 获取租约信息（Get lease information）

获取租约信息，以便续租或查看租约是否仍然存在或已过期

```shell
# --keys 指定为 leaseId
$ etcdctl lease timetolive --keys 694d881f54d58124
lease 694d881f54d58124 granted with TTL(60s), remaining(58s), attached keys([mark])

$ etcdctl lease timetolive --keys 694d881f54d58124
lease 694d881f54d58124 granted with TTL(60s), remaining(55s), attached keys([mark])

$ etcdctl lease timetolive --keys 694d881f54d58124
lease 694d881f54d58124 granted with TTL(60s), remaining(53s), attached keys([mark])

$ etcdctl lease timetolive --keys 694d881f54d58124
lease 694d881f54d58124 granted with TTL(60s), remaining(51s), attached keys([mark])

$ etcdctl lease timetolive --keys 694d881f54d58124
lease 694d881f54d58124 granted with TTL(60s), remaining(50s), attached keys([mark])
```

## 应用场景

- 服务发现
- 配置中心
- 负载均衡
- 分布式锁



