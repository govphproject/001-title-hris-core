package models

// Role represents an authorization role in the HRIS (e.g., Admin, HR, Employee).
type Role struct {
    RoleID   string `bson:"role_id" json:"role_id"`
    Name     string `bson:"name" json:"name"`
    Nickname string `bson:"nickname,omitempty" json:"nickname,omitempty"`
    Version  int    `bson:"version,omitempty" json:"version,omitempty"`
}
