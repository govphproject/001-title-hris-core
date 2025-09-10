package services

import (
    "context"
    "testing"
)

func TestInMemoryEmployeeRepo_CreateGetUpdateDelete(t *testing.T) {
    r := NewInMemoryEmployeeRepo()
    ctx := context.Background()
    doc := map[string]interface{}{"employee_id": "emp-1", "legal_name": map[string]interface{}{"first": "Alice"}, "version": 1}
    if _, err := r.Create(ctx, doc); err != nil {
        t.Fatalf("create failed: %v", err)
    }
    got, err := r.Get(ctx, "emp-1")
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    if got["employee_id"] != "emp-1" {
        t.Fatalf("unexpected id: %v", got["employee_id"])
    }
    // update with correct version
    upd := map[string]interface{}{"preferred_name": "Ally", "version": 1}
    out, err := r.Update(ctx, "emp-1", upd, func() *int { v := 1; return &v }())
    if err != nil {
        t.Fatalf("update failed: %v", err)
    }
    if out["preferred_name"] != "Ally" {
        t.Fatalf("update didn't persist")
    }
    // update with wrong version
    if _, err := r.Update(ctx, "emp-1", upd, func() *int { v := 999; return &v }()); err == nil {
        t.Fatalf("expected version mismatch")
    }
    // delete
    if err := r.Delete(ctx, "emp-1"); err != nil {
        t.Fatalf("delete failed: %v", err)
    }
}
