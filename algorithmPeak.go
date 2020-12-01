package main

import (
	"fmt"
	"math"
	"time"
)

func (input *PeakInput) SetPeakInput(_tagName string) {
	input.offset = mapMstDevice[_tagName].(map[string]interface{})["Offset"].(float64)
	input.peakPeriod = mapMstDevice[_tagName].(map[string]interface{})["PeakPeriod"].(float64) * 60000
	input.meanException = mapMstDevice[_tagName].(map[string]interface{})["MeanException"].(float64)
	input.peakWarningSet = mapMstDevice[_tagName].(map[string]interface{})["PeakWarningSet"].(float64)
	input.peakFaultSet = mapMstDevice[_tagName].(map[string]interface{})["PeakFaultSet"].(float64)
	input.resetPeriod = mapMstDevice[_tagName].(map[string]interface{})["ResetPeriod"].(float64) * 60000
	input.limitAlarmMsec = mapMstDevice[_tagName].(map[string]interface{})["LimitAlarmMsec"].(float64)
	input.peakFaultTimes = mapMstDevice[_tagName].(map[string]interface{})["PeakFaultTimes"].(float64)
	input.peakWarningTimes = mapMstDevice[_tagName].(map[string]interface{})["PeakWarnTimes"].(float64)
	input.peakFaultMsec = mapMstDevice[_tagName].(map[string]interface{})["PeakFaultMsec"].(float64)
	input.peakWarningMsec = mapMstDevice[_tagName].(map[string]interface{})["PeakWarnMsec"].(float64)
	input.endCount = mapMstDevice[_tagName].(map[string]interface{})["EndCount"].(float64)
}

var peak map[string]*PeakProcessVariable

