# aim

Star IM CLoud

## 简介

**aim 是一个高性能分布式即时通信系统。**


- Web SDK： [Typescript SDK](https://github.com/starine/aim_web_sdk)

## 环境准备

### 中间件安装

aim依赖mysql、Consul和Redis。因此，在本地测试时需要准备相应环境。这里提供两种方式：

方式一： 通过docker-compose启动

> docker-compose -f "docker-compose.yml" up -d --build

方式二： docker分别启动

```cmd
docker run -itd --name aim_mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql

docker run \
    -d \
    -p 8500:8500 \
    -p 8600:8600/udp \
    --name=aim_consul \
    consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

docker run -itd --name aim_redis -p 6379:6379 redis
```

### 数据准备

1. 进入Mysql，修改访问权限：
   1. docker exec -it aim_mysql /bin/sh
   2. mysql -uroot -p123456
   3. GRANT ALL ON *.* TO 'root'@'%';
   4. flush privileges;
2. 创建数据库
   1. create database aim_base default character set utf8mb4 collate utf8mb4_unicode_ci;
   2. create database aim_message default character set utf8mb4 collate utf8mb4_unicode_ci;

## 启动服务

首先进入services中，分别启动三个服务：

```
go run main.go gateway
go run main.go server
go run main.go royal
```

访问Consul，可以查看服务启动状态：

> http://localhost:8500/ui

