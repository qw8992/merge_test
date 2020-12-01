package main

import (
	"fmt"
	"time"
)

func (input *LevelInput) SetLevelInput(_tagName string) {
	fmt.Println(mapMstDevice[_tagName].(map[string]interface{})["LevelLimitAlarmMsec"].(float64))
	input.limitAlarmMsec = mapMstDevice[_tagName].(map[string]interface{})["LevelLimitAlarmMsec"].(float64)
	input.hfAlmTimes = mapMstDevice[_tagName].(map[string]interface{})["HFAlmTimes"].(float64)
	input.hwAlmTimes = mapMstDevice[_tagName].(map[string]interface{})["HWAlmTimes"].(float64)
	input.lfAlmTimes = mapMstDevice[_tagName].(map[string]interface{})["LFAlmTimes"].(float64)
	input.lwAlmTimes = mapMstDevice[_tagName].(map[string]interface{})["LWAlmTimes"].(float64)
	input.hfAlmMsec = mapMstDevice[_tagName].(map[string]interface{})["HFAlmMsec"].(float64)
	input.hwAlmMsec = mapMstDevice[_tagName].(map[string]interface{})["HWAlmMsec"].(float64)
	input.lfAlmMsec = mapMstDevice[_tagName].(map[string]interface{})["LFAlmMsec"].(float64)
	input.lwAlmMsec = mapMstDevice[_tagName].(map[string]interface{})["LWAlmMsec"].(float64)
	input.almPeriod = mapMstDevice[_tagName].(map[string]interface{})["AlmPeriod"].(float64) * 60000
	input.resetPeriod = mapMstDevice[_tagName].(map[string]interface{})["ResetPeriod"].(float64) * 60000
	input.hfSet = mapMstDevice[_tagName].(map[string]interface{})["HighFault"].(float64)
	input.hwSet = mapMstDevice[_tagName].(map[string]interface{})["HighWarning"].(float64)
	input.lfSet = mapMstDevice[_tagName].(map[string]interface{})["LowFault"].(float64)
	input.lwSet = mapMstDevice[_tagName].(map[string]interface{})["LowWarning"].(float64)
}

var level map[string]*LevelProcessVariable

