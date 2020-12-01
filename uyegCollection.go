package main

import (
	"fmt"
	//	"log"
	//	"os"
	"scada/uyeg"
	"time"
)

// UYeGDataCollection 함수는 데이터를 수집하는 함수입니다
func UYeGDataCollection(client *uyeg.ModbusClient, collChan chan<- map[string]interface{}) {
	var errCount, errCountConn = 0, 0
	ticker := time.NewTicker(10 * time.Millisecond)

	for {
		select {
		case <-client.Done1:
			fmt.Println(fmt.Sprintf("=> %s (%s:%d) 데이터 수집 종료", client.Device.MacId, client.Device.Host, client.Device.Port))
			return
		case <-ticker.C:
			//데이터를 수집
			readData := client.GetReadHoldingRegisters()

			//수집한 데이터가 없을 시 Error메세지 출력 및 로그 저장
			if readData == nil {
				ticker.Stop()
				errCount = errCount + 1
				fmt.Println(time.Now().In(Loc).Format(TimeFormat), fmt.Sprintf("Failed to read data Try again (%s:%d)..", client.Device.Host, client.Device.Port))
				if errCount > client.Device.RetryCount {
					client.Handler.Close()
					if client.Connect() {
						fmt.Println(time.Now().In(Loc).Format(TimeFormat), fmt.Sprintf("Succeded to reconnect the connection.. (%s:%d)..", client.Device.Host, client.Device.Port))
						errCount = 0
					} else {
						fmt.Println(time.Now().In(Loc).Format(TimeFormat), fmt.Sprintf("Failed to reconnect the connection.. (%s:%d)..", client.Device.Host, client.Device.Port))
						errCountConn = errCountConn + 1

						if errCountConn > client.Device.RetryConnFailedCount {
							derr := make(map[string]interface{})
							derr["Device"] = client.Device
							derr["Error"] = fmt.Sprintf("%s(%s): Connection failed..", client.Device.Name, client.Device.MacId)
							derr["Restart"] = false

							ErrChan <- derr

						}
					}
				}
				time.Sleep(time.Duration(client.Device.RetryCycle) * time.Millisecond)
				ticker = time.NewTicker(10 * time.Millisecond)
				continue
			} else {
				errCount = 0
			}

			//수집한 데이터를 채널로 UYeGProcessing 보냄
			collChan <- readData
		}
	}
}
