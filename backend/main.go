package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Employee struct {
	EmployeeID    string                 `json:"employee_id" bson:"employee_id"`
	LegalName     map[string]interface{} `json:"legal_name" bson:"legal_name"`
	PreferredName string                 `json:"preferred_name,omitempty" bson:"preferred_name,omitempty"`
	Email         string                 `json:"email,omitempty" bson:"email,omitempty"`
	HireDate      string                 `json:"hire_date,omitempty" bson:"hire_date,omitempty"`
	Version       int                    `json:"version,omitempty" bson:"version,omitempty"`
}

var (
	storeMu   sync.Mutex
	employees = map[string]Employee{}
)

// Payroll model and in-memory store for fallback
type PayrollRecord struct {
	ID         string  `json:"id" bson:"id"`
	EmployeeID string  `json:"employee_id" bson:"employee_id"`
	Gross      float64 `json:"gross" bson:"gross"`
	Deductions float64 `json:"deductions" bson:"deductions"`
	Taxes      float64 `json:"taxes" bson:"taxes"`
	Net        float64 `json:"net" bson:"net"`
	Period     string  `json:"period" bson:"period"`
}

var (
	payrollMu          sync.Mutex
	payrolls           = map[string]PayrollRecord{}
	payrollsByEmployee = map[string][]string{}
)

// JWT secret for the running server (can be overridden with HRIS_JWT_SECRET)
var jwtSecret = []byte(getEnv("HRIS_JWT_SECRET", "dev-jwt-secret"))

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Optional MongoDB backing
var (
	mongoClient *mongo.Client
	empColl     *mongo.Collection
	userColl    *mongo.Collection
	useMongo    bool
)

// simple user model for auth
type User struct {
	Username     string   `bson:"username" json:"username"`
	PasswordHash string   `bson:"password_hash" json:"-"`
	Roles        []string `bson:"roles,omitempty" json:"roles,omitempty"`
}

// in-memory users fallback when Mongo is not configured
var inMemoryUsers = map[string]string{}
var inMemoryRoles = map[string][]string{}

func initMongo(ctx context.Context) error {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil
	}
	dbName := getEnv("MONGO_DB", "hris")
	collName := getEnv("MONGO_COLLECTION", "employees")
	usersCollName := getEnv("MONGO_USERS_COLLECTION", "users")
	clientOpts := options.Client().ApplyURI(uri)
	c, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}
	if err := c.Ping(ctx, nil); err != nil {
		return err
	}
	mongoClient = c
	empColl = mongoClient.Database(dbName).Collection(collName)
	userColl = mongoClient.Database(dbName).Collection(usersCollName)
	useMongo = true
	return nil
}

// newRouter builds the Gin engine and performs initialization so tests can reuse it.
func NewRouter(ctx context.Context) *gin.Engine {
	r := gin.Default()

	// try to initialize Mongo if configured
	initCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := initMongo(initCtx); err != nil {
		// log to stdout for now
		fmt.Printf("mongo init failed: %v\n", err)
		useMongo = false
	}

	// ensure seeded users exist
	if err := initUsers(initCtx); err != nil {
		fmt.Printf("init users failed: %v\n", err)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// auth
		api.POST("/auth/login", loginHandler)

		api.GET("/employees", listEmployees)
		api.POST("/employees", createEmployee)
		api.GET("/employees/:id", getEmployee)
		api.PUT("/employees/:id", updateEmployee)
		api.DELETE("/employees/:id", deleteEmployee)

		// payroll endpoints
		api.POST("/payroll", createPayroll)
		api.GET("/payroll/employee/:id", listPayrollByEmployee)
	}

	// secure endpoints (require JWT)
	secure := api.Group("/secure")
	secure.Use(AuthMiddleware())
	{
		secure.GET("/employees", listEmployees)
		secure.POST("/employees", createEmployee)
		secure.GET("/employees/:id", getEmployee)
		secure.PUT("/employees/:id", updateEmployee)
		secure.DELETE("/employees/:id", deleteEmployee)
		// admin only example
		secure.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"admin": true})
		})
		secure.POST("/users", RequireRole("admin"), createUser)
	}

	return r
}

func main() {
	r := NewRouter(context.Background())
	r.Run(":8080")
}

