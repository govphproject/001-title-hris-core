package services

import (
    "context"
    "fmt"
    "testing"
)

func TestEmployeeService_CRUD(t *testing.T) {
    repo := NewInMemoryEmployeeRepo()
    svc := NewEmployeeService(repo)
    ctx := context.Background()

    emp := map[string]interface{}{"employee_id": "emp-svc-1", "legal_name": map[string]interface{}{"first": "Bob"}, "version": 1}
    created, err := svc.Create(ctx, emp)
    if err != nil {
        t.Fatalf("create failed: %v", err)
    }
    if created["employee_id"] != "emp-svc-1" {
        t.Fatalf("unexpected id")
    }
    got, err := svc.Get(ctx, "emp-svc-1")
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    if got["employee_id"] != "emp-svc-1" {
        t.Fatalf("unexpected get")
    }
    // update correctly
    patch := map[string]interface{}{"preferred_name": "Bobby", "version": 1}
    out, err := svc.Update(ctx, "emp-svc-1", patch, func() *int { v := 1; return &v }())
    if err != nil {
        t.Fatalf("update failed: %v", err)
    }
    if out["preferred_name"] != "Bobby" {
        t.Fatalf("update didn't apply")
    }
    // version conflict
    if _, err := svc.Update(ctx, "emp-svc-1", patch, func() *int { v := 999; return &v }()); err == nil {
        t.Fatalf("expected version conflict")
    }
    if err := svc.Delete(ctx, "emp-svc-1"); err != nil {
        t.Fatalf("delete failed: %v", err)
    }
}

func TestEmployeeService_ListPaginationFiltering(t *testing.T) {
    repo := NewInMemoryEmployeeRepo()
    svc := NewEmployeeService(repo)
    ctx := context.Background()
    // create 25 employees across two departments
    for i := 1; i <= 25; i++ {
        dept := "eng"
        if i%2 == 0 {
            dept = "hr"
        }
        emp := map[string]interface{}{"employee_id": fmt.Sprintf("emp-%02d", i), "department": dept, "version": 1}
        if _, err := repo.Create(ctx, emp); err != nil {
            t.Fatalf("create failed: %v", err)
        }
    }
    // page 1, 10 per page
    page1, total, err := svc.List(ctx, 1, 10, nil)
    if err != nil {
        t.Fatalf("list failed: %v", err)
    }
    if total != 25 || len(page1) != 10 {
        t.Fatalf("unexpected paging: total=%d len=%d", total, len(page1))
    }
    // filter department=hr
    pageHr, totalHr, err := svc.List(ctx, 1, 100, map[string]interface{}{"department": "hr"})
    if err != nil {
        t.Fatalf("filter list failed: %v", err)
    }
    if totalHr == 0 || len(pageHr) == 0 {
        t.Fatalf("expected filtered results")
    }
}

func TestEmployeeService_ListSorting(t *testing.T) {
    repo := NewInMemoryEmployeeRepo()
    svc := NewEmployeeService(repo)
    ctx := context.Background()
    // create 5 employees with ids out of order
    ids := []string{"e3", "e1", "e5", "e2", "e4"}
    for _, id := range ids {
        emp := map[string]interface{}{"employee_id": id, "version": 1}
        if _, err := repo.Create(ctx, emp); err != nil {
            t.Fatalf("create failed: %v", err)
        }
    }
    asc, _, err := svc.List(ctx, 1, 100, map[string]interface{}{"sort": "employee_id"})
    if err != nil {
        t.Fatalf("list failed: %v", err)
    }
    if len(asc) != 5 {
        t.Fatalf("expected 5 results")
    }
    // check ascending order
    prev := ""
    for _, e := range asc {
        id := e["employee_id"].(string)
        if prev != "" && prev > id {
            t.Fatalf("not ascending: %s came after %s", id, prev)
        }
        prev = id
    }
    desc, _, err := svc.List(ctx, 1, 100, map[string]interface{}{"sort": "-employee_id"})
    if err != nil {
        t.Fatalf("list failed: %v", err)
    }
    prev = ""
    for _, e := range desc {
        id := e["employee_id"].(string)
        if prev != "" && prev < id {
            t.Fatalf("not descending: %s came after %s", id, prev)
        }
        prev = id
    }
}

func TestEmployeeService_CreateValidation(t *testing.T) {
    repo := NewInMemoryEmployeeRepo()
    svc := NewEmployeeService(repo)
    ctx := context.Background()
    // invalid email
    badEmail := map[string]interface{}{"employee_id": "emp-x", "email": "not-an-email"}
    if _, err := svc.Create(ctx, badEmail); err == nil {
        t.Fatalf("expected email validation error")
    }
    // invalid hire_date
    badDate := map[string]interface{}{"employee_id": "emp-y", "hire_date": "01-02-2006"}
    if _, err := svc.Create(ctx, badDate); err == nil {
        t.Fatalf("expected hire_date validation error")
    }
}
