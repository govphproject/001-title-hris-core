package services

import (
    "context"
    "errors"
    "sync"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// PayrollRepo abstracts payroll storage operations.
type PayrollRepo interface {
    Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error)
    ListByEmployee(ctx context.Context, employeeID string) ([]map[string]interface{}, error)
}

// InMemoryPayrollRepo is a simple in-memory payroll store.
type InMemoryPayrollRepo struct {
    mu    sync.Mutex
    m     map[string]map[string]interface{}
    byEmp map[string][]string
}

func NewInMemoryPayrollRepo() *InMemoryPayrollRepo {
    return &InMemoryPayrollRepo{m: map[string]map[string]interface{}{}, byEmp: map[string][]string{}}
}

func (r *InMemoryPayrollRepo) Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    id, _ := doc["id"].(string)
    r.m[id] = doc
    eid, _ := doc["employee_id"].(string)
    r.byEmp[eid] = append(r.byEmp[eid], id)
    return doc, nil
}

func (r *InMemoryPayrollRepo) ListByEmployee(ctx context.Context, employeeID string) ([]map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    ids := r.byEmp[employeeID]
    out := make([]map[string]interface{}, 0, len(ids))
    for _, id := range ids {
        if d, ok := r.m[id]; ok {
            out = append(out, d)
        }
    }
    return out, nil
}

// MongoPayrollRepo stores payrolls in MongoDB.
type MongoPayrollRepo struct {
    coll *mongo.Collection
}

func NewMongoPayrollRepo(coll *mongo.Collection) *MongoPayrollRepo {
    return &MongoPayrollRepo{coll: coll}
}

func (r *MongoPayrollRepo) Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error) {
    if _, ok := doc["id"]; !ok {
        return nil, errors.New("id required")
    }
    if _, err := r.coll.InsertOne(ctx, doc); err != nil {
        return nil, err
    }
    return doc, nil
}

func (r *MongoPayrollRepo) ListByEmployee(ctx context.Context, employeeID string) ([]map[string]interface{}, error) {
    cur, err := r.coll.Find(ctx, bson.M{"employee_id": employeeID})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []map[string]interface{}
    for cur.Next(ctx) {
        var doc map[string]interface{}
        if err := cur.Decode(&doc); err != nil {
            continue
        }
        out = append(out, doc)
    }
    return out, nil
}
