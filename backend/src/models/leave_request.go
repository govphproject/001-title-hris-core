package models

// LeaveRequest represents an employee's request for leave/time off.
type LeaveRequest struct {
    RequestID   string  `bson:"request_id" json:"request_id"`
    EmployeeID  string  `bson:"employee_id" json:"employee_id"`
    Type        string  `bson:"type" json:"type"` // e.g. vacation, sick
    StartDate   string  `bson:"start_date" json:"start_date"`
    EndDate     string  `bson:"end_date" json:"end_date"`
    Reason      string  `bson:"reason,omitempty" json:"reason,omitempty"`
    Status      string  `bson:"status" json:"status"` // pending, approved, rejected
    ApproverID  *string `bson:"approver_id,omitempty" json:"approver_id,omitempty"`
    CreatedAt   int64   `bson:"created_at,omitempty" json:"created_at,omitempty"`
    Version     int     `bson:"version,omitempty" json:"version,omitempty"`
}
