package main

import (
	"fmt"
	"math"
	"time"
)

func (input *MeanInput) SetMeanInput(_tagName string) {
	input.offset = mapMstDevice[_tagName].(map[string]interface{})["Offset"].(float64)
	input.meanPeriod = mapMstDevice[_tagName].(map[string]interface{})["MeanPeriod"].(float64) * 60000
	input.meanException = mapMstDevice[_tagName].(map[string]interface{})["MeanException"].(float64)
	input.meanPercent = mapMstDevice[_tagName].(map[string]interface{})["MeanPercent"].(float64)
	input.meanDuration = mapMstDevice[_tagName].(map[string]interface{})["MeanDuration"].(float64)
	input.meanWarningSet = mapMstDevice[_tagName].(map[string]interface{})["MeanWarningSet"].(float64)
	input.meanFaultSet = mapMstDevice[_tagName].(map[string]interface{})["MeanFaultSet"].(float64)
	input.resetPeriod = mapMstDevice[_tagName].(map[string]interface{})["ResetPeriod"].(float64) * 60000
	input.limitAlarmMsec = mapMstDevice[_tagName].(map[string]interface{})["LimitAlarmMsec"].(float64)
	input.meanFaultTimes = mapMstDevice[_tagName].(map[string]interface{})["MeanFaultTimes"].(float64)
	input.meanWarningTimes = mapMstDevice[_tagName].(map[string]interface{})["MeanWarnTimes"].(float64)
	input.meanFaultMsec = mapMstDevice[_tagName].(map[string]interface{})["MeanFaultMsec"].(float64)
	input.meanWarningMsec = mapMstDevice[_tagName].(map[string]interface{})["MeanWarnMsec"].(float64)
	input.endCount = mapMstDevice[_tagName].(map[string]interface{})["EndCount"].(float64)
}

var mean map[string]*MeanProcessVariable

