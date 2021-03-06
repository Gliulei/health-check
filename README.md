# health-check

#### 介绍
**health-check**
1. 基于raft协议的健康探活工具，保证高可用
2. 多点决策、避免单点视角
3. 支持http和tcp方式两种方式探活
4. 方便部署、运维方便，对外提供http接口
5. 探活结果回调告知

#### 软件架构

![架构图](./docs/assets/health-check.png)

#### 安装教程

+ make编译
+ 最少三台机器部署

#### 使用说明

1.  [health-check的HTTP接口文档](./docs/api.md)
2.  [health-check配置说明](./docs/configuration.md)



#### Test
+ go run main.go -config etc/health-check.ini
+ go run main.go -config etc/health-check2.ini
+ go run main.go -config etc/health-check3.ini

#### 联系Author
email:liulei4567@qq.com