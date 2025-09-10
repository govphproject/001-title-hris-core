package models

// Employee represents a person in the HRIS.
type Employee struct {
    EmployeeID          string               `bson:"employee_id" json:"employee_id"`
    LegalName           map[string]string    `bson:"legal_name" json:"legal_name"`
    PreferredName       string               `bson:"preferred_name,omitempty" json:"preferred_name,omitempty"`
    Email               string               `bson:"email,omitempty" json:"email,omitempty"`
    Phone               string               `bson:"phone,omitempty" json:"phone,omitempty"`
    HireDate            string               `bson:"hire_date,omitempty" json:"hire_date,omitempty"`
    TerminationDate     *string              `bson:"termination_date,omitempty" json:"termination_date,omitempty"`
    EmploymentStatus    string               `bson:"employment_status,omitempty" json:"employment_status,omitempty"`
    JobHistory          []JobHistoryEntry    `bson:"job_history,omitempty" json:"job_history,omitempty"`
    CompensationRecords []CompensationRecord `bson:"compensation_records,omitempty" json:"compensation_records,omitempty"`
    ManagerID           string               `bson:"manager_id,omitempty" json:"manager_id,omitempty"`
    Version             int                  `bson:"version,omitempty" json:"version,omitempty"`
}

// JobHistoryEntry captures a single role/assignment the employee held.
type JobHistoryEntry struct {
    Title      string  `bson:"title" json:"title"`
    Department string  `bson:"department,omitempty" json:"department,omitempty"`
    StartDate  string  `bson:"start_date,omitempty" json:"start_date,omitempty"`
    EndDate    *string `bson:"end_date,omitempty" json:"end_date,omitempty"`
}

// CompensationRecord describes a pay/compensation change or entry.
type CompensationRecord struct {
    Amount        float64 `bson:"amount" json:"amount"`
    Currency      string  `bson:"currency,omitempty" json:"currency,omitempty"`
    EffectiveDate string  `bson:"effective_date,omitempty" json:"effective_date,omitempty"`
    Type          string  `bson:"type,omitempty" json:"type,omitempty"` // e.g. salary, bonus
}
