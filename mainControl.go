package main

import (
	"strings"
)

func algorithmRun(data map[string]interface{}) {
	if strings.Contains(data["time"].(string), "UTC") {
		return
	}

	date := data["time"].(string)
	_tagName := tagName
	for i := 0; i < len(_tagName); i++ {
		if value, ok := data[_tagName[i]]; ok {
			realData, err := getFloat(value)
			if err != nil {
				realData = float64(0)
			}

			if mapMstDevice[_tagName[i]] != nil {
				if mapMstDevice[_tagName[i]].(map[string]interface{})["Item"].(int8)%2 == 0 {
					peak[_tagName[i]].peakExtraction(date, realData, _tagName[i])
				}
				if mapMstDevice[_tagName[i]].(map[string]interface{})["Item"].(int8)%3 == 0 {
					mean[_tagName[i]].meanExtraction(date, realData, _tagName[i])
				}
				if mapMstDevice[_tagName[i]].(map[string]interface{})["Item"].(int8)%5 == 0 {
					level[_tagName[i]].levelExtraction(date, realData, _tagName[i])
				}
				if mapMstDevice[_tagName[i]].(map[string]interface{})["Item"].(int8)%7 == 0 {
					sqtime[_tagName[i]].sqtimeExtraction(date, realData, _tagName[i])
				}
			}
		} 		
	}
}
