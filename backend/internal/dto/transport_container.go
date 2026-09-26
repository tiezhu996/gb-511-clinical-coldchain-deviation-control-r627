package dto

import "time"

// CreateTransportContainer is the public write contract for 运输容器. Status is deliberately
// omitted so callers cannot bypass the service state machine.
type CreateTransportContainer struct {
	Code              string    `json:"code" binding:"required,min=2,max=64"`
	Name              string    `json:"name" binding:"required,min=2,max=160"`
	Description       string    `json:"description" binding:"max=1000"`
	Facility          string    `json:"facility" binding:"required,max=120"`
	Owner             string    `json:"owner" binding:"required,max=120"`
	Category          string    `json:"category" binding:"required,max=80"`
	RiskLevel         string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue       float64   `json:"metricValue"`
	MetricUnit        string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt       time.Time `json:"effectiveAt" binding:"required"`
	Evidence          string    `json:"evidence" binding:"max=2000"`
	RelatedCode       string    `json:"relatedCode" binding:"max=64"`
	SensorID          string    `json:"sensorId" binding:"max=80"`
	ContainerType     string    `json:"containerType" binding:"max=80"`
	CurrentLocation   string    `json:"currentLocation" binding:"max=120"`
	Custodian         string    `json:"custodian" binding:"max=120"`
	CurrentTempC      float64   `json:"currentTempC"`
	LastSensorReading time.Time `json:"lastSensorReading"`
}

type UpdateTransportContainer struct {
	ExpectedVersion   uint      `json:"expectedVersion" binding:"required"`
	Name              string    `json:"name" binding:"required,min=2,max=160"`
	Description       string    `json:"description" binding:"max=1000"`
	Facility          string    `json:"facility" binding:"required,max=120"`
	Owner             string    `json:"owner" binding:"required,max=120"`
	Category          string    `json:"category" binding:"required,max=80"`
	RiskLevel         string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue       float64   `json:"metricValue"`
	MetricUnit        string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt       time.Time `json:"effectiveAt" binding:"required"`
	Evidence          string    `json:"evidence" binding:"max=2000"`
	RelatedCode       string    `json:"relatedCode" binding:"max=64"`
	SensorID          string    `json:"sensorId" binding:"max=80"`
	ContainerType     string    `json:"containerType" binding:"max=80"`
	CurrentLocation   string    `json:"currentLocation" binding:"max=120"`
	Custodian         string    `json:"custodian" binding:"max=120"`
	CurrentTempC      float64   `json:"currentTempC"`
	LastSensorReading time.Time `json:"lastSensorReading"`
}
