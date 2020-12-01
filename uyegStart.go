package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"merge/uyeg"
	"time"
)

func UYeGStartFunc(client *uyeg.ModbusClient) {
	defer func() {
		v := recover()

		if v != nil {
			derr := make(map[string]interface{})
			derr["Device"] = client.Device
			derr["Error"] = v
			derr["Restart"] = true

			ErrChan <- derr
		}
	}()

	//장치의 연결상태를 확인 - 연결실패 시 자이의 이름과 MacId를 포함한 Error 메세지를 출력 및 DB 로그 저장
	if !client.Connect() {
		derr := make(map[string]interface{})
		derr["Device"] = client.Device
		derr["Error"] = fmt.Sprintf("%s(%s): Connection failed", client.Device.Name, client.Device.MacId)
		derr["Restart"] = false

		ErrChan <- derr
	}

	//채널을 미리 할당
	collChan := make(chan map[string]interface{}, 20)
	tfChan := make(chan []interface{}, 20)
	chInsertData := make(chan map[string]interface{})
	chMsg := make(chan map[string]interface{})
	chRawData := make(chan map[string]interface{})

	go insertDB()

	go preprocessing(chRawData)
	//데이터 저장(InfluxDB)
	go influxDataInsert(chInsertData)
	//데이터 전송(API)
	go UYeGTransfer(client, tfChan, chInsertData, chRawData)
	//데이터 처리
	go UYeGProcessing(client, collChan, tfChan)
	//데이터 수집
	go UYeGDataCollection(client, collChan)

	// tick := time.Tick(1 * time.Second)
	go func() {
		for {
			select {
			case data := <-chMsg:
				chRawData <- data

				// case <-tick:
				// 	timeInsert()
				// 	// case <-ticks:
				// 	algorithmInsert()

			}
		}
	}()
}

func timeInsert() {
	bytes, err := ioutil.ReadFile("time.txt")
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if bytes != nil {
		s := string(bytes)
		startTime := time.Now()
		if s != "" {
			go dbConn.NotResultQueryExec(s)
			elapsedTime := time.Since(startTime)
			if elapsedTime > 1*time.Second {
				fmt.Printf("time 실행시간 : %s \n", elapsedTime)
			}

			d := []byte("")

			err = ioutil.WriteFile("time.txt", d, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func algorithmInsert() {
	bytes, err := ioutil.ReadFile("algorithm.txt")
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if bytes != nil {
		s := string(bytes)
		startTime := time.Now()
		if s != "" {
			go dbConn.NotResultQueryExec(s)
			elapsedTime := time.Since(startTime)
			if elapsedTime > 1*time.Second {
				fmt.Printf("algorithm 실행시간 : %s \n", elapsedTime)
			}

			d := []byte("")

			err = ioutil.WriteFile("algorithm.txt", d, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}
