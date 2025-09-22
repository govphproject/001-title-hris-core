package services

import (
	"context"
	"errors"
)

// DepartmentService provides higher-level logic on top of DepartmentRepo.
type DepartmentService struct {
    repo DepartmentRepo
}

func NewDepartmentService(repo DepartmentRepo) *DepartmentService {
    return &DepartmentService{repo: repo}
}

func (s *DepartmentService) Create(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
    if in == nil {
        return nil, errors.New("input required")
    }
    if in["id"] == nil || in["name"] == nil {
        return nil, errors.New("id and name required")
    }
    // ensure employee_ids is present
    if _, ok := in["employee_ids"]; !ok {
        in["employee_ids"] = []string{}
    }
    return s.repo.Create(ctx, in)
}

func (s *DepartmentService) Get(ctx context.Context, id string) (map[string]interface{}, error) {
    return s.repo.Get(ctx, id)
}

func (s *DepartmentService) Update(ctx context.Context, id string, patch map[string]interface{}, expectedVersion *int) (map[string]interface{}, error) {
    return s.repo.Update(ctx, id, patch, expectedVersion)
}

func (s *DepartmentService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *DepartmentService) List(ctx context.Context) ([]map[string]interface{}, error) {
    return s.repo.List(ctx)
}

// AddEmployee links an employee ID to the department if not already present.
func (s *DepartmentService) AddEmployee(ctx context.Context, deptID, empID string) (map[string]interface{}, error) {
    d, err := s.repo.Get(ctx, deptID)
    if err != nil {
        return nil, err
    }
    
    // Get current version for optimistic locking
    version, _ := d["version"].(int)
    
    ids, _ := d["employee_ids"].([]string)
    for _, e := range ids {
        if e == empID {
            return d, nil
        }
    }
    ids = append(ids, empID)
    patch := map[string]interface{}{"employee_ids": ids}
    return s.repo.Update(ctx, deptID, patch, &version)
}

// RemoveEmployee unlinks an employee ID from the department.
func (s *DepartmentService) RemoveEmployee(ctx context.Context, deptID, empID string) (map[string]interface{}, error) {
    d, err := s.repo.Get(ctx, deptID)
    if err != nil {
        return nil, err
    }
    
    // Get current version for optimistic locking
    version, _ := d["version"].(int)
    
    ids, _ := d["employee_ids"].([]string)
    var out []string
    for _, e := range ids {
        if e == empID {
            continue
        }
        out = append(out, e)
    }
    patch := map[string]interface{}{"employee_ids": out}
    return s.repo.Update(ctx, deptID, patch, &version)
}
