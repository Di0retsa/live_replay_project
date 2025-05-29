# 仿小鹅通项目本地部署指引

> 整体项目架构：
>
> 前端：React + Vite（node:22.16.0）
>
> 后端：Gin（Go:1.24.1） + Redis-stack（7.4.0） + MySQL（8.3.0）
>
> 另：后端解析视频时长需要用到FFmpeg

### 前端：

```shell
# 移动至前端目录
$ cd /live_replay_project/frontend/frontend
# 下载项目依赖
$ npm install
# 运行
$ npm run dev
```

### 后端：

```shell
# 移动至后端目录
$ cd /live_replay_project/backend
# 下载项目依赖
$ go mod tidy
# 运行
$ go run main.go
```

### MySQL：

```sql
-- 创建名为live_replay_db的数据库
CREATE DATABASE live_replay_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

create table chat_message
(
    message_id int auto_increment
        primary key,
    user_id    int          not null,
    username   varchar(16)  not null,
    content    varchar(256) not null,
    timestamp  datetime     not null,
    replay_id  int          not null,
    type       varchar(20)  not null
)
    engine = InnoDB;

create table replay
(
    replay_id    int auto_increment
        primary key,
    title        varchar(255) not null,
    description  text         not null,
    duration     bigint       not null,
    storage_path varchar(512) not null,
    user_id      int          null,
    create_time  datetime     null,
    update_time  datetime     null,
    cover_path   varchar(512) not null,
    views        int          not null,
    comments     int          not null
)
    engine = InnoDB;

create table user
(
    user_id     int auto_increment
        primary key,
    username    varchar(16) not null,
    password    varchar(64) not null,
    create_time datetime    null,
    update_time datetime    null,
    phone       char(11)    not null,
    constraint phone
        unique (phone),
    constraint username
        unique (username)
)
    engine = InnoDB;
-- 创建replay_user，密码为password
CREATE USER 'replay_user'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON live_replay_db.* TO 'replay_user'@'localhost';
FLUSH PRIVILEGES;
```

### Redis-Stack(推荐使用docker部署):

```shell
# 除了这个没有什么特别要设置的
requirepass your_password
# 需要用到Redis-Stack集成的RedisJson2插件，请确保您的Redis可以执行JSON.SET等操作
```

>因为没有使用Nginx所以直接访问localhost:5173即可
>
>另：项目已部署至云服务器，欢迎访问http://113.45.49.106:5173（但是因为服务器网络带宽只有2Mb/s所以实际效果很差，还是推荐本地部署XD）
>
>最后：项目主要应用于移动端所以桌面端的UI界面可能存在显示问题XD
