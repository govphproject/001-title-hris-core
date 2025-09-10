package services

import (
    "context"
    "testing"
)

func TestInMemoryUserStore_CreateAndValidate(t *testing.T) {
    s := NewInMemoryUserStore()
    ctx := context.Background()
    if err := s.CreateUser(ctx, "alice", "s3cr3t", []string{"admin"}); err != nil {
        t.Fatalf("create user failed: %v", err)
    }
    ok, roles, err := s.ValidateCredentials(ctx, "alice", "s3cr3t")
    if err != nil {
        t.Fatalf("validate returned error: %v", err)
    }
    if !ok {
        t.Fatalf("expected credentials to validate")
    }
    if len(roles) != 1 || roles[0] != "admin" {
        t.Fatalf("unexpected roles: %v", roles)
    }
}
