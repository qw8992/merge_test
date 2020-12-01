package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"scada/config"
	"scada/db"
	"scada/uyeg"
	"sync"
	"syscall"
	"time"
)

var done = make(chan bool, 1)
var ErrChan = make(chan map[string]interface{}, 10)
var conf = config.GetConfiguration()
var dbConn = db.DataBase{
	Host:     conf.MYSQL_HOST,
	Port:     conf.MYSQL_PORT,
	Database: conf.MYSQL_DATABASE,
	User:     conf.MYSQL_USER,
	Password: conf.MYSQL_PASSWORD,
}
var gatewayID = "local"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("MAX PROCS", runtime.GOMAXPROCS(0))
	fmt.Println("\n===================")
	fmt.Println("Start Scada Program")
	fmt.Println("===================")

	wg := sync.WaitGroup{}
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	dbConn.Connect()     // 데이터베이스 연결
	defer dbConn.Close() // 데이터베이스 자동 해제
	wg.Add(1)
	go func() {
		<-sigs

		done <- true // 프로그램 종료 신호를 보냄

		dbConn.Close()

		fmt.Println("\n===================")
		fmt.Println("Stop Scada Program")
		fmt.Println("===================")
		os.Exit(0)
		wg.Done()
	}()

	go startProgram() // 프로그램 시작
	wg.Wait()
}

func startProgram() {
	addedDs := make(map[int]uyeg.Device)
	connDs := make(map[int]*uyeg.ModbusClient)
	for {
		select {
		case <-done: // 프로그램 종료 신호가 들어옴
			for _, client := range connDs {
				client.Close()
			}
		case derr := <-ErrChan:
			id := derr["Device"].(uyeg.Device).Id
			if derr["Restart"].(bool) {
				connDs[ id].Close()
				delete(connDs, id)
				delete(addedDs, id)
			}

		default:
			//DB에 설정된 장치를 읽어옴
			devices := GetEnabledDevices(&dbConn)
			if reflect.DeepEqual(addedDs, devices) {
				fmt.Println(time.Now().In(Loc).String(), fmt.Sprintf(" - 모든 장치가 연결됨. (%d 개) addedDs", len(addedDs)))
			} else {
				fmt.Println(time.Now().In(Loc).String(), " - 연결되지 않은 장치 또는 변경된 장치가 있음.")

				// 삭제 및 변경된 경우 장치 삭제.
				for id, device := range addedDs {
					// 삭제 또는 변경되지 않은 경우 통과
					if reflect.DeepEqual(devices[id], device) {
						continue
					}
					fmt.Println("=>", fmt.Sprintf("This device(%s) has been removed from the list.\n", device.MacId))
					// 장치 삭제
					connDs[id].Close()
					delete(connDs, id)
					delete(addedDs, id)
				}

				// 장치 추가하기.
				for id, device := range devices {
					// 이미 추가된 장치인경우 통과.
					if reflect.DeepEqual(addedDs[id], device) {
						continue
					}

					client := new(uyeg.ModbusClient)
					client.Device = device
					client.Done1 = make(chan bool)
					client.Done2 = make(chan bool)
					client.Done3 = make(chan bool)

					addedDs[id] = device
					connDs[id] = client

					//uyegStart.go - UYeGStartFunc 호출
					go UYeGStartFunc(client)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func SelectGateway() (gatewayID string) {
	db, err := sql.Open("mysql", "root:its@1234@tcp(127.0.0.1:3306)/UYeG_BM4")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 하나의 Row를 갖는 SQL 쿼리
	DeviceQuery := fmt.Sprintf("SELECT Distinct gateway_ID FROM gateway")
	err = db.QueryRow(DeviceQuery).Scan(&gatewayID)
	if err != nil {
		// log.Fatal(err)
	}
	return gatewayID
}
