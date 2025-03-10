package models

// Filter defines the structure for advanced filtering
type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
