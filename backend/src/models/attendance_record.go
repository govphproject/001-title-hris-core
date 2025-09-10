package models

// AttendanceRecord represents a time-in/time-out or daily attendance entry.
type AttendanceRecord struct {
    RecordID   string `bson:"record_id" json:"record_id"`
    EmployeeID string `bson:"employee_id" json:"employee_id"`
    Date       string `bson:"date" json:"date"` // YYYY-MM-DD
    TimeIn     string `bson:"time_in,omitempty" json:"time_in,omitempty"`
    TimeOut    string `bson:"time_out,omitempty" json:"time_out,omitempty"`
    Notes      string `bson:"notes,omitempty" json:"notes,omitempty"`
    Version    int    `bson:"version,omitempty" json:"version,omitempty"`
}
