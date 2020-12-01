package main

import (
	"fmt"
	"strings"
)

func outputExtraction(info string) {
	var sqlStr string
	s1 := strings.Split(info, "/")
	// fmt.Println(info)
	if s1[0] == "Avg" {
		// if strings.Contains(s1[4],"2C6A6FB0287F.VoltR"){
		// 	fmt.Println(info)
		// }
		s2 := strings.Split(s1[4], ".")
		sqlStr = fmt.Sprintf("Insert ignore Into His%s%s (DefServer, DefTable, DefColumn, DataSavedTime, Value, Area) ", s1[0], s1[5])
		sqlStr = fmt.Sprintf("%s \nSELECT DefServer, DefTable, DefColumn, '%s', ROUND(%s,3), ROUND(%s,3) from MstDevice ", sqlStr, s1[1], s1[2], s1[3])
		sqlStr = fmt.Sprintf("%s where mac = '%s' and DefTable = 'HisItem%s'; ", sqlStr, s2[0], s2[1])

	} else if s1[0] == "Mean" {
		s2 := strings.Split(s1[4], ".")
		sqlStr = fmt.Sprintf("Insert ignore Into His%s%s (DefServer, DefTable, DefColumn, DataSavedTime, Value, totalAVG) ", s1[0], s1[5])
		sqlStr = fmt.Sprintf("%s \nSELECT DefServer, DefTable, DefColumn, '%s', ROUND(%s,3), ROUND(%s,3) from MstDevice ", sqlStr, s1[1], s1[2], s1[3])
		sqlStr = fmt.Sprintf("%s where mac = '%s' and DefTable = 'HisItem%s'; ", sqlStr, s2[0], s2[1])

	} else {
		s2 := strings.Split(s1[3], ".")
		sqlStr = fmt.Sprintf("Insert ignore Into His%s%s (DefServer, DefTable, DefColumn, DataSavedTime, Value) ", s1[0], s1[4])
		sqlStr = fmt.Sprintf("%s \nSELECT DefServer, DefTable, DefColumn, '%s', ROUND(%s,3) from MstDevice ", sqlStr, s1[1], s1[2])
		sqlStr = fmt.Sprintf("%s where mac = '%s' and DefTable = 'HisItem%s'; ", sqlStr, s2[0], s2[1])
	}
	// fmt.Println(sqlStr)
	go dbConn.AlgorithmQueryExec(sqlStr)
}

func outputAlarm(alarmClass string) {
	s1 := strings.Split(alarmClass, "/")
	s2 := strings.Split(s1[5], ".")
	sqlStr := fmt.Sprintf("Insert ignore Into HisAlarm")
	sqlStr = fmt.Sprintf("%s (Monitored_Time, Level1, Level2, Level3, Level4, Level5, SensorItem, Mac, DefServer, DefTable, DefColumn", sqlStr)
	sqlStr = fmt.Sprintf("%s, AlarmItem, Class, Moment, Ack, Status, Value, AlarmCondition)", sqlStr)
	sqlStr = fmt.Sprintf("%s \nSELECT '%s', Level1, Level2, Level3, Level4, Level5, SensorItem, Mac, DefServer, DefTable, DefColumn", sqlStr, s1[0])
	sqlStr = fmt.Sprintf("%s, '%s', '%s', 'Generated', 'Un Ack' ,'Un Ack', '%s', '%s%s'", sqlStr, s1[1], s1[2], s1[3], s1[6], s1[4])
	sqlStr = fmt.Sprintf("%s \nfrom MstDevice \nwhere Mac = '%s' and DefTable = 'HisItem%s'", sqlStr, s2[0], s2[1])

	if strings.Contains(s1[2], "Warning") {
		faultItem := strings.Replace(s1[2], "Warning", "Fault", 1)
		sqlStr = fmt.Sprintf("%s AND NOT EXISTS ( SELECT * from HisAlarm where Mac = '%s' and DefTable = 'HisItem%s'", sqlStr, s2[0], s2[1])
		sqlStr = fmt.Sprintf("%s and AlarmItem = '%s' and (Class = '%s' or Class = '%s')", sqlStr, s1[1], s1[2], faultItem)
		sqlStr = fmt.Sprintf("%s and Released_Time is NULL order by Monitored_Time DESC LIMIT 1);", sqlStr)

	} else {
		sqlStr = fmt.Sprintf("%s AND NOT EXISTS ( SELECT * from HisAlarm where Mac = '%s' and DefTable = 'HisItem%s'", sqlStr, s2[0], s2[1])
		sqlStr = fmt.Sprintf("%s and AlarmItem = '%s' and Class = '%s'", sqlStr, s1[1], s1[2])
		sqlStr = fmt.Sprintf("%s and Released_Time is NULL order by Monitored_Time DESC LIMIT 1);", sqlStr)
	}

	// go dbConn.ResultQueryExec(sqlStr)
}
