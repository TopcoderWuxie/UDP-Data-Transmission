package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var receiveStrings string

var db = &sql.DB{}

var now = time.Now()
var time1 = now.Unix()
var time2 = now.Unix()
var logFile = "logInfo.log"
var debugLog *log.Logger

var str = make(chan string, 1000000)
var str2 = make(chan [][]string, 100000)

func init() {
	db, _ = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/udp_data?charset=utf8")
}

func main() {
	logFile, _ := os.Create(logFile)
	debugLog = log.New(logFile, "[Info]", log.LstdFlags)

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	checkErr(err)

	debugLog.Println(udpAddr)
	socket, err := net.ListenUDP("udp", udpAddr)
	checkErr(err)

	defer socket.Close()

	for {
		receiveData := make([]byte, 65535)
		n, udpAddr, err := socket.ReadFromUDP(receiveData)
		checkErr(err)
		if n == 0 {
			continue
		}

		receiveStrings += string(receiveData[:n])
		receiveStrings += "\n"

		for {
			now = time.Now()
			time2 = now.Unix()
			if time2-time1 < 10 {
				break
			}

			time1 = time2
			str <- receiveStrings
			go handleData()
			receiveStrings = ""
			goto Label
		}
	Label:
		sendData := []byte("本次数据接收完毕")
		_, err = socket.WriteToUDP(sendData, udpAddr)
		checkErr(err)
	}
}

func handleData() {
	data := <-str
	datas := strings.Split(data, "\n")

	var everyStrings [][]string

	for _, data := range datas {
		if len(data) == 0 {
			continue
		}
		s := strings.Split(data, "|")
		if len(s) != 6 {
			continue
		}
		everyStrings = append(everyStrings, s)
	}
	if len(everyStrings) != 0 {
		str2 <- everyStrings
		go insertData()
	}

	go countData()
}

func insertData() {
	datas := <-str2
	var s string = "insert into udp(`ip`, `requested`, `traced`, `api`, `mean`, `comment`) values"

	for x, data := range datas {
		s += " (\"" + data[0] + "\",\"" + data[1] + "\",\"" + data[2] + "\",\"" + data[3] + "\",\"" + data[4] + "\",\"" + data[5] + "\")"
		if (x + 1) != len(datas) {
			s += ","
		} else {
			s += ";"
		}
	}
	_, err := db.Exec(s)
	checkErr(err)
	debugLog.Println("数据插入成功")
}

func countData() {
	var rowCounts int64
	res := db.QueryRow("SELECT COUNT(*) FROM udp;")
	res.Scan(&rowCounts)
	debugLog.Printf("当前数据的行数：%d\n", rowCounts)
}

func checkErr(err error) {
	if err != nil {
		debugLog.SetPrefix("[Error]")
		debugLog.Fatalln(err)
	}
}
