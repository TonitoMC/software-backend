package models

// Business Hours represents a time interval in which
// the clinic is open / working, could be multiple
// for a single day if for example lunch break
type BusinessHourInterval struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
