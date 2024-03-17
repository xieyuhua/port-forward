package main

import (
	"flag"
	"sync"

	"goForward/conf"
	"goForward/forward"
	"goForward/sql"
	"goForward/web"
)

func main() {
	go web.Run()
	// 初始化通道
	conf.Ch = make(chan string)
	forwardList := sql.GetAction()
	if len(forwardList) == 0 {
		//添加测试数据
		testData := conf.ConnectionStats{
			LocalPort:  conf.WebPort,
			RemotePort: conf.WebPort,
			RemoteAddr: "127.0.0.1",
			OutTime:5,
			Protocol:   "udp",
		}
		sql.AddForward(testData)
		forwardList = sql.GetForwardList()
	}
	var largeStats forward.LargeConnectionStats
	largeStats.Connections = make([]*forward.ConnectionStats, len(forwardList))
	for i := range forwardList {
		connectionStats := &forward.ConnectionStats{
			ConnectionStats: conf.ConnectionStats{
				Id:         forwardList[i].Id,
				Protocol:   forwardList[i].Protocol,
				LocalPort:  forwardList[i].LocalPort,
				RemotePort: forwardList[i].RemotePort,
				RemoteAddr: forwardList[i].RemoteAddr,
				OutTime:    forwardList[i].OutTime,
				TotalBytes: forwardList[i].TotalBytes,
			},
			TotalBytesOld:  forwardList[i].TotalBytes,
			TotalBytesLock: sync.Mutex{},
			TCPConnections: make(map[string]*forward.IPStruct), 
		}

		largeStats.Connections[i] = connectionStats
	}
	// 设置 WaitGroup 计数为连接数
	conf.Wg.Add(len(largeStats.Connections))

	// 并发执行多个转发
	for _, stats := range largeStats.Connections {
		go forward.Run(stats, &conf.Wg)
	}
	conf.Wg.Wait()
	defer close(conf.Ch)
}
func init() {
	flag.StringVar(&conf.WebPort, "port", "8889", "Web Port")
	flag.StringVar(&conf.WebPass, "pass", "", "Web Password")
	flag.Parse()
}
