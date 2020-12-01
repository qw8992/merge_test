package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tempArr      []map[string]interface{}
	prevSec      int64
	tagName      []string
	mapMstDevice map[string]interface{}
	someMap      = map[string]string{}
    someMapMutex = sync.RWMutex{}
)

func preprocessing(chRawData chan map[string]interface{}) {

	//map variable define
	peak = make(map[string]*PeakProcessVariable)
	mean = make(map[string]*MeanProcessVariable)
	level = make(map[string]*LevelProcessVariable)
	sqtime = make(map[string]*SQTimeProcessVariable)

	//MstDevice Load
	startTime := time.Now()
	go mstDeviceLoad()

	for i := 0; i < 6; i++ {
		dataTime[i] = make(map[string]float64)
	}

	for {
		if len(tagName) != 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf(" SetupDevice.SelectTagNames  : %s \n", elapsedTime)

	// var mutex = &sync.Mutex{}

	for {
		select {
		case data := <-chRawData:
			//while date is piled wait
			// mutex.Lock()
			someMapMutex.Lock()
			timeDataDiv(data)
			someMapMutex.Unlock()
			// mutex.Unlock()
		}
	}
}

func timeDataDiv(preData map[string]interface{}) {

	//var tempArr []map[string]interface{}
	date := preData["time"].(string)

	if strings.Contains(preData["time"].(string), "UTC") {
		fmt.Println("DateError : ", preData["time"].(string))
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", date)

	if err != nil {
		panic(err)
	}

	nowSec := t.Unix()

	if prevSec == 0 {
		prevSec = nowSec
	}
	//fmt.Println("tick : ",nowSec)
	if nowSec != prevSec {
		prevSec = nowSec
		dataArr := tempArr

		// fmt.Println(dataArr[0]["time"].(string))
		//Processing Piled data
		timeDataProcess(dataArr)
		tempArr = tempArr[:0]
		tempArr = append(tempArr, preData)

	} else {
		tempArr = append(tempArr, preData)
		return
	}
}

func timeDataProcess(_arrSecData []map[string]interface{}) {
	startTime := time.Now()
	arrSecData := _arrSecData

	for i := 0; i < 10; i++ {
		dataMilli := make(map[string]interface{})
		myMapHandler := &SyncMap{v: dataMilli}

		for cnt := 0; cnt < len(arrSecData); cnt++ {
			keyValue := secData(arrSecData[cnt])
			tempValues := arrSecData[cnt]["Values"]

			switch reflect.TypeOf(tempValues).Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(tempValues)
				if s.Len() <= i {
					continue
				}
				tempMilli := s.Index(i).Interface().(map[string]interface{})

				if strings.Contains(tempMilli["time"].(string), "UTC") {
					fmt.Println(tempMilli["time"].(string))
					continue
				} else {
					for keys := range tempMilli {
						if cnt == 0 && keys == "time" {
							myMapHandler.Set("time", tempMilli[keys])
						}
						if (keys != "time") && (strings.Contains(keys, "CPU") == false) && (strings.Contains(keys, "Memory") == false) {
							tagName := fmt.Sprintf("%s.%s", arrSecData[cnt]["mac"].(string), keys)
							myMapHandler.Set(tagName, tempMilli[keys])
						}

						// if keys == "status" && tempMilli[keys].(bool) == false {
						// 	tagName := fmt.Sprintf("%s.%s", arrSecData[cnt]["mac"].(string), keys)
						// 	fmt.Println(tempMilli["time"], "\t", tagName, "\t", tempMilli[keys])
						// }
					}

					for k := 0; k < len(keyValue); k++ {
						tagName := fmt.Sprintf("%s.%s", arrSecData[cnt]["mac"].(string), keyValue[k])
						myMapHandler.Set(tagName, arrSecData[cnt][keyValue[k]])
					}
				}

			default:
				fmt.Println("Body Error")
			}
		}
		if dataMilli["time"] == nil {
			continue
		}
		copyData := myMapHandler.GetMap()
		algorithmRun(copyData)
		timeDataQuery(copyData)
	}
	_ = time.Since(startTime)
	// elapsedTime := time.Since(startTime)
	// elapsedTime = time.Since(startTime)
	//fmt.Printf("\n 실행시간 : %s \n", elapsedTime)
}


func mstDeviceLoad() {
	for {
		tempMapMstDevice := make(map[string]interface{})
		whereQuery := " SaveRate > 0 and DefServer is not null and DefTable is not null and DefColumn is not null and CollectScale is not null and Offset is not null"
		whereQuery += " and Level1 is not null and Level2 is not null and Level3 is not null and Level4 is not null and Level5 is not null"
		selectQuery := fmt.Sprintf("Select * From MstDevice where %s and gatewayID = '%s' order by mac asc", whereQuery, gatewayID)
		tempTagName, tempMapMstDevice := dbConn.SelectDataInsertQuery(selectQuery)
		//tagName, mapMstDevice = dbConn.SelectDataInsertQuery(selectQuery)

		selectquery := "SELECT mac, DefTable FROM MstDevice where MeanException is not null and PeakPeriod is not null and PeakFaultSet is not null and PeakWarningSet is not null"
		selectquery += " and LimitAlarmMsec > 0 and PeakFaultTimes is not null and PeakWarnTimes is not null and PeakWarnMsec > 0 and PeakFaultMsec > 0 and EndCount is not null"
		selectquery = fmt.Sprintf("%s and gatewayID = '%s' order by mac asc", selectquery, gatewayID)
		peakMst := dbConn.AlgorithmCheck(selectquery)
		tempMapMstDevice = algorithmSet(peakMst, tempMapMstDevice, 2)

		selectquery = "Select mac, DefTable from MstDevice where MeanException is not null and MeanPeriod is not null and MeanPercent is not null and MeanDuration is not null"
		selectquery += " and MeanFaultSet is not null and MeanWarningSet is not null and LimitAlarmMsec > 0 and MeanFaultTimes is not null and MeanWarnTimes is not null"
		selectquery += " and MeanFaultMsec > 0 and MeanWarnMsec > 0 and EndCount is not null"
		selectquery = fmt.Sprintf("%s and gatewayID = '%s' order by mac asc", selectquery, gatewayID)
		meanMst := dbConn.AlgorithmCheck(selectquery)
		tempMapMstDevice = algorithmSet(meanMst, tempMapMstDevice, 3)

		selectquery = "SELECT mac, DefTable FROM MstDevice where AlmPeriod is not null and ResetPeriod > 0 and HighFault is not null and HighWarning is not null"
		selectquery += " and LowFault is not null and LowWarning is not null and LevelLimitAlarmMsec > 0 and HFAlmTimes is not null and HWAlmTimes is not null and LFAlmTimes is not null and LWAlmTimes is not null" //예외처리 필요하면 추가
		selectquery += " and HFAlmMsec > 0 and HWAlmMsec > 0 and LFAlmMsec > 0 and LWAlmMsec > 0"
		selectquery = fmt.Sprintf("%s and gatewayID = '%s' order by mac asc", selectquery, gatewayID)
		levelMst := dbConn.AlgorithmCheck(selectquery)
		tempMapMstDevice = algorithmSet(levelMst, tempMapMstDevice, 5)

		selectquery = "Select mac, DefTable from MstDevice where EndCount is not null and ResetPeriod is not null and Offset > 0"
		selectquery = fmt.Sprintf("%s and gatewayID = '%s' order by mac asc", selectquery, gatewayID)
		sqtimeMst := dbConn.AlgorithmCheck(selectquery)
		tempMapMstDevice = algorithmSet(sqtimeMst, tempMapMstDevice, 7)

		if !reflect.DeepEqual(tempMapMstDevice, mapMstDevice) {
			myMapHandler := &SyncArrMap{v: dataTime}

			for j := 0; j < len(tempTagName); j++ {
				found := myMapHandler.Select(tempTagName[j])
				if !found {
					peak[tempTagName[j]] = &PeakProcessVariable{}
					mean[tempTagName[j]] = &MeanProcessVariable{}
					level[tempTagName[j]] = &LevelProcessVariable{}
					sqtime[tempTagName[j]] = &SQTimeProcessVariable{}
				}
			}

			tagName = tempTagName
			myCopyHandler := &SyncMap{v: mapMstDevice}
			myCopyHandler.MoveMap(tempMapMstDevice)
		}

		time.Sleep(10 * time.Second)
	}
}

func algorithmSet(altagName []string, tempMapMstDevice map[string]interface{}, algorithmType int8) map[string]interface{} {
	for i := 0; i < len(altagName); i++ {
		if tempMapMstDevice[altagName[i]] != nil {
			tempMapMstDevice[altagName[i]].(map[string]interface{})["Item"] = tempMapMstDevice[altagName[i]].(map[string]interface{})["Item"].(int8) * algorithmType
		}
	}
	return tempMapMstDevice
}

func secData(data map[string]interface{}) []string {
	keySec := orderKey(data)
	tempStrSec := strings.Join(keySec[:], ",")
	tempStrSec = strings.Replace(tempStrSec, ",time", "", 1)
	tempStrSec = strings.Replace(tempStrSec, ",gateway", "", 1)
	tempStrSec = strings.Replace(tempStrSec, ",ver", "", 1)
	tempStrSec = strings.Replace(tempStrSec, ",Values", "", 1)
	tempStrSec = strings.Replace(tempStrSec, ",mac", "", 1)
	arrKeySec := strings.Split(tempStrSec, ",")
	return arrKeySec
}

func getFloat(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		return float64(0), nil
	}
}

func orderKey(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys) //sort by key
	return keys
}
