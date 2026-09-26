package model

import "time"

// TemperatureWindow models 温控规则 as an independently versioned aggregate. The fields
// cover ownership, operational context, evidence and measured risk so later
// changes naturally span persistence, service and UI layers.
type TemperatureWindow struct {
	BaseModel
	ProductClass        string    `json:"productClass" gorm:"size:100;index"`
	MinimumCelsius      float64   `json:"minimumCelsius"`
	MaximumCelsius      float64   `json:"maximumCelsius"`
	MaxExcursionMinutes int       `json:"maxExcursionMinutes"`
	QualityOwner        string    `json:"qualityOwner" gorm:"size:120"`
	Facility            string    `json:"facility" gorm:"size:120;index"`
	Owner               string    `json:"owner" gorm:"size:120;index"`
	Category            string    `json:"category" gorm:"size:80;index"`
	RiskLevel           string    `json:"riskLevel" gorm:"size:32;index"`
	MetricValue         float64   `json:"metricValue"`
	MetricUnit          string    `json:"metricUnit" gorm:"size:24"`
	EffectiveAt         time.Time `json:"effectiveAt"`
	Evidence            string    `json:"evidence" gorm:"size:2000"`
	RelatedCode         string    `json:"relatedCode" gorm:"size:64;index"`
}

func (item *TemperatureWindow) GetBase() *BaseModel { return &item.BaseModel }

func (item TemperatureWindow) TableName() string { return "temperature_windows" }

var TemperatureWindowInitialStatus = "draft"
