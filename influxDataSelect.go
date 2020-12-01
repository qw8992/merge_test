package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	client "github.com/Heo-youngseo/influxdb1-client/v2"
)

type RemapFormatV2 struct {
	Version          string     `json:"ver"`
	GatewayID        string     `json:"gateway"`
	MacID            string     `json:"mac"`
	Time             string     `json:"time"`
	Temp             string     `json:"Temp"`
	Humid            string     `json:"Humid"`
	ReactivePower    string     `json:"ReactivePower"`
	ActiveConsum     string     `json:"ActiveConsum"`
	ReactiveConsum   string     `json:"ReactiveConsum"`
	Power            string     `json:"Power"`
	RunningHour      string     `json:"RunningHour"`
	TotalRunningHour string     `json:"TotalRunningHour"`
	MCCounter        string     `json:"MCCounter"`
	PT100            string     `json:"PT100"`
	FaultNumber      string     `json:"FaultNumber"`
	OverCurrR        string     `json:"OverCurrR"`
	OverCurrS        string     `json:"OverCurrS"`
	OverCurrT        string     `json:"OverCurrT"`
	FaultRST         string     `json:"FaultRST"`
	Values           []Depth2V1 `json:"Values"`
}

type Depth2V1 struct {
	Time        string `json:"time"`
	Status      string `json:"status"`
	Curr        string `json:"Curr"`
	CurrR       string `json:"CurrR"`
	CurrS       string `json:"CurrS"`
	CurrT       string `json:"CurrT"`
	Volt        string `json:"Volt"`
	VoltR       string `json:"VoltR"`
	VoltS       string `json:"VoltS"`
	VoltT       string `json:"VoltT"`
	ActivePower string `json:"ActivePower"`
	Ground      string `json:"Ground"`
	V420        string `json:"420"`
}

func influxDataSelect(chMsg chan map[string]interface{}) {
	c := influxDBClient()
	tick := time.Tick(1 * time.Second)
	// device := CountDevice()
	for {
		select {
		case <-tick:

			resultmac, err := selectMac(c)
			if err != nil {
				panic(err)
			}
			macmap := make(map[string]interface{})
			if len(resultmac) == 0 {
				fmt.Println("Check influxDB is running")
			} else {
				m, err := json.Marshal(resultmac[0])
				if err != nil {
					panic(err)
				}
				json.Unmarshal(m, &macmap)
				jsonmac := macmap["Series"]
				macdata := jsonmac.([]interface{})[0].(map[string]interface{})["values"]

				switch reflect.TypeOf(macdata).Kind() {
				case reflect.Slice:
					s := reflect.ValueOf(macdata)

					for i := 0; i < s.Len(); i++ {

						mac := macdata.([]interface{})[i].([]interface{})[1]
						res, err := selectQuery(c, mac)
						if err != nil {
							log.Fatal(err)
						}

						b, _ := json.Marshal(res[0])
						jsonmap := make(map[string]interface{})
						json.Unmarshal(b, &jsonmap)
						jsonvalue := jsonmap["Series"]
						Check := fmt.Sprint(jsonvalue)

						var jsondata interface{}
						if Check == "<nil>" {
							resp, _ := InitselectQuery(c, mac)
							if len(resp) == 0 {
								fmt.Println("influxDB select Error")
							} else {
								c, _ := json.Marshal(resp[0])

								jsonmap := make(map[string]interface{})
								json.Unmarshal(c, &jsonmap)
								jsonvalue := jsonmap["Series"]
								jsondata = jsonvalue.([]interface{})[0].(map[string]interface{})["values"]
							}
						} else {
							jsondata = jsonvalue.([]interface{})[0].(map[string]interface{})["values"]

							dataSecond := processingData(jsondata)
							chMsg <- dataSecond
						}

					}
					// saveTime := fmt.Sprint(time.Now())
					// r := fmt.Sprintf("%s %s(%d) UYeG-SA", saveTime[:19], gatewayID, device)
					// fmt.Println(r)
				}
			}
		}
	}
}

func selectMac(c client.Client) (res []client.Result, err error) {

	// var InfluxTable = "SmartEOCR"
	query := fmt.Sprintf("SHOW TAG VALUES FROM \"SmartEOCR\" WITH KEY = \"mac\" where gateway='%s'", gatewayID)
	q := client.Query{
		Command:  query,
		Database: database,
	}

	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results

	} else {
		return res, err
	}

	return res, nil
}

func selectQuery(c client.Client, mac interface{}) (res []client.Result, err error) {
	tNow := time.Now()
	before := tNow.Add(-2 * time.Second).Format("2006-01-02 15:04:05")
	after := tNow.Add(-1 * time.Second).Format("2006-01-02 15:04:05")
	// var InfluxTable = "SmartEOCR"
	query := fmt.Sprintf("select time, \"420\", ActiveConsum, ActivePower, Curr, CurrR, CurrS, CurrT, FaultNumber, FaultRST, Ground, Humid, MCCounter, OverCurrR, OverCurrS, OverCurrT, PT100, Power, ReactiveConsum, ReactivePower, RunningHour, Temp, TotalRunningHour, Volt, VoltR, VoltS, VoltT, gateway, mac, status from SmartEOCR where mac='%s' and time<'%s' and time>='%s' order by time desc limit 10", mac, after, before)
	q := client.Query{
		Command:  query,
		Database: database,
	}

	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results

	} else {
		return res, err
	}
	return res, nil
}

