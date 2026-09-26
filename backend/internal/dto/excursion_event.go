package dto

import "time"

// CreateExcursionEvent is the public write contract for 偏差事件. Status is deliberately
// omitted so callers cannot bypass the service state machine.
type CreateExcursionEvent struct {
	Code            string    `json:"code" binding:"required,min=2,max=64"`
	Name            string    `json:"name" binding:"required,min=2,max=160"`
	Description     string    `json:"description" binding:"max=1000"`
	Facility        string    `json:"facility" binding:"required,max=120"`
	Owner           string    `json:"owner" binding:"required,max=120"`
	Category        string    `json:"category" binding:"required,max=80"`
	RiskLevel       string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue     float64   `json:"metricValue"`
	MetricUnit      string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt     time.Time `json:"effectiveAt" binding:"required"`
	Evidence        string    `json:"evidence" binding:"max=2000"`
	RelatedCode     string    `json:"relatedCode" binding:"max=64"`
	ContainerCode   string    `json:"containerCode" binding:"max=64"`
	WindowCode      string    `json:"windowCode" binding:"max=64"`
	ObservedTempC   float64   `json:"observedTempC"`
	DurationMinutes int       `json:"durationMinutes" binding:"min=0,max=10080"`
	DetectedAt      time.Time `json:"detectedAt"`
	SensorEvidence  string    `json:"sensorEvidence" binding:"max=2000"`
	Reviewer        string    `json:"reviewer" binding:"max=80"`
}

type UpdateExcursionEvent struct {
	ExpectedVersion uint      `json:"expectedVersion" binding:"required"`
	Name            string    `json:"name" binding:"required,min=2,max=160"`
	Description     string    `json:"description" binding:"max=1000"`
	Facility        string    `json:"facility" binding:"required,max=120"`
	Owner           string    `json:"owner" binding:"required,max=120"`
	Category        string    `json:"category" binding:"required,max=80"`
	RiskLevel       string    `json:"riskLevel" binding:"required,oneof=low medium high critical"`
	MetricValue     float64   `json:"metricValue"`
	MetricUnit      string    `json:"metricUnit" binding:"max=24"`
	EffectiveAt     time.Time `json:"effectiveAt" binding:"required"`
	Evidence        string    `json:"evidence" binding:"max=2000"`
	RelatedCode     string    `json:"relatedCode" binding:"max=64"`
	ContainerCode   string    `json:"containerCode" binding:"max=64"`
	WindowCode      string    `json:"windowCode" binding:"max=64"`
	ObservedTempC   float64   `json:"observedTempC"`
	DurationMinutes int       `json:"durationMinutes" binding:"min=0,max=10080"`
	DetectedAt      time.Time `json:"detectedAt"`
	SensorEvidence  string    `json:"sensorEvidence" binding:"max=2000"`
	Reviewer        string    `json:"reviewer" binding:"max=80"`
}
