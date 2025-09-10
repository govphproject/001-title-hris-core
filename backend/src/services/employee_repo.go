package services

import (
    "context"
    "errors"
    "sync"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// EmployeeRepo defines storage operations for employee-like documents.
type EmployeeRepo interface {
    List(ctx context.Context) ([]map[string]interface{}, error)
    Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error)
    Get(ctx context.Context, id string) (map[string]interface{}, error)
    Update(ctx context.Context, id string, doc map[string]interface{}, expectedVersion *int) (map[string]interface{}, error)
    Delete(ctx context.Context, id string) error
}

// InMemoryEmployeeRepo is a simple in-memory repo used when Mongo is not configured.
type InMemoryEmployeeRepo struct {
    mu sync.Mutex
    m  map[string]map[string]interface{}
}

func NewInMemoryEmployeeRepo() *InMemoryEmployeeRepo {
    return &InMemoryEmployeeRepo{m: map[string]map[string]interface{}{}}
}

func (r *InMemoryEmployeeRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    out := make([]map[string]interface{}, 0, len(r.m))
    for _, v := range r.m {
        out = append(out, v)
    }
    return out, nil
}

func (r *InMemoryEmployeeRepo) Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    id, _ := doc["employee_id"].(string)
    r.m[id] = doc
    return doc, nil
}

func (r *InMemoryEmployeeRepo) Get(ctx context.Context, id string) (map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if v, ok := r.m[id]; ok {
        return v, nil
    }
    return nil, mongo.ErrNoDocuments
}

func (r *InMemoryEmployeeRepo) Update(ctx context.Context, id string, doc map[string]interface{}, expectedVersion *int) (map[string]interface{}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    cur, ok := r.m[id]
    if !ok {
        return nil, mongo.ErrNoDocuments
    }
    if expectedVersion != nil {
        // support stored version as float64 (from JSON) or int
        if cvf, ok := cur["version"].(float64); ok {
            if int(cvf) != *expectedVersion {
                return nil, errors.New("version mismatch")
            }
        } else if cvi, ok := cur["version"].(int); ok {
            if cvi != *expectedVersion {
                return nil, errors.New("version mismatch")
            }
        }
    }
    // merge simple fields
    for k, v := range doc {
        cur[k] = v
    }
    // bump version
    if ver, ok := cur["version"].(float64); ok {
        cur["version"] = int(ver) + 1
    } else if ver2, ok := cur["version"].(int); ok {
        cur["version"] = ver2 + 1
    } else {
        cur["version"] = 1
    }
    r.m[id] = cur
    return cur, nil
}

func (r *InMemoryEmployeeRepo) Delete(ctx context.Context, id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.m[id]; !ok {
        return mongo.ErrNoDocuments
    }
    delete(r.m, id)
    return nil
}

// MongoEmployeeRepo stores employees in MongoDB.
type MongoEmployeeRepo struct {
    coll *mongo.Collection
}

func NewMongoEmployeeRepo(coll *mongo.Collection) *MongoEmployeeRepo {
    return &MongoEmployeeRepo{coll: coll}
}

func (r *MongoEmployeeRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
    cur, err := r.coll.Find(ctx, bson.M{})
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

func (r *MongoEmployeeRepo) Create(ctx context.Context, doc map[string]interface{}) (map[string]interface{}, error) {
    if _, ok := doc["employee_id"]; !ok {
        return nil, errors.New("employee_id required")
    }
    _, err := r.coll.InsertOne(ctx, doc)
    if err != nil {
        return nil, err
    }
    return doc, nil
}

func (r *MongoEmployeeRepo) Get(ctx context.Context, id string) (map[string]interface{}, error) {
    var doc map[string]interface{}
    filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
    if err := r.coll.FindOne(ctx, filter).Decode(&doc); err != nil {
        return nil, err
    }
    return doc, nil
}

func (r *MongoEmployeeRepo) Update(ctx context.Context, id string, doc map[string]interface{}, expectedVersion *int) (map[string]interface{}, error) {
    filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
    // fetch current
    var cur map[string]interface{}
    if err := r.coll.FindOne(ctx, filter).Decode(&cur); err != nil {
        return nil, err
    }
    if expectedVersion != nil {
        if cv, ok := cur["version"].(int); ok {
            if cv != *expectedVersion {
                return nil, errors.New("version mismatch")
            }
        }
    }
    for k, v := range doc {
        cur[k] = v
    }
    if v, ok := cur["version"].(int); ok {
        cur["version"] = v + 1
    } else {
        cur["version"] = 1
    }
    if _, err := r.coll.UpdateOne(ctx, filter, bson.M{"$set": cur}); err != nil {
        return nil, err
    }
    return cur, nil
}

func (r *MongoEmployeeRepo) Delete(ctx context.Context, id string) error {
    filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
    res, err := r.coll.DeleteOne(ctx, filter)
    if err != nil {
        return err
    }
    if res.DeletedCount == 0 {
        return mongo.ErrNoDocuments
    }
    return nil
}
