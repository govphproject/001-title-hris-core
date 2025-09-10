package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	// jwt used by api and middleware packages
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/ronaldpalay/hris/src/services"
	"github.com/ronaldpalay/hris/src/middleware"
	apipkg "github.com/ronaldpalay/hris/src/api"
)

type Employee struct {
	EmployeeID    string                 `json:"employee_id" bson:"employee_id"`
	LegalName     map[string]interface{} `json:"legal_name" bson:"legal_name"`
	PreferredName string                 `json:"preferred_name,omitempty" bson:"preferred_name,omitempty"`
	Email         string                 `json:"email,omitempty" bson:"email,omitempty"`
	HireDate      string                 `json:"hire_date,omitempty" bson:"hire_date,omitempty"`
	Version       int                    `json:"version,omitempty" bson:"version,omitempty"`
}

// (removed unused in-memory globals left from earlier iterations)

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

// (removed unused payroll in-memory globals)

// JWT secret for the running server (can be overridden with HRIS_JWT_SECRET)
var jwtSecret = []byte(getEnv("HRIS_JWT_SECRET", "dev-jwt-secret"))

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ...existing code...

// Optional MongoDB backing
var (
	mongoClient *mongo.Client
	empColl     *mongo.Collection
	userColl    *mongo.Collection
	useMongo    bool
	authStore   services.AuthStore
	employeeRepo services.EmployeeRepo
	payrollRepo services.PayrollRepo
)

// simple user model for auth
type User struct {
	Username     string   `bson:"username" json:"username"`
	PasswordHash string   `bson:"password_hash" json:"-"`
	Roles        []string `bson:"roles,omitempty" json:"roles,omitempty"`
}

// in-memory users fallback when Mongo is not configured
// (now handled by services.NewInMemoryUserStore; globals removed)

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

	// decide auth store (mongo-backed when available, otherwise in-memory)
	if useMongo && mongoClient != nil && userColl != nil {
		dbName := getEnv("MONGO_DB", "hris")
		usersCollName := getEnv("MONGO_USERS_COLLECTION", "users")
		authStore = services.NewMongoUserStore(mongoClient, dbName, usersCollName)
	} else {
		authStore = services.NewInMemoryUserStore()
	}

	// wire employee repo
	if useMongo && mongoClient != nil && empColl != nil {
		employeeRepo = services.NewMongoEmployeeRepo(empColl)
	} else {
		employeeRepo = services.NewInMemoryEmployeeRepo()
	}

	// wire payroll repo
	if useMongo && mongoClient != nil {
		payrollRepo = services.NewMongoPayrollRepo(mongoClient.Database(getEnv("MONGO_DB", "hris")).Collection(getEnv("MONGO_PAYROLL_COLLECTION", "payroll")))
	} else {
		payrollRepo = services.NewInMemoryPayrollRepo()
	}

	// ensure seeded users exist (will use authStore)
	if err := initUsers(initCtx); err != nil {
		fmt.Printf("init users failed: %v\n", err)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	apiGroup := r.Group("/api")
	// register auth and user routes (authStore needs to be passed)
	apipkg.RegisterAuthRoutes(apiGroup, authStore, jwtSecret)
	// register employee and payroll routes
	apipkg.RegisterEmployeeRoutes(apiGroup, employeeRepo)
	apipkg.RegisterPayrollRoutes(apiGroup, payrollRepo)

	// secure endpoints (require JWT)
	secure := apiGroup.Group("/secure")
	secure.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// protected employees listing for authenticated clients (used by tests)
		secure.GET("/employees", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			items, err := employeeRepo.List(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
		})

		// admin only example
		secure.GET("/admin", middleware.RequireRole("admin", jwtSecret), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"admin": true})
		})
		// user creation (admin only)
		secure.POST("/users", middleware.RequireRole("admin", jwtSecret), createUser)
	}

	return r
}

func main() {
	r := NewRouter(context.Background())
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("server run error: %v\n", err)
		os.Exit(1)
	}
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
	// use authStore to create/upsert admin user
	return authStore.CreateUserWithHash(ctx, adminUser, string(h), []string{"admin"})
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
	// delegate to authStore
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := authStore.CreateUser(ctx, in.Username, in.Password, in.Roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"username": in.Username, "roles": in.Roles})
}