func (pv *LevelProcessVariable) levelExtraction(date string, _tagValue float64, _tagName string) {
	input := new(LevelInput)
	input.SetLevelInput(_tagName)

	t, err := time.Parse("2006-01-02 15:04:05.000", date)
	if err != nil {
		panic(err)
	}

	_nowTime := t.UnixNano() / int64(time.Millisecond)

	if pv.almPeriodEnd == 0 {
		pv.almPeriodEnd = _nowTime + int64(input.almPeriod)
	}

	if pv.almPeriodEnd < _nowTime {
		// Console.WriteLine("peakperiod 리셋");
		pv.almPeriodEnd = 0
		pv.alarmCountHF = 0
		pv.alarmCountHW = 0
		pv.alarmCountLF = 0
		pv.alarmCountLW = 0
	}

	if input.resetPeriod > 0 {
		if pv.resetStart == 0 {
			pv.resetStart = _nowTime
		}

		if pv.resetStart+int64(input.resetPeriod) < _nowTime {
			pv.alarmCheckStartHF = 0
			pv.alarmCheckStartHW = 0
			pv.alarmCheckStartLF = 0
			pv.alarmCheckStartLW = 0
			pv.resetStart = 0
			pv.statusClass = ""
			pv.statusClassHL = ""
		}
	}

	pv.prevStatusClassHL = pv.statusClass
	pv.prevStatusClass = pv.statusClassHL

	if input.lwSet < _tagValue && _tagValue < input.hwSet {
		pv.statusClassHL = "Normal"
		pv.statusClass = "Normal"
	} else if _tagValue <= input.lfSet {
		pv.statusClassHL = "Low"
		pv.statusClass = "Fault"
	} else if input.lfSet < _tagValue && _tagValue <= input.lwSet {
		pv.statusClassHL = "Low"
		pv.statusClass = "Warning"
	} else if input.hwSet <= _tagValue && _tagValue < input.hfSet {
		pv.statusClassHL = "High"
		pv.statusClass = "Warning"
	} else if input.hfSet <= _tagValue {
		pv.statusClassHL = "High"
		pv.statusClass = "Fault"
	}

	if pv.statusClassHL != pv.prevStatusClassHL || pv.statusClass != pv.prevStatusClass { //이전과 상태가 다를 때 변동시점
		if pv.prevStatusClass == "Fault" {
			if pv.prevStatusClassHL == "High" {
				if pv.alarmCheckStartHF != 0 && pv.alarmCheckStartHF+int64(input.hfAlmMsec) <= _nowTime {
					pv.alarmCountHF++

					if pv.alarmCountHF >= int32(input.hfAlmTimes) {
						strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/Count/%s/%d", date, pv.prevStatusClassHL, pv.prevStatusClass, input.hfSet, _tagName, pv.alarmCountHF)
						outputAlarm(strAlarm)
					}
					pv.alarmCheckStartHW = 0
				}
				pv.alarmCheckStartHF = 0

			} else {
				if pv.alarmCheckStartLF != 0 && pv.alarmCheckStartLF+int64(input.lfAlmMsec) <= _nowTime {
					pv.alarmCountLF++

					if pv.alarmCountLF >= int32(input.lfAlmTimes) {
						strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/Count/%s/%d", date, pv.prevStatusClassHL, pv.prevStatusClass, input.lfSet, _tagName, pv.alarmCountLF)
						outputAlarm(strAlarm)
					}

					pv.alarmCheckStartLW = 0
				}
				pv.alarmCheckStartLF = 0
			}
		}

		if pv.prevStatusClass == "Warning" && pv.prevStatusClass != "Fault" {
			if pv.prevStatusClassHL == "High" {
				if pv.alarmCheckStartHW != 0 && pv.alarmCheckStartHW+int64(input.hwAlmMsec) <= _nowTime {
					pv.alarmCountHW++

					if pv.alarmCountHW >= int32(input.hwAlmTimes) {
						strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/Count/%s/%d", date, pv.prevStatusClassHL, pv.prevStatusClass, input.hwSet, _tagName, pv.alarmCountHW)
						outputAlarm(strAlarm)
					}
				}
				pv.alarmCheckStartHW = 0

			} else {
				if pv.alarmCheckStartLW != 0 && pv.alarmCheckStartLW+int64(input.lwAlmMsec) <= _nowTime {
					pv.alarmCountLW++

					if pv.alarmCountLW >= int32(input.lwAlmTimes) {
						strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/Count/%s/%d", date, pv.prevStatusClassHL, pv.prevStatusClass, input.lwSet, _tagName, pv.alarmCountLW)
						outputAlarm(strAlarm)
					}
				}
				pv.alarmCheckStartLW = 0
			}
		}
	}

	if pv.statusClass == "Normal" {
		pv.alarmCheckStartHF = 0
		pv.alarmCheckStartHW = 0
		pv.alarmCheckStartLF = 0
		pv.alarmCheckStartLW = 0
		return
	}

	if pv.statusClassHL == "High" {
		if pv.statusClass == "Fault" && pv.alarmCheckStartHF == 0 {
			pv.alarmCheckStartHF = _nowTime
		}

		if pv.alarmCheckStartHF+int64(input.limitAlarmMsec) <= _nowTime && pv.alarmCheckStartHF != 0 {
			strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/OverLoad/%s/%ds", date, pv.statusClassHL, pv.statusClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
			outputAlarm(strAlarm)
			pv.alarmCheckStartHW = 0
			pv.alarmCheckStartHF = 0
		}

		if pv.statusClass == "Warning" && pv.alarmCheckStartHW == 0 {
			pv.alarmCheckStartHW = _nowTime
		}

		if pv.alarmCheckStartHW+int64(input.limitAlarmMsec) <= _nowTime && pv.alarmCheckStartHW != 0 {
			strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/OverLoad/%s/%ds", date, pv.statusClassHL, pv.statusClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
			outputAlarm(strAlarm)
			pv.alarmCheckStartHW = 0
		}

	} else {
		if pv.statusClass == "Fault" && pv.alarmCheckStartLF == 0 {
			pv.alarmCheckStartLF = _nowTime
		}

		if pv.alarmCheckStartLF+int64(input.limitAlarmMsec) <= _nowTime && pv.alarmCheckStartLF != 0 {
			strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/OverLoad/%s/%ds", date, pv.statusClassHL, pv.statusClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
			outputAlarm(strAlarm)
			pv.alarmCheckStartLW = 0
			pv.alarmCheckStartLF = 0
		}

		if pv.statusClass == "Warning" && pv.alarmCheckStartLW == 0 {
			pv.alarmCheckStartLW = _nowTime
		}

		if pv.alarmCheckStartLW+int64(input.limitAlarmMsec) <= _nowTime && pv.alarmCheckStartLW != 0 {
			strAlarm := fmt.Sprintf("%s/Level/%s%s/%f/OverLoad/%s/%ds", date, pv.statusClassHL, pv.statusClass, _tagValue, _tagName, int64(input.limitAlarmMsec/1000))
			outputAlarm(strAlarm)
			pv.alarmCheckStartLW = 0
		}
	}

}
