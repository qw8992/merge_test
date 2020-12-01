package uyeg

type RemapFormatV1 struct {
	Version          uint16     `json:"ver"`
	GatewayID        string     `json:"gateway"`
	MacID            string     `json:"mac"`
	Time             string     `json:"time"`
	Temp             float64    `json:"Temp"`
	Humid            float64    `json:"Humid"`
	ActiveConsum     float64    `json:"ActiveConsum"`
	ReactiveConsum   float64    `json:"ReactiveConsum"`
	Power            float64    `json:"Power"`
	TotalRunningHour float64    `json:"TotalRunningHour"`
	MCCounter        float64    `json:"MCCounter"`
	PT100            float64    `json:"PT100"`
	FaultNumber      float64    `json:"FaultNumber"`
	FaultRST         float64    `json:"FaultRST"`
	Values           []Depth2V1 `json:"Values"`
}

type RemapFormatV2 struct {
	Version          uint16     `json:"ver"`
	GatewayID        string     `json:"gateway"`
	MacID            string     `json:"mac"`
	Time             string     `json:"time"`
	Temp             float64    `json:"Temp"`
	Humid            float64    `json:"Humid"`
	ReactivePower    float64    `json:"ReactivePower"`
	ActiveConsum     float64    `json:"ActiveConsum"`
	ReactiveConsum   float64    `json:"ReactiveConsum"`
	Power            float64    `json:"Power"`
	RunningHour      float64    `json:"RunningHour"`
	TotalRunningHour float64    `json:"TotalRunningHour"`
	MCCounter        float64    `json:"MCCounter"`
	PT100            float64    `json:"PT100"`
	FaultNumber      float64    `json:"FaultNumber"`
	OverCurrR        float64    `json:"OverCurrR"`
	OverCurrS        float64    `json:"OverCurrS"`
	OverCurrT        float64    `json:"OverCurrT"`
	FaultRST         float64    `json:"FaultRST"`
	Values           []Depth2V1 `json:"Values"`
}

type RemapFormatV1331 struct {
	Version              uint16     `json:"ver"`
	GatewayID            string     `json:"gateway"`
	MacID                string     `json:"mac"`
	Time                 string     `json:"time"`
	Temp                 float64    `json:"Temp"`
	Humid                float64    `json:"Humid"`
	ReactivePower        float64    `json:"ReactivePower"`
	ActiveConsum         float64    `json:"ActiveConsum"`
	ReactiveConsum       float64    `json:"ReactiveConsum"`
	Power                float64    `json:"Power"`
	RunningHour          float64    `json:"RunningHour"`
	TotalRunningDay      float64    `json:"TotalRunningDay"`
	TotalRunningHour     float64    `json:"TotalRunningHour"`
	MCCounter            float64    `json:"MCCounter"`
	PT100                float64    `json:"PT100"`
	Currentloop420Input0 float64    `json:"Currentloop420Input0"`
	Currentloop420Input1 float64    `json:"Currentloop420Input1"`
	Event1Type           float64    `json:"Event1Type"`
	Event1IL1Current     float64    `json:"Event1IL1Current"`
	Event1IL2Current     float64    `json:"Event1IL2Current"`
	Event1IL3Current     float64    `json:"Event1IL3Current"`
	PLPhaseI             float64    `json:"PLPhaseI"`
	PLPhaseV             float64    `json:"PLPhaseV"`
	LogicIN1             float64    `json:"LogicIN1"`
	LogicIN2             float64    `json:"LogicIN2"`
	Values               []Depth2V2 `json:"Values"`
}

type Depth2V1 struct {
	Time        string  `json:"time"`
	Status      bool    `json:"status"`
	Curr        float64 `json:"Curr"`
	CurrR       float64 `json:"CurrR"`
	CurrS       float64 `json:"CurrS"`
	CurrT       float64 `json:"CurrT"`
	Volt        float64 `json:"Volt"`
	VoltR       float64 `json:"VoltR"`
	VoltS       float64 `json:"VoltS"`
	VoltT       float64 `json:"VoltT"`
	ActivePower float64 `json:"ActivePower"`
	Ground      float64 `json:"Ground"`
	V420        float64 `json:"420"`
}

type Depth2V2 struct {
	Time        string  `json:"time"`
	Status      bool    `json:"status"`
	Curr        float64 `json:"Curr"`
	CurrR       float64 `json:"CurrR"`
	CurrS       float64 `json:"CurrS"`
	CurrT       float64 `json:"CurrT"`
	Volt        float64 `json:"Volt"`
	VoltR       float64 `json:"VoltR"`
	VoltS       float64 `json:"VoltS"`
	VoltT       float64 `json:"VoltT"`
	ActivePower float64 `json:"ActivePower"`
	Ground      float64 `json:"Ground"`
	V420Input0  float64 `json:"Currentloop420Input0"`
	V420Input1  float64 `json:"Currentloop420Input1"`
}