// loginHandler issues a JWT for valid credentials (very small demo; replace with real auth)
func loginHandler(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	// lookup user
	var storedHash string
	if useMongo && userColl != nil {
		var u User
		filter := bson.M{"username": creds.Username}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := userColl.FindOne(ctx, filter).Decode(&u); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
			return
		}
		storedHash = u.PasswordHash
	} else {
		h, ok := inMemoryUsers[creds.Username]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
			return
		}
		storedHash = h
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
		return
	}
	// determine roles (from Mongo if available, otherwise default admin user to admin role)
	var roles []string
	if useMongo && userColl != nil {
		var u User
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := userColl.FindOne(ctx, bson.M{"username": creds.Username}).Decode(&u); err == nil {
			roles = u.Roles
		}
	}
	if roles == nil && creds.Username == getEnv("HRIS_ADMIN_USER", "admin") {
		roles = []string{"admin"}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   creds.Username,
		"roles": roles,
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})
	s, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": s})
}

// AuthMiddleware validates JWT from the Authorization header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if len(h) < 8 || h[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok := h[7:]
		// parse and validate signing method
		p, err := jwt.ParseWithClaims(tok, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			// ensure HMAC signing is used
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || p == nil || !p.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}

// RequireRole checks that the JWT contains the required role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if len(h) < 8 || h[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok := h[7:]
		// parse and validate signing method + claims
		p, err := jwt.ParseWithClaims(tok, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || p == nil || !p.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := p.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		// roles may be []interface{} (from JSON decoding), []string, or a single string
		switch rs := claims["roles"].(type) {
		case []interface{}:
			for _, r := range rs {
				if s, ok := r.(string); ok && s == role {
					c.Next()
					return
				}
			}
		case []string:
			for _, s := range rs {
				if s == role {
					c.Next()
					return
				}
			}
		case string:
			if rs == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func listEmployees(c *gin.Context) {
	if useMongo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cur, err := empColl.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer cur.Close(ctx)
		var items []Employee
		for cur.Next(ctx) {
			var e Employee
			if err := cur.Decode(&e); err != nil {
				continue
			}
			items = append(items, e)
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
		return
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	items := make([]Employee, 0, len(employees))
	for _, e := range employees {
		items = append(items, e)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func createEmployee(c *gin.Context) {
	var in map[string]interface{}
	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	legal, _ := in["legal_name"].(map[string]interface{})
	hireDate, _ := in["hire_date"].(string)
	email, _ := in["email"].(string)
	id := fmt.Sprintf("emp-%d", time.Now().UnixNano())
	emp := Employee{EmployeeID: id, LegalName: legal, Email: email, HireDate: hireDate, Version: 1}
	if useMongo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		doc := bson.M{
			"employee_id":    emp.EmployeeID,
			"legal_name":     emp.LegalName,
			"preferred_name": emp.PreferredName,
			"email":          emp.Email,
			"hire_date":      emp.HireDate,
			"version":        emp.Version,
		}
		if _, err := empColl.InsertOne(ctx, doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
		}
		c.JSON(http.StatusCreated, emp)
		return
	}
	storeMu.Lock()
	employees[id] = emp
	storeMu.Unlock()
	c.JSON(http.StatusCreated, emp)
}

func getEmployee(c *gin.Context) {
	id := c.Param("id")
	if useMongo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var e Employee
		// support legacy documents that used "employeeid" (no underscore)
		filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
		if err := empColl.FindOne(ctx, filter).Decode(&e); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, e)
		return
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	e, ok := employees[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, e)
}

func updateEmployee(c *gin.Context) {
	id := c.Param("id")
	var in map[string]interface{}
	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if useMongo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// fetch current
		var cur Employee
		// support legacy documents that used "employeeid" (no underscore)
		filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
		if err := empColl.FindOne(ctx, filter).Decode(&cur); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		// check version
		if v, ok := in["version"].(float64); ok {
			if int(v) != cur.Version {
				c.JSON(http.StatusConflict, gin.H{"error": "version mismatch"})
				return
			}
		}
		if pn, ok := in["preferred_name"].(string); ok {
			cur.PreferredName = pn
		}
		cur.Version += 1
		// update
		upd := bson.M{"$set": cur}
		if _, err := empColl.UpdateOne(ctx, filter, upd); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db update failed"})
			return
		}
		c.JSON(http.StatusOK, cur)
		return
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	e, ok := employees[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// optimistic locking
	if v, ok := in["version"].(float64); ok {
		if int(v) != e.Version {
			c.JSON(http.StatusConflict, gin.H{"error": "version mismatch"})
			return
		}
	}
	if pn, ok := in["preferred_name"].(string); ok {
		e.PreferredName = pn
	}
	e.Version += 1
	employees[id] = e
	c.JSON(http.StatusOK, e)
}

func deleteEmployee(c *gin.Context) {
	id := c.Param("id")
	if useMongo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// support legacy documents that used "employeeid" (no underscore)
		filter := bson.M{"$or": []bson.M{{"employee_id": id}, {"employeeid": id}}}
		res, err := empColl.DeleteOne(ctx, filter)
		if err != nil || res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Status(http.StatusNoContent)
		return
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	if _, ok := employees[id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	delete(employees, id)
	c.Status(http.StatusNoContent)
}

// createPayroll creates a payroll record; stores in Mongo if configured otherwise in-memory
func createPayroll(c *gin.Context) {
	var in PayrollRecord
	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if in.ID == "" {
		in.ID = fmt.Sprintf("pay-%d", time.Now().UnixNano())
	}
	if useMongo && mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		coll := mongoClient.Database(getEnv("MONGO_DB", "hris")).Collection(getEnv("MONGO_PAYROLL_COLLECTION", "payroll"))
		doc := bson.M{"id": in.ID, "employee_id": in.EmployeeID, "gross": in.Gross, "deductions": in.Deductions, "taxes": in.Taxes, "net": in.Net, "period": in.Period}
		if _, err := coll.InsertOne(ctx, doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
		}
		c.JSON(http.StatusCreated, in)
		return
	}
	// in-memory fallback
	payrollMu.Lock()
	payrolls[in.ID] = in
	payrollsByEmployee[in.EmployeeID] = append(payrollsByEmployee[in.EmployeeID], in.ID)
	payrollMu.Unlock()
	c.JSON(http.StatusCreated, in)
}

// listPayrollByEmployee returns payroll records for a given employee id
func listPayrollByEmployee(c *gin.Context) {
	id := c.Param("id")
	if useMongo && mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		coll := mongoClient.Database(getEnv("MONGO_DB", "hris")).Collection(getEnv("MONGO_PAYROLL_COLLECTION", "payroll"))
		cur, err := coll.Find(ctx, bson.M{"employee_id": id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer cur.Close(ctx)
		var items []PayrollRecord
		for cur.Next(ctx) {
			var p PayrollRecord
			if err := cur.Decode(&p); err != nil {
				continue
			}
			items = append(items, p)
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
		return
	}
	payrollMu.Lock()
	ids := payrollsByEmployee[id]
	var out []PayrollRecord
	for _, pid := range ids {
		if p, ok := payrolls[pid]; ok {
			out = append(out, p)
		}
	}
	payrollMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// initUsers ensures there's at least an admin user available for login.
func initUsers(ctx context.Context) error {
	adminUser := getEnv("HRIS_ADMIN_USER", "admin")
	adminPass := getEnv("HRIS_ADMIN_PASSWORD", "password")

	// create hash
	h, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if useMongo && userColl != nil {
		// upsert admin user
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		filter := bson.M{"username": adminUser}
		update := bson.M{"$set": bson.M{"username": adminUser, "password_hash": string(h), "roles": []string{"admin"}}}
		opts := options.Update().SetUpsert(true)
		if _, err := userColl.UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
		return nil
	}

	// in-memory fallback
	inMemoryUsers[adminUser] = string(h)
	inMemoryRoles[adminUser] = []string{"admin"}
	return nil
}

// createUser allows admin to create new users with password and roles
func createUser(c *gin.Context) {
	var in struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Roles    []string `json:"roles"`
	}
	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	if in.Username == "" || in.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	h, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}
	if useMongo && userColl != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		filter := bson.M{"username": in.Username}
		update := bson.M{"$set": bson.M{"username": in.Username, "password_hash": string(h), "roles": in.Roles}}
		opts := options.Update().SetUpsert(true)
		if _, err := userColl.UpdateOne(ctx, filter, update, opts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"username": in.Username, "roles": in.Roles})
		return
	}
	inMemoryUsers[in.Username] = string(h)
	inMemoryRoles[in.Username] = in.Roles
	c.JSON(http.StatusCreated, gin.H{"username": in.Username, "roles": in.Roles})
}
