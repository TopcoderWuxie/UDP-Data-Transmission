package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// 十次的最大缓冲量为655350
var str = make(chan string, 655350)
var str2 = make(chan [][]string, 12)

var db = &sql.DB{}

// 存储监听多次的字符串
var sumString string
var num = 0

// 设置日志文件的参数
// 日志设置的详情：http://studygolang.com/articles/9184
var logFile = "logInfo.log"
var debugLog *log.Logger

func init() {
	/*
		连接数据库的参数设置
		http://blog.csdn.net/x369201170/article/details/27207893
		http://blog.csdn.net/jesseyoung/article/details/40398321
	*/
	db, _ = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/udp_data?charset=utf8")
}

func main() {
	logFile, _ := os.Create(logFile)

	debugLog = log.New(logFile, "[Info]", log.LstdFlags)

	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8080")
	checkErr(err)

	// 查看连接地址
	fmt.Println(udpAddr)
	socket, err := net.ListenUDP("udp", udpAddr)
	checkErr(err)

	// 等待关闭UDP连接
	defer socket.Close()

	// 持续监听
	for {
		data := make([]byte, 65507)

		/*
			readNumbers 记录读取的数据量
			remoteAddr 连接的地址
			data 保存读取的数据量
		*/
		readNumbers, remoteAddr, err := socket.ReadFromUDP(data)

		checkErr(err)

		// 读取的数据的数量为0的时候，表明发送的UDP数据包为空，不需要往下继续执行
		if readNumbers == 0 {
			continue
		}

		// 把读取到的数据记录下来，先不进行存储
		data1 := string(data[:readNumbers])
		sumString += data1
		sumString += "\n"
		num++
		// 设置接收10次数据存储一次
		if num%10 == 0 {
			debugLog.SetPrefix("[Info]")
			debugLog.Println("达到指定次数，开始存储数据")
			// 把批处理的数据推送到channel中
			str <- sumString
			// 创建协程
			go handleData()
			// 清空全部变量的字符串，进行下次获取字符串操作
			sumString = ""
		}

		sendData := []byte("本次数据接收完毕")
		_, err = socket.WriteToUDP(sendData, remoteAddr)
		checkErr(err)
	}
}

// 处理数据
func handleData() {
	// 取出管道中的数据
	data := <-str

	// 对数据进行处理
	datas := strings.Split(data, "\n")

	// 数据的临时存储，用于把数据放入管道
	var everyStrings [][]string

	for _, data := range datas {
		if len(data) == 0 {
			continue
		}
		// 对数据进行切割，然后存储
		s := strings.Split(data, "|")
		if len(s) != 6 {
			continue
		}
		everyStrings = append(everyStrings, s)
	}

	// 有数据的时候才执行写入操作，不然会报错
	if len(everyStrings) != 0 && len(everyStrings[0]) != 0 {
		str2 <- everyStrings
		go insertData()
	}

	go countData()
}

// 插入数据（一次性插入多条）
func insertData() {
	datas := <-str2

	var s string = "insert into udp(`ip`, `requested`, `traced`, `api`, `mean`, `comment`) values"
	// 多条语句的values值拼接
	for x, data := range datas {
		s += " (\"" + data[0] + "\",\"" + data[1] + "\",\"" + data[2] + "\",\"" + data[3] + "\",\"" + data[4] + "\",\"" + data[5] + "\")"
		if (x + 1) != len(datas) {
			s += ","
		} else {
			s += ";"
		}
	}
	// 执行插入sql的操作
	_, err := db.Exec(s)
	checkErr(err)
}

// 插入数据（一次性插入一条）
func insertData1() {
	datas := <-str2

	// db.Prepare() 返回一个Stmt。Stmt对象可以执行Exec, Query, QueryRow等操作
	stmt, err := db.Prepare("insert into udp(`ip`, `requested`, `traced`, `api`, `mean`, `comment`) values(?, ?, ?, ?, ?, ?)")

	checkErr(err)
	defer stmt.Close()
	for _, data := range datas {
		// 插入数据
		stmt.Exec(data[0], data[1], data[2], data[3], data[4], data[5])
	}
}

// 统计当前数据表中数据一共有多少行
func countData() {
	var rowCounts int64
	res := db.QueryRow("SELECT COUNT(*) FROM udp;")
	res.Scan(&rowCounts)
	debugLog.Printf("当前数据的行数：%d\n", rowCounts)
}

// 错误判断
func checkErr(err error) {
	if err != nil {
		debugLog.SetPrefix("[Error]")
		debugLog.Fatalln(err)
	}
}
