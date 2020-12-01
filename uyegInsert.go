package main

import (
	"log"
	"reflect"
	"strings"
	"time"

	client "github.com/Heo-youngseo/influxdb1-client/v2"
)

//influxDB 사용하기 위한 모듈을 import, DB 연결 정보를 전역변수로 선언
const (
	database = "test"
	username = "its"
	password = "its@1234"
)

//influxDB 연결 & 수집된 데이터를 InfluxDB에 저장하는 함수 실행
func influxDataInsert(chInserData chan map[string]interface{}) {
	for {
		select {
		case <-chInserData:
			c := influxDBClient()
			createMetrics(c, chInserData)
		}
	}
}

//influxDB에 연결시키는 함수
func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086/",
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

//수집한 데이터를 influxdb에 저장하는 함수, 데이터를 받으면 새로운 배치포인트를 만듦
func createMetrics(c client.Client, chInserData chan map[string]interface{}) {

	for {
		select {
		case data := <-chInserData:
			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:  database,
				Precision: "ms",
			})

			if err != nil {
				log.Fatalln("Error: ", err)
			}

			values := data["Values"]

			//1초 데이터 key 오름차순으로 정렬
			keySec := orderKey(data)
			tempStrSec := strings.Join(keySec[:], ",")
			tempStrSec = strings.Replace(tempStrSec, ",time", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",ver", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",gateway", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",mac", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",Values", "", 1)
			arrKeySec := strings.Split(tempStrSec, ",")

			switch reflect.TypeOf(values).Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(values)

				//0.1초 데이터 key 오름차순으로 정렬
				keyMilli := orderKey(s.Index(0).Interface().(map[string]interface{}))
				tempStrMilli := strings.Join(keyMilli[:], ",")
				tempStrMilli = strings.Replace(tempStrMilli, "time", "DataSavedTime", 1)
				tempStrMilli = strings.Replace(tempStrMilli, "420", "`420`", 1)

				//insert문 생성
				for i := 0; i < s.Len(); i++ {

					dataMilli := s.Index(i).Interface().(map[string]interface{})

					//clusterIndex := rand.Intn(len(dataMilli))
					tags := map[string]string{
						"mac":     data["mac"].(string),
						"gateway": data["gateway"].(string),
					}

					// fields := map[string]interface{}{
					// 	"cpu_usage":  rand.Float64() * 100.0,
					// 	"disk_usage": rand.Float64() * 100.0,
					// }

					fields := make(map[string]interface{})
					for j := 0; j < len(arrKeySec); j++ {
						fields[arrKeySec[j]] = data[arrKeySec[j]]
					}

					for j := 0; j < len(keyMilli); j++ {
						fields[keyMilli[j]] = dataMilli[keyMilli[j]]
					}

					date := dataMilli["time"].(string)
					t, err := time.Parse("2006-01-02 15:04:05.000", date)

					point, err := client.NewPoint(
						"SmartEOCR",
						tags,
						fields,
						t,
					)

					if err != nil {
						log.Fatalln("Error: ", err)
					}

					bp.AddPoint(point)
				}
			}
			//_ = bp
			//fmt.Println(bp)
			go func() {
				err = c.Write(bp)
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
	}
}

// func orderKey(m map[string]interface{}) []string {
// 	keys := make([]string, 0, len(m))
// 	for key := range m {
// 		keys = append(keys, key)
// 	}
// 	sort.Strings(keys) //sort by key
// 	return keys
// }
