package main

import (
	"fmt"
	//"log"
	"math"
	//"os"
	"merge/uyeg"
	"strings"
	"time"
)

// UYeGProcessing 함수는 수집된 데이터를 정제, 처리하는 함수
func UYeGProcessing(client *uyeg.ModbusClient, collChan <-chan map[string]interface{}, tfChan chan<- []interface{}) {

	var queue ItemQueue
	if queue.items == nil {
		queue = ItemQueue{}
		queue.New()
	}

	go QueueProcess(client, &queue, tfChan)

	for {
		select {
		case <-client.Done2:
			fmt.Println(fmt.Sprintf("=> %s (%s:%d) 데이터 처리 종료", client.Device.MacId, client.Device.Host, client.Device.Port))
			return
		case data := <-collChan:
			queue.Enqueue(data)

		}
		time.Sleep(1 * time.Millisecond)
	}
}

//수집된 데이터를 처리
func QueueProcess(client *uyeg.ModbusClient, queue *ItemQueue, tfChan chan<- []interface{}) {
	wsyncMap := wSyncMap{vi: make(map[string]interface{})}
	ds := make([]interface{}, 0, 100)   // 미리 공간 할당해둠1
	var lastData map[string]interface{} // 미리 공간 할당해둠2

	//var preTime  *string = new(string)

	for {
		for len(queue.items) > 0 {
			data := (*queue).Dequeue()

			t := (*data).(map[string]interface{})["time"].(time.Time).Truncate(time.Duration(client.Device.ProcessInterval) * time.Millisecond).Format(TimeFormat)

			if v := wsyncMap.wGet(t); v != nil { // 데이터가 있는경우

				tv := make(map[string]interface{})
				for k, v := range v.(map[string]interface{}) {
					var tmp float64
					if strings.Contains(k, "time") {
						tv[k] = v
						continue
					} else if strings.Contains(k, "Volt") {
						tmp = math.Min(v.(float64), (*data).(map[string]interface{})[k].(float64))
					} else {
						tmp = math.Max(v.(float64), (*data).(map[string]interface{})[k].(float64))
					}
					tv[k] = tmp
				}
				wsyncMap.wSet(t, tv)

				//}
			} else {
				(*data).(map[string]interface{})["time"] = t
				wsyncMap.wSet(t, (*data).(map[string]interface{}))

				tmillisecond := t[len(t)-4:]
				t2, _ := time.Parse(TimeFormat[:len(TimeFormat)-4], t)
				if tmillisecond == ".000" && wsyncMap.wSize() >= 10 {
					sMap := wsyncMap.wGetMap()
					bSecT, _ := time.Parse(TimeFormat[:len(TimeFormat)-4], t2.Add(-1 * time.Second).Format(TimeFormat)[:len(TimeFormat)-4])

					// .000 부터 데이터 비교.
					for i := 0; i < 1000/client.Device.ProcessInterval; i++ {
						vT := bSecT.Add(time.Duration(i*client.Device.ProcessInterval) * time.Millisecond).Format(TimeFormat)
						if val, exists := sMap[vT]; exists == true { // 데이터가 있는 경우.
							value := val.(map[string]interface{})
							value["status"] = true
							ds = append(ds, value)
							lastData = val.(map[string]interface{}) // 마지막 데이터를 초기화 시킨다.
							wsyncMap.wDelete(vT)                      // 추가한 데이터는 삭제한다.
						} else { // 데이터가 없는 경우.
							if lastData != nil {
								//마지막 데이터를 복사해서 가져옴
								ld := CopyMap(lastData)
								ld["time"] = vT
								ld["status"] = false
								ds = append(ds, ld)
							}
							fmt.Println(" No Data ", vT)

							//데이터가  없을 시 에러메세지를 출력 및 로그 저장
							derr := make(map[string]interface{})
							derr["Device"] = client.Device
							derr["Error"] = fmt.Sprintf(" No Data ", vT)
							derr["Restart"] = false

							ErrChan <- derr
						}
					}

					//tfChan 채널로 uyegTransfer 보냄
					tfChan <- ds
					ds = ds[:0] // 데이터 삭제
				}
			}
			time.Sleep(1 * time.Millisecond)
		}
		time.Sleep(1 * time.Millisecond)
	}
}
