package model

import "time"

// TransportContainer models 运输容器 as an independently versioned aggregate. The fields
// cover ownership, operational context, evidence and measured risk so later
// changes naturally span persistence, service and UI layers.
type TransportContainer struct {
	BaseModel
	SensorID          string    `json:"sensorId" gorm:"size:80;index"`
	ContainerType     string    `json:"containerType" gorm:"size:80;index"`
	CurrentLocation   string    `json:"currentLocation" gorm:"size:120;index"`
	Custodian         string    `json:"custodian" gorm:"size:120"`
	CurrentTempC      float64   `json:"currentTempC"`
	LastSensorReading time.Time `json:"lastSensorReading"`
	Facility          string    `json:"facility" gorm:"size:120;index"`
	Owner             string    `json:"owner" gorm:"size:120;index"`
	Category          string    `json:"category" gorm:"size:80;index"`
	RiskLevel         string    `json:"riskLevel" gorm:"size:32;index"`
	MetricValue       float64   `json:"metricValue"`
	MetricUnit        string    `json:"metricUnit" gorm:"size:24"`
	EffectiveAt       time.Time `json:"effectiveAt"`
	Evidence          string    `json:"evidence" gorm:"size:2000"`
	RelatedCode       string    `json:"relatedCode" gorm:"size:64;index"`
}

func (item *TransportContainer) GetBase() *BaseModel { return &item.BaseModel }

func (item TransportContainer) TableName() string { return "transport_containers" }

var TransportContainerInitialStatus = "ready"
