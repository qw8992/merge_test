package main

type PeakInput struct {
	offset           float64
	peakPeriod       float64
	meanException    float64
	peakWarningSet   float64
	peakFaultSet     float64
	resetPeriod      float64
	limitAlarmMsec   float64
	peakFaultTimes   float64
	peakWarningTimes float64
	peakFaultMsec    float64
	peakWarningMsec  float64
	endCount         float64
}

type PeakProcessVariable struct {
	alarmClass string //경고인지 주의 인지

	peakFaultCnt     int32
	peakWarningCnt   int32
	peakPeriodStart  int64
	resetStart       int64
	waveformstart    int64
	peakEnd          int64
	peakFaultStart   int64
	peakWarningStart int64
	endCountStart    int64
	maxValue         float64
	maxValueTime     string
	maxPerValue         [4]float64
	maxPerTime         [4]int64
}

type MeanInput struct {
	offset           float64
	meanPeriod       float64
	meanException    float64
	meanPercent      float64
	meanDuration     float64
	meanWarningSet   float64
	meanFaultSet     float64
	resetPeriod      float64
	limitAlarmMsec   float64
	meanFaultTimes   float64
	meanWarningTimes float64
	meanFaultMsec    float64
	meanWarningMsec  float64
	endCount         float64
}

type MeanProcessVariable struct {
	alarmClass string //경고인지 주의 인지

	meanFaultCnt      int32
	meanWarningCnt    int32
	durationCount     int32
	durationFlag      int32
	meanPeriodStart   int64
	resetStart        int64
	waveformstart     int64
	meanDurationStart int64
	meanFaultStart    int64
	meanWarningStart  int64
	endCountStart     int64
	meanValueTime     string
	meanValue         float64
	sumValues         float64
	meanStandard      float64
	meanErr           float64
	totalSumValues    float64
	totalLength       float64
	maxPerValue         [4]float64
	maxPerTotalValue         [4]float64
	maxPerTime         [4]int64	
}

type LevelInput struct {
	limitAlarmMsec float64
	hfAlmTimes     float64
	hwAlmTimes     float64
	lfAlmTimes     float64
	lwAlmTimes     float64
	hfAlmMsec      float64
	hwAlmMsec      float64
	lfAlmMsec      float64
	lwAlmMsec      float64
	almPeriod      float64
	resetPeriod    float64
	hfSet          float64
	hwSet          float64
	lfSet          float64
	lwSet          float64
}

type LevelProcessVariable struct {
	//경고인지 주의 인지
	prevStatusClassHL string
	statusClassHL     string
	prevStatusClass   string
	statusClass       string

	almPeriodEnd      int64
	alarmCheckStart   int64
	resetStart        int64
	alarmCheckStartHF int64
	alarmCheckStartHW int64
	alarmCheckStartLF int64
	alarmCheckStartLW int64
	endcountstart     int64
	alarmCountLF      int32
	alarmCountLW      int32
	alarmCountHF      int32
	alarmCountHW      int32
	maxValue          float64
	minValue          float64
}

type SQTimeInput struct {
	offset      float64
	resetPeriod float64
	endCount    float64
}

type SQTimeProcessVariable struct {
	alarmClass string //경고인지 주의 인지

	resetStart    int64
	endCountStart int64
	waveformStart int64
	waveformEnd   int64
	cnt           int32
	sumValues     float64
	avgValues     float64
	maxPerValue         [4]float64
	maxPerTotalValue         [4]float64
	maxPerTime         [4]int64
}
