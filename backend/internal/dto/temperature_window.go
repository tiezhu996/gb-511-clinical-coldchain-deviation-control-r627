package dto

import "time"

// CreateTemperatureWindow is the public write contract for 温控规则. Status is deliberately
// omitted so callers cannot bypass the service state machine.
type CreateTemperatureWindow struct {
	Code                string    `json:"code" binding:"required,min=2,max=64"`
	Name                string    `json:"name" binding:"required,min=2,max=160"`
	Description         string    `json:"description" binding:"max=1000"`
	Facility            string    `json:"facility" binding:"required,max=120"`
	Owner               string    `json:"owner" binding:"required,max=120"`
	Category            string    `json:"category" binding:"required,max=80"`
	RiskLevel           string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue         float64   `json:"metricValue"`
	MetricUnit          string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt         time.Time `json:"effectiveAt" binding:"required"`
	Evidence            string    `json:"evidence" binding:"max=2000"`
	RelatedCode         string    `json:"relatedCode" binding:"max=64"`
	ProductClass        string    `json:"productClass" binding:"max=100"`
	MinimumCelsius      float64   `json:"minimumCelsius"`
	MaximumCelsius      float64   `json:"maximumCelsius"`
	MaxExcursionMinutes int       `json:"maxExcursionMinutes" binding:"min=0,max=10080"`
	QualityOwner        string    `json:"qualityOwner" binding:"max=120"`
}

type UpdateTemperatureWindow struct {
	ExpectedVersion     uint      `json:"expectedVersion" binding:"required"`
	Name                string    `json:"name" binding:"required,min=2,max=160"`
	Description         string    `json:"description" binding:"max=1000"`
	Facility            string    `json:"facility" binding:"required,max=120"`
	Owner               string    `json:"owner" binding:"required,max=120"`
	Category            string    `json:"category" binding:"required,max=80"`
	RiskLevel           string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue         float64   `json:"metricValue"`
	MetricUnit          string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt         time.Time `json:"effectiveAt" binding:"required"`
	Evidence            string    `json:"evidence" binding:"max=2000"`
	RelatedCode         string    `json:"relatedCode" binding:"max=64"`
	ProductClass        string    `json:"productClass" binding:"max=100"`
	MinimumCelsius      float64   `json:"minimumCelsius"`
	MaximumCelsius      float64   `json:"maximumCelsius"`
	MaxExcursionMinutes int       `json:"maxExcursionMinutes" binding:"min=0,max=10080"`
	QualityOwner        string    `json:"qualityOwner" binding:"max=120"`
}
