package models

// UserAccount represents an authentication user for the HRIS.
type UserAccount struct {
    Username     string   `bson:"username" json:"username"`
    PasswordHash string   `bson:"password_hash" json:"-"`
    Email        string   `bson:"email,omitempty" json:"email,omitempty"`
    Roles        []string `bson:"roles,omitempty" json:"roles,omitempty"`
    CreatedAt    int64    `bson:"created_at,omitempty" json:"created_at,omitempty"`
}