func InitselectQuery(c client.Client, mac interface{}) (res []client.Result, err error) {
	// var InfluxTable = "SmartEOCR"
	query := fmt.Sprintf("select time, \"420\", ActiveConsum, ActivePower, Curr, CurrR, CurrS, CurrT, FaultNumber, FaultRST, Ground, Humid, MCCounter, OverCurrR, OverCurrS, OverCurrT, PT100, Power, ReactiveConsum, ReactivePower, RunningHour, Temp, TotalRunningHour, Volt, VoltR, VoltS, VoltT, gateway, mac, status from SmartEOCR where mac='%s' order by time desc limit 10", mac)
	q := client.Query{
		Command:  query,
		Database: database,
	}

	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results

	} else {
		return res, err
	}
	return res, nil
}

func processingData(jsondata interface{}) (dataSecond map[string]interface{}) {
	rpFormat := &RemapFormatV2{}
	rpFormat.Values = []Depth2V1{}

	for i := 9; i >= 0; i-- {
		value := Depth2V1{}

		Time := ""
		Curr := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[4])
		CurrR := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[5])
		CurrS := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[6])
		CurrT := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[7])
		Volt := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[23])
		VoltR := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[24])
		VoltS := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[25])
		VoltT := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[26])
		ActivePower := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[3])
		Ground := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[10])
		V420 := fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[1])

		DateTime := strings.Replace(fmt.Sprint(jsondata.([]interface{})[i].([]interface{})[0]), "T", " ", -1)

		if len(DateTime) == 20 {
			Time = strings.Replace(DateTime, "Z", ".000", -1)
		} else {
			Time = strings.Replace(DateTime, "Z", "00", -1)
		}

		value.Time = nilCheck(Time)
		value.Curr = nilCheck(Curr)
		value.CurrR = nilCheck(CurrR)
		value.CurrS = nilCheck(CurrS)
		value.CurrT = nilCheck(CurrT)
		value.Volt = nilCheck(Volt)
		value.VoltR = nilCheck(VoltR)
		value.VoltS = nilCheck(VoltS)
		value.VoltT = nilCheck(VoltT)
		value.ActivePower = nilCheck(ActivePower)
		value.Ground = nilCheck(Ground)
		value.V420 = nilCheck(V420)
		value.Status = "true"
		rpFormat.Values = append(rpFormat.Values, value)
	}

	Temp := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[21])
	Humid := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[11])
	ActiveConsum := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[2])
	RunningHour := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[20])
	OverCurrR := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[13])
	OverCurrS := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[14])
	OverCurrT := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[15])
	Power := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[17])
	ReactiveConsum := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[18])
	ReactivePower := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[19])
	TotalRunningHour := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[22])
	MCCounter := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[12])
	PT100 := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[16])
	FaultNumber := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[8])
	FaultRST := fmt.Sprint(jsondata.([]interface{})[0].([]interface{})[9])

	rpFormat.Temp = nilCheck(Temp)
	rpFormat.Humid = nilCheck(Humid)
	rpFormat.ActiveConsum = nilCheck(ActiveConsum)
	rpFormat.RunningHour = nilCheck(RunningHour)
	rpFormat.OverCurrR = nilCheck(OverCurrR)
	rpFormat.OverCurrS = nilCheck(OverCurrS)
	rpFormat.OverCurrT = nilCheck(OverCurrT)
	rpFormat.Power = nilCheck(Power)
	rpFormat.ReactiveConsum = nilCheck(ReactiveConsum)
	rpFormat.ReactivePower = nilCheck(ReactivePower)
	rpFormat.TotalRunningHour = nilCheck(TotalRunningHour)
	rpFormat.MCCounter = nilCheck(MCCounter)
	rpFormat.PT100 = nilCheck(PT100)
	rpFormat.FaultNumber = nilCheck(FaultNumber)
	rpFormat.FaultRST = nilCheck(FaultRST)
	Time := strings.Replace(fmt.Sprint(jsondata.([]interface{})[9].([]interface{})[0]), "T", " ", -1)
	rpFormat.Time = strings.Replace(Time, "Z", "", -1)
	rpFormat.Version = "2"
	rpFormat.GatewayID = gatewayID
	rpFormat.MacID = fmt.Sprint(jsondata.([]interface{})[9].([]interface{})[28])

	jsonBytes, _ := json.Marshal(rpFormat)
	json.Unmarshal(jsonBytes, &dataSecond)

	return dataSecond
}

func nilCheck(data string) (res string) {
	if data == "nil" || data == "" {
		data = "0"
	}
	return data
}

func CountDevice() (device int) {
	db, err := sql.Open("mysql", "root:its@1234@tcp(127.0.0.1:3306)/UYeG_BM4")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 하나의 Row를 갖는 SQL 쿼리
	DeviceQuery := fmt.Sprintf("SELECT Count(Enabled) FROM gateway where Enabled=1")
	err = db.QueryRow(DeviceQuery).Scan(&device)
	if err != nil {
		// log.Fatal(err)
	}
	return device
}
