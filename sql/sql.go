package sql

import (
	"os/exec"  
	"fmt"
	"bufio"  
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
    "strconv"
	"goForward/conf"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 定义数据库指针
var db *gorm.DB

func init() {
	var err error
	var dbPath string
	executablePath, err := os.Executable()
	if err != nil {
		log.Println("获取可执行文件路径失败:", err)
		log.Println("使用默认获取的路径")
		dbPath = "goForward.db"
	} else {
		dbPath = filepath.Join(filepath.Dir(executablePath), "goForward.db")
	}
	fmt.Println("Data:", dbPath)
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Println("连接数据库失败")
		return
	}
	db.AutoMigrate(&conf.ConnectionStats{})
	db.AutoMigrate(&conf.IpBan{})
}

// 获取转发列表
func GetForwardList() []conf.ConnectionStats {
	var res []conf.ConnectionStats
	db.Model(&conf.ConnectionStats{}).Find(&res)
    var size string  
    var bytes float64  
    for i := range res {  
        bytes = float64(res[i].TotalBytes)  
        if bytes > 0 {  
            if bytes < (1024 * 1024 * 0.5) {  
                size = strconv.FormatFloat(bytes/1024, 'f', 2, 64) + " KB"  
            } else {  
                size = strconv.FormatFloat(bytes/(1024*1024), 'f', 2, 64) + " MB"  
            }  
            res[i].TolBytes = size // 假设您有一个TolBytes字段来存储格式化后的值  
        } else {
            res[i].TolBytes = " 0KB" 
        }
    }
	return res
}

// 获取启用的转发列表
func GetAction() []conf.ConnectionStats {
	var res []conf.ConnectionStats
	db.Model(&conf.ConnectionStats{}).Where("status = ?", 0).Find(&res)
	return res
}

// 获取ipban列表
func GetIpBan() []conf.IpBan {
	var res []conf.IpBan
	db.Model(&conf.IpBan{}).Find(&res)
	return res
}

// 修改指定转发统计流量(byte)
func UpdateForwardBytes(id int, bytes uint64) bool {
	res := db.Model(&conf.ConnectionStats{}).Where("id = ?", id).Update("total_bytes", bytes)
	if res.Error != nil {
		fmt.Println(res.Error)
		return false
	}
	return true
}

// 修改指定转发统计流量(byte)
func UpdateForwardGb(id int, gb uint64) bool {
	res := db.Model(&conf.ConnectionStats{}).Where("id = ?", id).Update("total_gigabyte", gb)
	if res.Error != nil {
		fmt.Println(res.Error)
		return false
	}
	return true
}

// 修改指定转发状态
func UpdateForwardStatus(id int, status int) bool {
	res := db.Model(&conf.ConnectionStats{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		fmt.Println(res.Error)
		return false
	}
	return true
}

// 获取指定转发内容
func GetForward(id int) conf.ConnectionStats {
	var get conf.ConnectionStats
	db.Model(&conf.ConnectionStats{}).Where("id = ?", id).Find(&get)
	return get
}

// checkPortWithNetstat 使用netstat命令检查端口是否启用  
func checkPortWithNetstat(port string) bool {  
	cmd := exec.Command("netstat", "-tuln")  
	output, err := cmd.Output()  
	if err != nil {  
		return false  
	}  
	scanner := bufio.NewScanner(strings.NewReader(string(output)))  
	for scanner.Scan() {  
		line := scanner.Text()  
		// 查找包含指定端口号的行  
		if strings.Contains(line, ":"+port+" ") {  
			return true  
		}  
	}  
	return false  
}


// 判断指定端口转发是否可添加
func FreeForward(localPort, protocol string) bool {
    
    //  return false
    if checkPortWithNetstat(localPort) {
        return false
    }
    
	var get conf.ConnectionStats
	res := db.Model(&conf.ConnectionStats{}).Where("local_port = ? And protocol = ?", localPort, protocol).Find(&get)
	if res.Error == nil {
		if get.Id == 0 {
			return true
		} else {
			return false
		}
	}
	return false
}

// 去掉所有空格
func rmSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "")
}

// 增加转发
func AddForward(newForward conf.ConnectionStats) int {
	//预处理
	newForward.RemoteAddr = rmSpaces(newForward.RemoteAddr)
	newForward.RemotePort = rmSpaces(newForward.RemotePort)
	newForward.LocalPort = rmSpaces(newForward.LocalPort)
	newForward.Blacklist = rmSpaces(newForward.Blacklist)
	newForward.Whitelist = rmSpaces(newForward.Whitelist)
	newForward.Protocol = rmSpaces(newForward.Protocol)
	if newForward.Protocol != "udp" {
		newForward.Protocol = "tcp"
	}
	
	
	if !FreeForward(newForward.LocalPort, newForward.Protocol) {
		return 0
	}
	//开启事务
	tx := db.Begin()
	if tx.Error != nil {
		log.Println("开启事务失败")
		return 0
	}
	// 在事务中执行插入操作
	if err := tx.Create(&newForward).Error; err != nil {
		log.Println("插入新转发失败")
		log.Println(err)
		tx.Rollback() // 回滚事务
		return 0
	}
	// 提交事务
	tx.Commit()
	return newForward.Id
}

// 删除转发
func DelForward(id int) bool {
	if err := db.Where("id = ?", id).Delete(&conf.ConnectionStats{}).Error; err != nil {
		log.Println(err)
		return false
	}
	return true
}

// 增加错误登录
func AddBan(ip conf.IpBan) bool {
	//开启事务
	tx := db.Begin()
	if tx.Error != nil {
		return false
	}
	// 在事务中执行插入操作
	if err := tx.Create(&ip).Error; err != nil {
		log.Println(err)
		tx.Rollback() // 回滚事务
		return false
	}
	// 提交事务
	tx.Commit()
	return true
}

// 检查过去一天内指定IP地址的记录条数是否超过三条
func IpFree(ip string) bool {
	// 获取过去一天的时间戳
	oneDayAgo := time.Now().Add(-24 * time.Hour).Unix()

	// 查询过去一天内指定IP地址的记录条数
	var count int64
	if err := db.Model(&conf.IpBan{}).Where("ip = ? AND time_stamp > ?", ip, oneDayAgo).Count(&count).Error; err != nil {
		log.Println(err)
		return false
	}

	// 如果记录条数超过三条，则返回false；否则返回true
	return count < 3
}
