package main

import (
	"encoding/json"
	"fmt"

	//"log"
	//"os"
	"scada/uyeg"
)

//수집 및 처리한 데이터를
func UYeGTransfer(client *uyeg.ModbusClient, tfChan <-chan []interface{}, chInsertData chan map[string]interface{}, chRawData chan map[string]interface{}) {
	for {
		select {
		case <-client.Done3:
			fmt.Println(fmt.Sprintf("=> %s (%s:%d) 데이터 전송 종료", client.Device.MacId, client.Device.Host, client.Device.Port))
			return
		case data := <-tfChan:
			//수집한 데이터를 리맵데이터로 변환
			d := data[0].(map[string]interface{})
			if t, exists := d["time"]; exists {
				bSecT := t.(string)[:len(TimeFormat)-4]
				jsonBytes := client.GetRemapJson(bSecT, data)
				dataSecond := make(map[string]interface{})
				//변환된 데이터를 Map으로 변환해서 chInsertData 채널로 uyegInsert로 보냄
				json.Unmarshal(jsonBytes, &dataSecond)
				chInsertData <- dataSecond
				chRawData <- dataSecond
			}
		}
	}
}
