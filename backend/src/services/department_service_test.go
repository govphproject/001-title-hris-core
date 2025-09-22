package services

import (
	"context"
	"testing"
)

func TestDepartmentService_CRUDAndLinking(t *testing.T) {
    ctx := context.Background()
    repo := NewInMemoryDepartmentRepo()
    svc := NewDepartmentService(repo)

    in := map[string]interface{}{"id": "dept-1", "name": "HR"}
    created, err := svc.Create(ctx, in)
    if err != nil {
        t.Fatalf("create failed: %v", err)
    }
    if created["id"] != "dept-1" {
        t.Fatalf("unexpected id")
    }

    got, err := svc.Get(ctx, "dept-1")
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    if got["name"] != "HR" {
        t.Fatalf("name mismatch")
    }

    // add employee
    _, err = svc.AddEmployee(ctx, "dept-1", "emp-1")
    if err != nil {
        t.Fatalf("add employee failed: %v", err)
    }
    after, _ := svc.Get(ctx, "dept-1")
    ids, _ := after["employee_ids"].([]string)
    if len(ids) != 1 || ids[0] != "emp-1" {
        t.Fatalf("employee link missing: %v", ids)
    }

    // remove employee
    _, err = svc.RemoveEmployee(ctx, "dept-1", "emp-1")
    if err != nil {
        t.Fatalf("remove employee failed: %v", err)
    }
    after2, _ := svc.Get(ctx, "dept-1")
    ids2, _ := after2["employee_ids"].([]string)
    if len(ids2) != 0 {
        t.Fatalf("employee not removed: %v", ids2)
    }

    // update
    patched, err := svc.Update(ctx, "dept-1", map[string]interface{}{"description": "People"}, nil)
    if err != nil {
        t.Fatalf("update failed: %v", err)
    }
    if patched["description"] != "People" {
        t.Fatalf("update not applied")
    }

    if err := svc.Delete(ctx, "dept-1"); err != nil {
        t.Fatalf("delete failed: %v", err)
    }
    if _, err := svc.Get(ctx, "dept-1"); err == nil {
        t.Fatalf("expected not found after delete")
    }
}