func (pv *MeanProcessVariable) meanExtraction(date string, _tagValue float64, _tagName string) {
	input := new(MeanInput)
	input.SetMeanInput(_tagName)

	t, err := time.Parse("2006-01-02 15:04:05.000", date)
	if err != nil {
		panic(err)
	}

	_nowTime := t.UnixNano() / int64(time.Millisecond)

	if _tagValue < input.meanWarningSet {
		if pv.meanFaultStart != 0 && pv.meanFaultStart+int64(input.meanFaultMsec) <= _nowTime {
			pv.meanFaultCnt++

			if int32(input.meanFaultTimes) >= pv.meanFaultCnt {
				pv.alarmClass = "Fault"
				strAlarm := fmt.Sprintf("%s/Mean/%s/%f/Count/%s/%d", date, pv.alarmClass, input.meanFaultSet, _tagName, pv.meanFaultCnt)
				outputAlarm(strAlarm)
			}
		}

		if pv.meanWarningStart != 0 && pv.meanWarningStart+int64(input.meanWarningMsec) <= _nowTime {
			pv.meanWarningCnt++

			if int32(input.meanWarningTimes) >= pv.meanWarningCnt {
				pv.alarmClass = "Warning"
				strAlarm := fmt.Sprintf("%s/Mean/%s/%f/Count/%s/%d", date, pv.alarmClass, input.meanWarningSet, _tagName, pv.meanWarningCnt)
				outputAlarm(strAlarm)
			}
		}

		pv.meanFaultStart = 0
		pv.meanWarningStart = 0
	}

	if _tagValue > input.offset {
		pv.endCountStart = 0

		if input.resetPeriod > 0 {
			if pv.resetStart == 0 {
				pv.resetStart = _nowTime
			}

			if pv.resetStart+int64(input.resetPeriod) < _nowTime {
				if pv.totalSumValues != 0 {
					strExtaction := fmt.Sprintf("Mean/%s/%f/%f/%s/", pv.meanValueTime, pv.meanValue, (pv.totalSumValues / pv.totalLength), _tagName)
					outputExtraction(strExtaction)
					pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime = timeMeanExtraction(pv.meanValue, (pv.totalSumValues / pv.totalLength), _nowTime, pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime, _tagName)
				}

				pv.waveformstart = 0
				pv.resetStart = 0
				pv.meanDurationStart = 0
				pv.meanWarningStart = 0
				pv.meanFaultStart = 0
				pv.meanValue = 0
				pv.meanValueTime = ""
				pv.meanStandard = 0
				pv.meanErr = 0
				pv.durationFlag = 0
				pv.sumValues = 0
				pv.totalSumValues = 0
				pv.totalLength = 0
			}
		}

		if pv.waveformstart == 0 {
			pv.waveformstart = _nowTime
		}

		if pv.waveformstart+int64(input.meanException) > _nowTime {
			return
		}

		pv.totalSumValues += _tagValue
		pv.totalLength += 1

		if pv.meanStandard == 0 {
			pv.meanStandard = _tagValue
		}

		pv.meanErr = (pv.meanStandard - _tagValue) / pv.meanStandard * 100
		if math.Abs(pv.meanErr) > input.meanPercent {
			pv.meanStandard = _tagValue //기준점을 새로 정함
			pv.sumValues = 0            //리셋
			pv.meanDurationStart = 0
			pv.durationCount = 0
			//Console.WriteLine("Mean 재측정");
		}

		if math.Abs(pv.meanErr) <= input.meanPercent && (pv.durationFlag == 0) { //오차율 범위 안이고 정속구간이 발견 된적 없다면
			if pv.meanDurationStart == 0 { //오차율 범위 안에 처음 들어왔다면
				pv.meanDurationStart = _nowTime
			}
			pv.sumValues += _tagValue //합계
			pv.durationCount++        //값의 갯수

			if pv.meanDurationStart+int64(input.meanDuration) < _nowTime { //정속구간의 끝 시간보다 현재시간이 크다면 정속구간의 평균값을 구함
				pv.meanValue = (pv.sumValues / float64(pv.durationCount))
				pv.meanValueTime = date
				pv.durationFlag++
				pv.durationCount = 0
				pv.sumValues = 0

				//Console.WriteLine(input.TagName+ ", Mean 구간 끝");
			}

		}

		if input.meanFaultSet <= _tagValue {
			if pv.meanFaultStart == 0 {
				pv.meanFaultStart = _nowTime
			}

			if pv.meanFaultStart+int64(input.limitAlarmMsec) <= _nowTime {
				pv.alarmClass = "Fault"
				strAlarm := fmt.Sprintf("%s/Mean/%s/%f/OverLoad/%s/%ds", date, pv.alarmClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
				outputAlarm(strAlarm)
				pv.meanWarningStart = 0
				pv.meanFaultStart = 0
			}

		} else if input.meanWarningSet <= _tagValue && _tagValue < input.meanFaultSet {
			if pv.meanWarningStart == 0 {
				pv.meanWarningStart = _nowTime
			}

			if pv.meanWarningStart+int64(input.limitAlarmMsec) <= _nowTime {
				pv.alarmClass = "Warning"
				strAlarm := fmt.Sprintf("%s/Mean/%s/%f/OverLoad/%s/%ds", date, pv.alarmClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
				outputAlarm(strAlarm)
				pv.meanWarningStart = 0
			}

			if pv.meanFaultStart != 0 && pv.meanFaultStart+int64(input.meanFaultMsec) <= _nowTime {
				pv.meanFaultCnt++
				pv.meanWarningStart = 0

				if int32(input.meanFaultTimes) >= pv.meanFaultCnt {
					pv.alarmClass = "Fault"
					strAlarm := fmt.Sprintf("%s/Mean/%s/%f/Count/%s/%d", date, pv.alarmClass, input.meanFaultSet, _tagName, pv.meanFaultCnt)
					outputAlarm(strAlarm)
				}
			}

			pv.meanFaultStart = 0
		}

	} else {

		if pv.endCountStart == 0 {
			pv.endCountStart = _nowTime
		}

		if pv.endCountStart+int64(input.endCount) > _nowTime {
			return
		} else {
			pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime = timeMeanExtraction(pv.meanValue, (pv.totalSumValues / pv.totalLength), _nowTime, pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime, _tagName)

		}

		if pv.totalSumValues != 0 {
			strExtaction := fmt.Sprintf("Mean/%s/%f/%f/%s/", pv.meanValueTime, pv.meanValue, (pv.totalSumValues / pv.totalLength), _tagName)
			outputExtraction(strExtaction)
		}
		pv.waveformstart = 0
		pv.resetStart = 0
		pv.meanDurationStart = 0
		pv.meanWarningStart = 0
		pv.meanFaultStart = 0
		pv.meanValue = 0
		pv.meanValueTime = ""
		pv.meanStandard = 0
		pv.meanErr = 0
		pv.durationFlag = 0
		pv.sumValues = 0
		pv.totalSumValues = 0
		pv.totalLength = 0

		if pv.meanPeriodStart+int64(input.meanPeriod) >= _nowTime {
			pv.meanPeriodStart = 0
			pv.meanFaultCnt = 0
			pv.meanWarningCnt = 0
		}
	}
}

func timeMeanExtraction(maxValue float64, maxTotalValue float64, _nowTime int64, maxPerValue [4]float64, maxPerTotalValue [4]float64, maxPerTime [4]int64, _tagName string) ([4]float64, [4]float64, [4]int64) {
	// 각초
	divTime := []int64{10000, 60000, 600000, 3600000}
	divTimeStr := []string{"10Seconds", "Minute", "10Minutes", "Hour"}
	for i := 0; i < len(divTime); i++ {
		if maxValue == 0 && maxPerTime[i] == 0 {
			continue
		}
		if maxPerTime[i] > _nowTime || maxPerTime[i] == 0 {
			//비교
			if maxPerTime[i] == 0 {
				maxPerTime[i] = (_nowTime/divTime[i] + 1) * divTime[i]
			}
			if maxPerValue[i] < maxValue {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
			}

		} else if maxPerTime[i] == _nowTime {
			//비교 후 insert
			if maxPerValue[i] < maxValue {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
			}
			lastTimeStr := time.Unix(0, maxPerTime[i]*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
			strExtaction := fmt.Sprintf("Mean/%s/%f/%f/%s/%s", lastTimeStr, maxPerValue[i], maxPerTotalValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)
			maxPerValue[i] = 0
			maxPerTotalValue[i] = 0
			maxPerTime[i] = 0

		} else {
			//insert 후 새로운 값
			lastTimeStr := time.Unix(0, maxPerTime[i]*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
			strExtaction := fmt.Sprintf("Mean/%s/%f/%f/%s/%s", lastTimeStr, maxPerValue[i], maxPerTotalValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)
			if maxValue != 0 {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
				maxPerTime[i] = (_nowTime/divTime[i] + 1) * divTime[i]
			} else {
				maxPerValue[i] = 0
				maxPerTotalValue[i] = 0
				maxPerTime[i] = 0
			}
		}
	}
	return maxPerValue, maxPerTotalValue, maxPerTime
}