func (pv *PeakProcessVariable) peakExtraction(date string, _tagValue float64, _tagName string) {
	input := new(PeakInput)
	input.SetPeakInput(_tagName)

	t, err := time.Parse("2006-01-02 15:04:05.000", date)
	if err != nil {
		panic(err)
	}

	_nowTime := t.UnixNano() / int64(time.Millisecond)

	if _tagValue < input.peakWarningSet {
		if pv.peakFaultStart != 0 && pv.peakFaultStart+int64(input.peakFaultMsec) <= _nowTime {
			pv.peakFaultCnt++
			if int32(input.peakFaultTimes) >= pv.peakFaultCnt {
				pv.alarmClass = "Fault"
				strAlarm := fmt.Sprintf("%s/Peak/%s/%f/Count/%s/%d", date, pv.alarmClass, input.peakFaultSet, _tagName, pv.peakFaultCnt)
				outputAlarm(strAlarm)
			}
		}

		if pv.peakWarningStart != 0 && pv.peakWarningStart+int64(input.peakWarningMsec) <= _nowTime {
			pv.peakWarningCnt++
			if int32(input.peakWarningTimes) >= pv.peakWarningCnt {
				pv.alarmClass = "Warning"
				strAlarm := fmt.Sprintf("%s/Peak/%s/%f/Count/%s/%d", date, pv.alarmClass, input.peakWarningSet, _tagName, pv.peakWarningCnt)
				outputAlarm(strAlarm)
			}
		}

		pv.peakWarningStart = 0
		pv.peakFaultStart = 0
	}

	if _tagValue > input.offset {
		//fmt.Println(_tagValue, "\t", pv.peakEnd)
		if pv.waveformstart == 0 {
			pv.waveformstart = _nowTime
		}

		if pv.peakPeriodStart == 0 {
			pv.peakPeriodStart = _nowTime
		}

		pv.endCountStart = 0

		if input.resetPeriod > 0 {
			if pv.resetStart == 0 {
				pv.resetStart = _nowTime
			}

			if pv.resetStart+int64(input.resetPeriod) < _nowTime {
				if pv.maxValue != 0 {
					strExtaction := fmt.Sprintf("Peak/%s/%f/%s/", date, pv.maxValue, _tagName)
					outputExtraction(strExtaction)
					pv.maxPerValue, pv.maxPerTime = timePeakExtraction(pv.maxValue, _nowTime, pv.maxPerValue, pv.maxPerTime, _tagName)
				}

				pv.waveformstart = 0
				pv.resetStart = 0
				pv.peakWarningStart = 0
				pv.peakFaultStart = 0
				pv.maxValue = 0
				pv.maxValueTime = ""
				pv.peakEnd = 0
			}
		}

		if int64(input.meanException) < (_nowTime-pv.waveformstart) && pv.peakWarningStart == 0 && pv.peakFaultStart == 0 && input.peakWarningSet > _tagValue {
			pv.peakEnd = 1
		}

		if pv.peakEnd == 1 {
			return
		}
		if input.peakFaultSet <= _tagValue {
			if pv.peakFaultStart == 0 {
				pv.peakFaultStart = _nowTime
			}

			if pv.peakFaultStart+int64(input.limitAlarmMsec) <= _nowTime {
				pv.alarmClass = "Fault"
				strAlarm := fmt.Sprintf("%s/Peak/%s/%f/OverLoad/%s/%f", date, pv.alarmClass, _tagValue, _tagName, input.limitAlarmMsec)
				outputAlarm(strAlarm)
				pv.peakWarningStart = 0
				pv.peakFaultStart = 0
			}

		} else if input.peakWarningSet <= _tagValue && _tagValue < input.peakFaultSet {
			if pv.peakWarningStart == 0 {
				pv.peakWarningStart = _nowTime
				// if _tagName == "2C6A6FB028D7.Curr" {
				// 	fmt.Println("warningStart : ", _tagName, pv.peakWarningStart, _nowTime, "\ttag값", _tagValue, input.peakWarningSet)
				// }
			}

			if pv.peakWarningStart+int64(input.limitAlarmMsec) <= _nowTime {
				pv.alarmClass = "Warning"
				strAlarm := fmt.Sprintf("%s/Peak/%s/%f/OverLoad/%s/%f", date, pv.alarmClass, _tagValue, _tagName, input.limitAlarmMsec)
				outputAlarm(strAlarm)
				pv.peakWarningStart = 0
			}

			if pv.peakFaultStart != 0 && pv.peakFaultStart+int64(input.peakFaultMsec) <= _nowTime {
				pv.peakFaultCnt++
				pv.peakWarningStart = 0

				if int32(input.peakFaultTimes) >= pv.peakFaultCnt {
					pv.alarmClass = "Fault"
					strAlarm := fmt.Sprintf("%s/Peak/%s/%f/Count/%s/%d", date, pv.alarmClass, input.peakFaultSet, _tagName, pv.peakFaultCnt)
					outputAlarm(strAlarm)
				}
			}

			pv.peakFaultStart = 0
		}

		pv.maxValue = math.Max(_tagValue, pv.maxValue)
		if pv.maxValue == _tagValue {
			pv.maxValueTime = date
		}

	} else {
		if pv.endCountStart == 0 {
			pv.endCountStart = _nowTime
		}

		if pv.endCountStart+int64(input.endCount) > _nowTime {
			return
		} else {
			pv.maxPerValue, pv.maxPerTime = timePeakExtraction(pv.maxValue, _nowTime, pv.maxPerValue, pv.maxPerTime, _tagName)
		}

		if pv.maxValue != 0 {
			strExtaction := fmt.Sprintf("Peak/%s/%f/%s/", pv.maxValueTime, pv.maxValue, _tagName)
			outputExtraction(strExtaction)
		}

		pv.waveformstart = 0
		pv.resetStart = 0
		pv.peakWarningStart = 0
		pv.peakFaultStart = 0
		pv.maxValue = 0
		pv.peakEnd = 0

		if pv.peakPeriodStart+int64(input.peakPeriod) >= _nowTime {
			pv.peakPeriodStart = 0
			pv.peakFaultCnt = 0
			pv.peakWarningCnt = 0
		}
	}
}

func timePeakExtraction(maxValue float64, _nowTime int64, maxPerValue [4]float64, maxPerTime [4]int64, _tagName string) ([4]float64, [4]int64) {
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
			}

		} else if maxPerTime[i] == _nowTime {
			//비교 후 insert
			if maxPerValue[i] < maxValue {
				maxPerValue[i] = maxValue
			}
			lastTimeStr := time.Unix(0, maxPerTime[i]*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
			strExtaction := fmt.Sprintf("Peak/%s/%f/%s/%s", lastTimeStr, maxPerValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)
			maxPerValue[i] = 0
			maxPerTime[i] = 0

		} else {
			//insert 후 새로운 값
			lastTimeStr := time.Unix(0, maxPerTime[i]*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
			strExtaction := fmt.Sprintf("Peak/%s/%f/%s/%s", lastTimeStr, maxPerValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)

			if maxValue != 0 {
				maxPerValue[i] = maxValue
				maxPerTime[i] = (_nowTime/divTime[i] + 1) * divTime[i]
			} else {
				maxPerValue[i] = 0
				maxPerTime[i] = 0
			}
		}
	}
	return maxPerValue, maxPerTime
}
