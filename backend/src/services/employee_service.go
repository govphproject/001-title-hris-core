package services

import (
    "context"
    "errors"
    "fmt"
    "net/mail"
    "sort"
    "strings"
    "time"
)

// EmployeeService provides CRUD for employees and validation around the repo.
type EmployeeService struct {
    repo EmployeeRepo
}

func NewEmployeeService(repo EmployeeRepo) *EmployeeService {
    return &EmployeeService{repo: repo}
}

func (s *EmployeeService) Create(ctx context.Context, emp map[string]interface{}) (map[string]interface{}, error) {
    if emp == nil {
        return nil, errors.New("employee is required")
    }
    if emp["employee_id"] == nil {
        return nil, errors.New("employee_id required")
    }
    // basic email validation if provided
    if e, ok := emp["email"].(string); ok && e != "" {
        if _, err := mail.ParseAddress(e); err != nil {
            return nil, errors.New("invalid email")
        }
    }
    // basic hire_date validation if provided (YYYY-MM-DD)
    if hd, ok := emp["hire_date"].(string); ok && hd != "" {
        if _, err := time.Parse("2006-01-02", hd); err != nil {
            return nil, errors.New("invalid hire_date")
        }
    }
    return s.repo.Create(ctx, emp)
}

func (s *EmployeeService) Get(ctx context.Context, id string) (map[string]interface{}, error) {
    return s.repo.Get(ctx, id)
}

func (s *EmployeeService) Update(ctx context.Context, id string, patch map[string]interface{}, expectedVersion *int) (map[string]interface{}, error) {
    return s.repo.Update(ctx, id, patch, expectedVersion)
}

func (s *EmployeeService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

// List returns employees with simple filtering and pagination. `filter` matches
// exact values for fields present in the document. `page` is 1-based.
func (s *EmployeeService) List(ctx context.Context, page, perPage int, filter map[string]interface{}) ([]map[string]interface{}, int, error) {
    items, err := s.repo.List(ctx)
    if err != nil {
        return nil, 0, err
    }
    // apply simple filters (ignore special keys like "sort")
    var filtered []map[string]interface{}
    for _, it := range items {
        ok := true
        for k, v := range filter {
            if k == "sort" {
                continue
            }
            if iv, has := it[k]; !has {
                ok = false
                break
            } else {
                if fmt.Sprintf("%v", iv) != fmt.Sprintf("%v", v) {
                    ok = false
                    break
                }
            }
        }
        if ok {
            filtered = append(filtered, it)
        }
    }
    // support simple sort instructions via filter["sort"], e.g. "employee_id" or "-employee_id"
    // multiple fields can be comma-separated. Remove sort key from filter processing above if present.
    var sortInstr string
    if filter != nil {
        if sraw, has := filter["sort"]; has {
            if ss, ok := sraw.(string); ok {
                sortInstr = ss
            }
        }
    }
    if sortInstr != "" {
        fields := strings.Split(sortInstr, ",")
        // perform a stable multi-key sort: apply keys in reverse so the first field has highest priority
        for i := len(fields) - 1; i >= 0; i-- {
            f := strings.TrimSpace(fields[i])
            desc := false
            if strings.HasPrefix(f, "-") {
                desc = true
                f = strings.TrimPrefix(f, "-")
            }
            // sort by field f
            sort.SliceStable(filtered, func(a, b int) bool {
                av := getStringValue(filtered[a], f)
                bv := getStringValue(filtered[b], f)
                if desc {
                    return av > bv
                }
                return av < bv
            })
        }
    }
    total := len(filtered)
    if perPage <= 0 {
        perPage = 20
    }
    if page <= 0 {
        page = 1
    }
    start := (page - 1) * perPage
    if start >= total {
        return []map[string]interface{}{}, total, nil
    }
    end := start + perPage
    if end > total {
        end = total
    }
    return filtered[start:end], total, nil
}

// getStringValue returns a string representation of the value at key in the doc.
// key supports dot notation for nested maps, e.g. "legal_name.first".
func getStringValue(doc map[string]interface{}, key string) string {
    if key == "" {
        return ""
    }
    parts := strings.Split(key, ".")
    var cur interface{} = doc
    for _, p := range parts {
        if m, ok := cur.(map[string]interface{}); ok {
            if v, has := m[p]; has {
                cur = v
                continue
            }
            return ""
        }
        return ""
    }
    return fmt.Sprintf("%v", cur)
}
