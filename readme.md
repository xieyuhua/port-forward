使用golang实现的tcp udp端口转发

Fork https://github.com/csznet/goForward

目前已实现：

 - 规则热加载
 - web管理面板
 - 流量统计
 - 空闲时长断开连接设置

**截图**

![image](https://github.com/xieyuhua/portforward/assets/29120060/7a9355d7-b63e-495b-808d-5c262441e159)

```
2024-03-18 08:17:48 【tcp】端口 1521 当前连接数: 2, 统计流量: 653.72MB
2024-03-18 08:17:48 【1521】端口 1、  192.167.1.6:49095->192.168.9.38:1521
2024-03-18 08:17:48 【1521】端口 2、  192.168.9.38:1521->192.167.1.6:49095
```

**使用**

运行
```
./goForward -h

  -debug
    	Print connection
  -pass string
    	Web Password
  -port string
    	Web Port (default "8889")

```

**参数**

设置web管理访问密码

```
./goForward  -port 8899 -pass 666
```

当24H内同一IP密码试错超过3次将会ban掉

## 开机自启

**创建 Systemd 服务**

```
sudo nano /etc/systemd/system/goForward.service
```

**输入内容**

```
[Unit]
Description=Start goForward on boot

[Service]
ExecStart=/full/path/to/your/goForward

[Install]
WantedBy=default.target
```

其中的```/full/path/to/your/goForward```改为二进制文件地址，后面可接参数

**重新加载 Systemd 配置**
```
sudo systemctl daemon-reload
```

**启用服务**
```
sudo systemctl enable goForward
```
**启动服务**
```
sudo systemctl start goForward
```
**检查状态**
```
sudo systemctl status goForward.service
```
