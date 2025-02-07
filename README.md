# XcXcCache

## 介绍
分布式缓存

## Prerequisites

- **Golang** 1.23.3 or later
- **Etcd** v3.4.0 or later
- **gRPC-go** v1.38.0 or later
- **protobuf** v1.26.0 or later

## Installation

1.  git clone https://gitee.com/Xccccee/xc-xc-cache.git

2.  go mod tidy

3.  go build -o server .

4. ./server

## 使用说明

1.Configure the etcd service 配置etcd

2.RUN ./server 启动server

3.输入端口号
![输入图片说明](https://foruda.gitee.com/images/1738933621654175220/a3b364d5_14206221.png "屏幕截图 2025-02-07 210518.png")
![输入图片说明](https://foruda.gitee.com/images/1738933646909013202/90241a8c_14206221.png "屏幕截图 2025-02-07 210540.png")

4.添加节点
![输入图片说明](https://foruda.gitee.com/images/1738933697259137632/f50547ba_14206221.png "屏幕截图 2025-02-07 210758.png")

5.节点发现
![输入图片说明](https://foruda.gitee.com/images/1738933753851936037/92253edb_14206221.png "屏幕截图 2025-02-07 210840.png")

6.删除节点
![输入图片说明](https://foruda.gitee.com/images/1738933769379488465/b45ab53e_14206221.png "屏幕截图 2025-02-07 210852.png")
