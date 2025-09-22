package services

import (
	"context"
	"errors"
	"sync"
)

// DepartmentRepo defines persistence operations for OrganizationUnit
type DepartmentRepo interface {
	Create(ctx context.Context, d map[string]interface{}) (map[string]interface{}, error)
	Get(ctx context.Context, id string) (map[string]interface{}, error)
	Update(ctx context.Context, id string, patch map[string]interface{}, expectedVersion *int) (map[string]interface{}, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]map[string]interface{}, error)
}

// InMemoryDepartmentRepo is a simple in-memory implementation for tests.
type InMemoryDepartmentRepo struct {
	mu    sync.Mutex
	store map[string]map[string]interface{}
}

func NewInMemoryDepartmentRepo() *InMemoryDepartmentRepo {
	return &InMemoryDepartmentRepo{store: map[string]map[string]interface{}{}}
}

func (r *InMemoryDepartmentRepo) Create(ctx context.Context, d map[string]interface{}) (map[string]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, _ := d["id"].(string)
	if id == "" {
		return nil, errors.New("id required")
	}
	if _, ok := r.store[id]; ok {
		return nil, errors.New("exists")
	}
	// ensure version
	if _, ok := d["version"]; !ok {
		d["version"] = 1
	}
	r.store[id] = cloneMap(d)
	return cloneMap(d), nil
}

func (r *InMemoryDepartmentRepo) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.store[id]; ok {
		return cloneMap(v), nil
	}
	return nil, errors.New("not found")
}

func (r *InMemoryDepartmentRepo) Update(ctx context.Context, id string, patch map[string]interface{}, expectedVersion *int) (map[string]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.store[id]
	if !ok {
		return nil, errors.New("not found")
	}
	if expectedVersion != nil {
		if curV, _ := cur["version"].(int); curV != *expectedVersion {
			return nil, errors.New("version mismatch")
		}
	}
	for k, v := range patch {
		cur[k] = v
	}
	// bump version
	if cv, _ := cur["version"].(int); cv > 0 {
		cur["version"] = cv + 1
	} else {
		cur["version"] = 1
	}
	r.store[id] = cloneMap(cur)
	return cloneMap(cur), nil
}

func (r *InMemoryDepartmentRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return errors.New("not found")
	}
	delete(r.store, id)
	return nil
}

func (r *InMemoryDepartmentRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]map[string]interface{}, 0, len(r.store))
	for _, v := range r.store {
		out = append(out, cloneMap(v))
	}
	return out, nil
}

// cloneMap makes a shallow copy of a map[string]interface{}
func cloneMap(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{}, len(m))
	for k, v := range m {
		o[k] = v
	}
	return o
}
