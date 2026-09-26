package dto

import "strings"

// ExcursionCumulativeQuery carries the container/window scope for the rolling
// 24h cumulative excursion read. Container is mandatory; window is optional and,
// when supplied, restricts the total to the referenced temperature rule.
type ExcursionCumulativeQuery struct {
	ContainerCode string `form:"containerCode" binding:"required,max=64"`
	WindowCode    string `form:"windowCode" binding:"max=64"`
}

type ExcursionCumulativeView struct {
	ContainerCode     string `json:"containerCode"`
	WindowCode        string `json:"windowCode"`
	WindowHours       int    `json:"windowHours"`
	CumulativeMinutes int    `json:"cumulativeMinutes"`
	OpenEventCount    int    `json:"openEventCount"`
	MaxAllowedMinutes int    `json:"maxAllowedMinutes"`
	Exceeded          bool   `json:"exceeded"`
}

func (q *ExcursionCumulativeQuery) Normalized() (string, string) {
	return strings.ToUpper(strings.TrimSpace(q.ContainerCode)), strings.ToUpper(strings.TrimSpace(q.WindowCode))
}
