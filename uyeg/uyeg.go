package uyeg

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

func high_low_concat(high int, low int) int {
	return int(high*65536 + low)
}

func toFixed(n float64) float64 {
	return math.Round(n*100) / 100
}

func getRowValue(row map[string]interface{}, vName string) float64 {
	value := .0

	if v := row[vName]; v != nil {
		value = v.(float64)
	}
	return value
}

type Device struct {
	Id                   int
	GatewayId            string
	MacId                string
	Name                 string
	Host                 string
	Port                 int
	UnitId               uint16
	Version              uint16
	ProcessInterval      int
	RetryCycle           int
	RetryCount           int
	RetryConnFailedCount int
}

type ModbusClient struct {
	Device  Device
	Handler *modbus.TCPClientHandler
	Done1   chan bool
	Done2   chan bool
	Done3   chan bool
}

func (mb *ModbusClient) Connect() bool {
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", mb.Device.Host, mb.Device.Port))
	handler.Timeout = 2 * time.Second
	handler.SlaveId = byte(mb.Device.UnitId)
	handler.IdleTimeout = 1 * time.Second
	mb.Handler = handler

	err := handler.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (mb *ModbusClient) Close() {
	// Modbus Connect Close
	mb.Done1 <- true
	close(mb.Done1)
	mb.Done2 <- true
	close(mb.Done2)
	// mb.Done3 <- true
	// close(mb.Done3)
	mb.Handler.Close()
	fmt.Println(fmt.Sprintf("%s(%s): Close connection", mb.Device.Name, mb.Device.MacId))
}

func (mb *ModbusClient) CheckMacId() bool {
	var macId string

	client := modbus.NewClient(mb.Handler)
	var results []byte
	if mb.Device.Version < 1300 {
		results, _ = client.ReadHoldingRegisters(3012, 6)
	} else {
		results, _ = client.ReadHoldingRegisters(680, 6)
	}

	for i := 0; i < len(results); i += 2 {
		macId += strings.ToUpper(strconv.FormatInt(int64(binary.BigEndian.Uint16(results[i:i+2])), 16))
	}

	if macId != mb.Device.MacId {
		return false
	}
	return true
}

// 데이터 수집.
func (mb *ModbusClient) GetReadHoldingRegisters() map[string]interface{} {
	defer func() {
		v := recover()
		if v != nil {
			fmt.Println("GetReadHoldingRegisters:", v)
		}
	}()

	var results []byte
	var err error
	data := []int{}

	client := modbus.NewClient(mb.Handler)
	if mb.Device.Version == 1 {
		results, err = client.ReadHoldingRegisters(900, 31)
	} else if mb.Device.Version == 2 {
		results, err = client.ReadHoldingRegisters(900, 41)
	} else if mb.Device.Version == 3 {
		results, err = client.ReadHoldingRegisters(900, 27)
	} else if mb.Device.Version == 1331 {
		results, err = client.ReadHoldingRegisters(1299, 31)
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	for i := 0; i < len(results); i += 2 {
		data = append(data, int(binary.BigEndian.Uint16(results[i:i+2])))
	}
	client = nil
	return mb.GetDataToRemapData(data)
}

// 수집된 데이터를 리맵 데이터로 변환
func (mb *ModbusClient) GetDataToRemapData(dl []int) map[string]interface{} {
	var loc, _ = time.LoadLocation("Asia/Seoul")
	dmap := make(map[string]interface{})

	t := time.Now().In(loc)
	if mb.Device.Version == 1 && len(dl) == 31 {
		dmap["time"] = t
		dmap["Curr"] = toFixed(float64(high_low_concat(dl[0], dl[1])) * 0.01)              // max_current - 0
		dmap["CurrR"] = toFixed(float64(high_low_concat(dl[2], dl[3])) * 0.01)             // current_r - 1
		dmap["CurrS"] = toFixed(float64(high_low_concat(dl[4], dl[5])) * 0.01)             // current_s - 2
		dmap["CurrT"] = toFixed(float64(high_low_concat(dl[6], dl[7])) * 0.01)             // current_t - 3
		dmap["Volt"] = toFixed(float64(dl[8]) * 0.1)                                       // avg_voltage - 4
		dmap["VoltR"] = toFixed(float64(dl[9]) * 0.1)                                      // voltage_l3l1 - 5
		dmap["VoltS"] = toFixed(float64(dl[10]) * 0.1)                                     // voltage_l1l2 - 6
		dmap["VoltT"] = toFixed(float64(dl[11]) * 0.1)                                     // voltage_l2l3 - 7
		dmap["Temp"] = toFixed(float64(dl[12]) * 0.1)                                      // temperature - 8
		dmap["Humid"] = toFixed(float64(dl[13]))                                           // humidity - 9
		dmap["ActivePower"] = toFixed(float64(high_low_concat(dl[14], dl[15])))            // active_power - 10
		dmap["ActiveConsum"] = toFixed(float64(high_low_concat(dl[16], dl[17])) * 0.01)    // active_power_consumption - 11
		dmap["ReactiveConsum"] = toFixed(float64(high_low_concat(dl[18], dl[19])) * 0.01)  // reactive_power_consumption - 12
		dmap["Power"] = toFixed(float64(dl[20]) * 0.01)                                    // power_factor - 13
		dmap["TotalRunningHour"] = toFixed(float64(high_low_concat(dl[21], dl[22])) * 0.1) // total_running_hour - 14
		dmap["MCCounter"] = toFixed(float64(high_low_concat(dl[23], dl[24])))              // mc_count_display - 15
		dmap["Ground"] = toFixed(float64(high_low_concat(dl[25], dl[26])) * 0.001)         // ground_current - 16
		dmap["PT100"] = toFixed(float64(dl[27]) * 0.1)                                     // temperature of pt100 in PDM - 17
		dmap["420"] = toFixed(float64(dl[28]) * 0.01)                                      // 420 input in PDM - 18
		dmap["FaultNumber"] = toFixed(float64(dl[29]))                                     // FaultNumber - 19
		dmap["FaultRST"] = toFixed(float64(dl[30]))                                        // FaultRST - 20
	} else if mb.Device.Version == 2 && len(dl) == 41 {
		dmap["time"] = t
		dmap["Curr"] = toFixed(float64(high_low_concat(dl[0], dl[1])) * 0.01)              // max_current - 0
		dmap["CurrR"] = toFixed(float64(high_low_concat(dl[2], dl[3])) * 0.01)             // current_r - 1
		dmap["CurrS"] = toFixed(float64(high_low_concat(dl[4], dl[5])) * 0.01)             // current_s - 2
		dmap["CurrT"] = toFixed(float64(high_low_concat(dl[6], dl[7])) * 0.01)             // current_t - 3
		dmap["Volt"] = toFixed(float64(dl[8]) * 0.1)                                       // avg_voltage - 4
		dmap["VoltR"] = toFixed(float64(dl[9]) * 0.1)                                      // voltage_l3l1 - 5
		dmap["VoltS"] = toFixed(float64(dl[10]) * 0.1)                                     // voltage_l1l2 - 6
		dmap["VoltT"] = toFixed(float64(dl[11]) * 0.1)                                     // voltage_l2l3 - 7
		dmap["Temp"] = toFixed(float64(dl[12]) * 0.1)                                      // temperature - 8
		dmap["Humid"] = toFixed(float64(dl[13]))                                           // humidity - 9
		dmap["ActivePower"] = toFixed(float64(high_low_concat(dl[14], dl[15])))            // active_power - 10
		dmap["ReactivePower"] = toFixed(float64(high_low_concat(dl[16], dl[17])))          // reactive_power - 11
		dmap["ActiveConsum"] = toFixed(float64(high_low_concat(dl[18], dl[19])) * 0.01)    // active_power_consumption - 12
		dmap["ReactiveConsum"] = toFixed(float64(high_low_concat(dl[20], dl[21])) * 0.01)  // reactive_power_consumption - 13
		dmap["Power"] = toFixed(float64(dl[22]) * 0.01)                                    // power_factor - 14
		dmap["Running_hour"] = toFixed(float64(high_low_concat(dl[23], dl[24])) * 0.1)     // running_hour - 15
		dmap["TotalRunningHour"] = toFixed(float64(high_low_concat(dl[25], dl[26])) * 0.1) // total_running_hour - 16
		dmap["MCCounter"] = toFixed(float64(high_low_concat(dl[27], dl[28])))              // mc_count_display - 17
		dmap["Ground"] = toFixed(float64(high_low_concat(dl[29], dl[30])) * 0.001)         // ground_current - 18
		dmap["PT100"] = toFixed(float64(dl[31]) * 0.1)                                     // temperature of pt100 in PDM - 19
		dmap["420"] = toFixed(float64(dl[32]) * 0.01)                                      // 420 input in PDM - 20
		dmap["FaultNumber"] = toFixed(float64(dl[33]))                                     // FaultNumber - 21
		dmap["OverCurrR"] = toFixed(float64(high_low_concat(dl[34], dl[35])) * 0.01)       // OverCurrR - 22
		dmap["OverCurrS"] = toFixed(float64(high_low_concat(dl[36], dl[37])) * 0.01)       // OverCurrS - 23
		dmap["OverCurrT"] = toFixed(float64(high_low_concat(dl[38], dl[39])) * 0.01)       // OverCurrT - 24
		dmap["FaultRST"] = toFixed(float64(dl[40]))                                        // FaultRST - 25
	} else if mb.Device.Version == 3 && len(dl) == 27 {
		dmap["time"] = t
		dmap["Curr"] = toFixed(float64(high_low_concat(dl[0], dl[1])) * 0.01)              // max_current - 0
		dmap["CurrR"] = toFixed(float64(high_low_concat(dl[2], dl[3])) * 0.01)             // current_r - 1
		dmap["CurrS"] = toFixed(float64(high_low_concat(dl[4], dl[5])) * 0.01)             // current_s - 2
		dmap["CurrT"] = toFixed(float64(high_low_concat(dl[6], dl[7])) * 0.01)             // current_t - 3
		dmap["Volt"] = toFixed(float64(dl[8]) * 0.1)                                       // avg_voltage - 4
		dmap["VoltR"] = toFixed(float64(dl[9]) * 0.1)                                      // voltage_l3l1 - 5
		dmap["VoltS"] = toFixed(float64(dl[10]) * 0.1)                                     // voltage_l1l2 - 6
		dmap["VoltT"] = toFixed(float64(dl[11]) * 0.1)                                     // voltage_l2l3 - 7
		dmap["Temp"] = toFixed(float64(dl[12]) * 0.1)                                      // temperature - 8
		dmap["Humid"] = toFixed(float64(dl[13]))                                           // humidity - 9
		dmap["ActivePower"] = toFixed(float64(high_low_concat(dl[14], dl[15])))            // active_power - 10
		dmap["ActiveConsum"] = toFixed(float64(high_low_concat(dl[16], dl[17])) * 0.01)    // active_power_consumption - 11
		dmap["ReactiveConsum"] = toFixed(float64(high_low_concat(dl[18], dl[19])) * 0.01)  // reactive_power_consumption - 12
		dmap["TotalRunningHour"] = toFixed(float64(high_low_concat(dl[20], dl[21])) * 0.1) // total_running_hour - 13
		dmap["MCCounter"] = toFixed(float64(high_low_concat(dl[22], dl[23])))              // mc_count_display - 14
		dmap["Ground"] = toFixed(float64(high_low_concat(dl[24], dl[25])) * 0.001)         // ground_current - 15
		dmap["PT100"] = toFixed(float64(dl[26]) * 0.1)                                     // temperature of pt100 in PDM - 16
	} else if mb.Device.Version == 1331 && len(dl) == 31 { ////////// ****** UYeG-SE v1.1c6 ****** ////////////
		dmap["time"] = t
		dmap["Curr"] = toFixed(float64(dl[0]) * 0.01)                  // max_current - 0
		dmap["CurrR"] = toFixed(float64(dl[1]) * 0.01)                 // current_r - 1
		dmap["CurrS"] = toFixed(float64(dl[2]) * 0.01)                 // current_s - 2
		dmap["CurrT"] = toFixed(float64(dl[3]) * 0.01)                 // current_t - 3
		dmap["Volt"] = toFixed(float64(dl[4]))                         // avg_voltage - 4
		dmap["VoltR"] = toFixed(float64(dl[5]))                        // voltage_l3l1 - 5
		dmap["VoltS"] = toFixed(float64(dl[6]))                        // voltage_l1l2 - 6
		dmap["VoltT"] = toFixed(float64(dl[7]))                        // voltage_l2l3 - 7
		dmap["Temp"] = toFixed(float64(dl[8]) * 0.01)                  // Internal Temperature - 8
		dmap["Humid"] = toFixed(float64(dl[9]) * 0.01)                 // Internal Humidity - 9
		dmap["ActivePower"] = toFixed(float64(dl[10]) * 0.01)          // active_power - 10
		dmap["ReactivePower"] = toFixed(float64(dl[11]) * 0.01)        // reactive_power - 11
		dmap["ActiveConsum"] = toFixed(float64(dl[12]) * 0.01)         // P Active hours(유효 전력적산) - 12
		dmap["ReactiveConsum"] = toFixed(float64(dl[13]) * 0.01)       // P ReActive hours(무효 전력적산) - 13
		dmap["Power"] = toFixed(float64(dl[14]) * 0.01)                // P Factor(역률) - 14
		dmap["Running_hour"] = toFixed(float64(dl[15]))                // Current Motor Hour(현재 구동시간) - 15
		dmap["TotalRunningDay"] = toFixed(float64(dl[16]))             // Motor LIFE Day(구동 일수) - 16
		dmap["TotalRunningHour"] = toFixed(float64(dl[17]))            // Motor LIFE Hour(구동 시간) - 17
		dmap["MCCounter"] = toFixed(float64(dl[18]))                   // Motor Count(모터 동작 횟수) - 18
		dmap["Ground"] = toFixed(float64(dl[19]))                      // ZI (mA) - 19
		dmap["PT100"] = toFixed(float64(dl[20]) * 0.1)                 // PT100 (0.1C) - 20
		dmap["Currentloop420Input0"] = toFixed(float64(dl[21]) * 0.01) // 420 Current-loop Input 0 (0.01mA) - 21
		dmap["Currentloop420Input1"] = toFixed(float64(dl[22]) * 0.01) // 420 Current-loop Input 1 (0.01mA) - 22
		dmap["Event1Type"] = toFixed(float64(dl[23]))                  // Event1_Type(Fault) - 23
		dmap["Event1IL1Current"] = toFixed(float64(dl[24]))            // Event1 IL1 Current - 24
		dmap["Event1IL2Current"] = toFixed(float64(dl[25]))            // Event1 IL2 Current - 25
		dmap["Event1IL3Current"] = toFixed(float64(dl[26]))            // Event1 IL3 Current - 26
		dmap["PLPhaseI"] = toFixed(float64(dl[27]))                    // PL Phase I - 27
		dmap["PLPhaseV"] = toFixed(float64(dl[28]))                    // PL Phase V - 28
		dmap["LogicIN1"] = toFixed(float64(dl[29]))                    // Logic IN1 - 29
		dmap["LogicIN2"] = toFixed(float64(dl[30]))                    // Logic IN2 - 30
	} else {
		dmap = nil
	}
	return dmap
}

func (mb *ModbusClient) GetRemapJson(bt string, data []interface{}) []byte {
	bytes := []byte{}
	if mb.Device.Version == 1 {
		rpFormat := &RemapFormatV1{}
		rpFormat.Values = []Depth2V1{}
		for _, row := range data {
			r := row.(map[string]interface{})
			rpFormat.Time = bt
			rpFormat.Version = mb.Device.Version
			rpFormat.GatewayID = mb.Device.GatewayId
			rpFormat.MacID = mb.Device.MacId
			rpFormat.Temp = math.Max(rpFormat.Temp, getRowValue(r, "Temp"))
			rpFormat.Humid = math.Max(rpFormat.Humid, getRowValue(r, "Humid"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactivePower"))
			rpFormat.ActiveConsum = math.Max(rpFormat.ActiveConsum, getRowValue(r, "ActiveConsum"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactiveConsum"))
			rpFormat.Power = math.Max(rpFormat.Power, getRowValue(r, "Power"))
			rpFormat.TotalRunningHour = math.Max(rpFormat.TotalRunningHour, getRowValue(r, "TotalRunningHour"))
			rpFormat.MCCounter = math.Max(rpFormat.MCCounter, getRowValue(r, "MCCounter"))
			rpFormat.PT100 = math.Max(rpFormat.PT100, getRowValue(r, "PT100"))
			rpFormat.FaultNumber = math.Max(rpFormat.FaultNumber, getRowValue(r, "FaultNumber"))
			rpFormat.FaultRST = math.Max(rpFormat.FaultRST, getRowValue(r, "FaultRST"))

			value := Depth2V1{}
			value.Time = r["time"].(string)
			value.Status = r["status"].(bool)
			value.Curr = getRowValue(r, "Curr")
			value.CurrR = getRowValue(r, "CurrR")
			value.CurrS = getRowValue(r, "CurrS")
			value.CurrT = getRowValue(r, "CurrT")
			value.Volt = getRowValue(r, "Volt")
			value.VoltR = getRowValue(r, "VoltR")
			value.VoltS = getRowValue(r, "VoltS")
			value.VoltT = getRowValue(r, "VoltT")
			value.ActivePower = getRowValue(r, "ActivePower")
			value.Ground = getRowValue(r, "Ground")
			value.V420 = getRowValue(r, "V420")

			rpFormat.Values = append(rpFormat.Values, value)
		}
		jsonBytes, _ := json.Marshal(rpFormat)
		bytes = jsonBytes
	} else if mb.Device.Version == 2 {
		rpFormat := &RemapFormatV2{}
		rpFormat.Values = []Depth2V1{}
		for _, row := range data {
			r := row.(map[string]interface{})
			rpFormat.Time = bt
			rpFormat.Version = mb.Device.Version
			rpFormat.GatewayID = mb.Device.GatewayId
			rpFormat.MacID = mb.Device.MacId
			rpFormat.Temp = math.Max(rpFormat.Temp, getRowValue(r, "Temp"))
			rpFormat.Humid = math.Max(rpFormat.Humid, getRowValue(r, "Humid"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactivePower"))
			rpFormat.ActiveConsum = math.Max(rpFormat.ActiveConsum, getRowValue(r, "ActiveConsum"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactiveConsum"))
			rpFormat.Power = math.Max(rpFormat.Power, getRowValue(r, "Power"))
			rpFormat.TotalRunningHour = math.Max(rpFormat.TotalRunningHour, getRowValue(r, "TotalRunningHour"))
			rpFormat.MCCounter = math.Max(rpFormat.MCCounter, getRowValue(r, "MCCounter"))
			rpFormat.PT100 = math.Max(rpFormat.PT100, getRowValue(r, "PT100"))
			rpFormat.FaultNumber = math.Max(rpFormat.FaultNumber, getRowValue(r, "FaultNumber"))
			rpFormat.FaultRST = math.Max(rpFormat.FaultRST, getRowValue(r, "FaultRST"))

			value := Depth2V1{}
			value.Time = r["time"].(string)
			value.Status = r["status"].(bool)
			value.Curr = getRowValue(r, "Curr")
			value.CurrR = getRowValue(r, "CurrR")
			value.CurrS = getRowValue(r, "CurrS")
			value.CurrT = getRowValue(r, "CurrT")
			value.Volt = getRowValue(r, "Volt")
			value.VoltR = getRowValue(r, "VoltR")
			value.VoltS = getRowValue(r, "VoltS")
			value.VoltT = getRowValue(r, "VoltT")
			value.ActivePower = getRowValue(r, "ActivePower")
			value.Ground = getRowValue(r, "Ground")
			value.V420 = getRowValue(r, "V420")

			rpFormat.Values = append(rpFormat.Values, value)
		}
		jsonBytes, _ := json.Marshal(rpFormat)
		bytes = jsonBytes
	} else if mb.Device.Version == 3 {
		rpFormat := &RemapFormatV2{}
		rpFormat.Values = []Depth2V1{}
		for _, row := range data {
			r := row.(map[string]interface{})
			rpFormat.Time = bt
			rpFormat.Version = mb.Device.Version
			rpFormat.GatewayID = mb.Device.GatewayId
			rpFormat.MacID = mb.Device.MacId
			rpFormat.Temp = math.Max(rpFormat.Temp, getRowValue(r, "Temp"))
			rpFormat.Humid = math.Max(rpFormat.Humid, getRowValue(r, "Humid"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactivePower"))
			rpFormat.ActiveConsum = math.Max(rpFormat.ActiveConsum, getRowValue(r, "ActiveConsum"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactiveConsum"))
			rpFormat.Power = math.Max(rpFormat.Power, getRowValue(r, "Power"))
			rpFormat.TotalRunningHour = math.Max(rpFormat.TotalRunningHour, getRowValue(r, "TotalRunningHour"))
			rpFormat.MCCounter = math.Max(rpFormat.MCCounter, getRowValue(r, "MCCounter"))
			rpFormat.PT100 = math.Max(rpFormat.PT100, getRowValue(r, "PT100"))
			rpFormat.FaultNumber = math.Max(rpFormat.FaultNumber, getRowValue(r, "FaultNumber"))
			rpFormat.FaultRST = math.Max(rpFormat.FaultRST, getRowValue(r, "FaultRST"))

			value := Depth2V1{}
			value.Time = r["time"].(string)
			value.Status = r["status"].(bool)
			value.Curr = getRowValue(r, "Curr")
			value.CurrR = getRowValue(r, "CurrR")
			value.CurrS = getRowValue(r, "CurrS")
			value.CurrT = getRowValue(r, "CurrT")
			value.Volt = getRowValue(r, "Volt")
			value.VoltR = getRowValue(r, "VoltR")
			value.VoltS = getRowValue(r, "VoltS")
			value.VoltT = getRowValue(r, "VoltT")
			value.ActivePower = getRowValue(r, "ActivePower")
			value.Ground = getRowValue(r, "Ground")
			value.V420 = getRowValue(r, "V420")

			rpFormat.Values = append(rpFormat.Values, value)
		}
		jsonBytes, _ := json.Marshal(rpFormat)
		bytes = jsonBytes
	} else if mb.Device.Version == 1331 {
		rpFormat := &RemapFormatV1331{}
		rpFormat.Values = []Depth2V2{}
		for _, row := range data {
			r := row.(map[string]interface{})
			rpFormat.Time = bt
			rpFormat.Version = mb.Device.Version
			rpFormat.GatewayID = mb.Device.GatewayId
			rpFormat.MacID = mb.Device.MacId
			rpFormat.Temp = math.Max(rpFormat.Temp, getRowValue(r, "Temp"))
			rpFormat.Humid = math.Max(rpFormat.Humid, getRowValue(r, "Humid"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactivePower"))
			rpFormat.ActiveConsum = math.Max(rpFormat.ActiveConsum, getRowValue(r, "ActiveConsum"))
			rpFormat.ReactiveConsum = math.Max(rpFormat.ReactiveConsum, getRowValue(r, "ReactiveConsum"))
			rpFormat.Power = math.Max(rpFormat.Power, getRowValue(r, "Power"))
			rpFormat.TotalRunningDay = math.Max(rpFormat.TotalRunningDay, getRowValue(r, "TotalRunningDay"))
			rpFormat.TotalRunningHour = math.Max(rpFormat.TotalRunningHour, getRowValue(r, "TotalRunningHour"))
			rpFormat.MCCounter = math.Max(rpFormat.MCCounter, getRowValue(r, "MCCounter"))
			rpFormat.PT100 = math.Max(rpFormat.PT100, getRowValue(r, "PT100"))
			rpFormat.Event1Type = math.Max(rpFormat.Event1Type, getRowValue(r, "Event1Type"))
			rpFormat.Event1IL1Current = math.Max(rpFormat.Event1IL1Current, getRowValue(r, "Event1IL1Current"))
			rpFormat.Event1IL2Current = math.Max(rpFormat.Event1IL2Current, getRowValue(r, "Event1IL2Current"))
			rpFormat.Event1IL3Current = math.Max(rpFormat.Event1IL3Current, getRowValue(r, "Event1IL3Current"))
			rpFormat.PLPhaseI = math.Max(rpFormat.PLPhaseI, getRowValue(r, "PLPhaseI"))
			rpFormat.PLPhaseV = math.Max(rpFormat.PLPhaseV, getRowValue(r, "PLPhaseV"))
			rpFormat.LogicIN1 = math.Max(rpFormat.LogicIN1, getRowValue(r, "LogicIN1"))
			rpFormat.LogicIN2 = math.Max(rpFormat.LogicIN2, getRowValue(r, "LogicIN2"))

			value := Depth2V2{}
			value.Time = r["time"].(string)
			value.Status = r["status"].(bool)
			value.Curr = getRowValue(r, "Curr")
			value.CurrR = getRowValue(r, "CurrR")
			value.CurrS = getRowValue(r, "CurrS")
			value.CurrT = getRowValue(r, "CurrT")
			value.Volt = getRowValue(r, "Volt")
			value.VoltR = getRowValue(r, "VoltR")
			value.VoltS = getRowValue(r, "VoltS")
			value.VoltT = getRowValue(r, "VoltT")
			value.ActivePower = getRowValue(r, "ActivePower")
			value.Ground = getRowValue(r, "Ground")
			value.V420Input0 = getRowValue(r, "V420Input0")
			value.V420Input1 = getRowValue(r, "V420Input1")

			rpFormat.Values = append(rpFormat.Values, value)
		}
		jsonBytes, _ := json.Marshal(rpFormat)
		bytes = jsonBytes
	}
	return bytes
}
