package models

// OrganizationUnit represents a department or unit in the org chart.
type OrganizationUnit struct {
    UnitID   string `bson:"unit_id" json:"unit_id"`
    Name     string `bson:"name" json:"name"`
    ParentID string `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
    Version  int    `bson:"version,omitempty" json:"version,omitempty"`
}
