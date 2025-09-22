package models

// OrganizationUnit represents a department or unit within the organization.
type OrganizationUnit struct {
	ID          string   `json:"id" bson:"_id"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	EmployeeIDs []string `json:"employee_ids,omitempty" bson:"employee_ids,omitempty"`
	Version     int      `json:"version" bson:"version"`
}
