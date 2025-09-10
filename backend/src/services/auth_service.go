package services

import (
    "context"

    "golang.org/x/crypto/bcrypt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// AuthStore defines the methods an auth backing store must implement.
type AuthStore interface {
    CreateUser(ctx context.Context, username, password string, roles []string) error
    ValidateCredentials(ctx context.Context, username, password string) (bool, []string, error)
    // CreateUserWithHash allows creating a user when you already have a password hash
    CreateUserWithHash(ctx context.Context, username, passwordHash string, roles []string) error
}

// InMemoryUserStore is a tiny store used for tests and simple setups.
type InMemoryUserStore struct {
    users map[string]string // username -> passwordHash
    roles map[string][]string
}

func NewInMemoryUserStore() *InMemoryUserStore {
    return &InMemoryUserStore{users: map[string]string{}, roles: map[string][]string{}}
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, username, password string, roles []string) error {
    h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    return s.CreateUserWithHash(ctx, username, string(h), roles)
}

func (s *InMemoryUserStore) CreateUserWithHash(ctx context.Context, username, passwordHash string, roles []string) error {
    s.users[username] = passwordHash
    s.roles[username] = roles
    _ = ctx
    return nil
}

func (s *InMemoryUserStore) ValidateCredentials(ctx context.Context, username, password string) (bool, []string, error) {
    h, ok := s.users[username]
    if !ok {
        return false, nil, nil
    }
    if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(password)); err != nil {
        return false, nil, nil
    }
    return true, s.roles[username], nil
}

// MongoUserStore implements AuthStore using a MongoDB collection.
type MongoUserStore struct {
    coll *mongo.Collection
}

// NewMongoUserStore creates a MongoUserStore using an existing mongo client and collection names.
func NewMongoUserStore(client *mongo.Client, dbName, collName string) *MongoUserStore {
    if client == nil {
        return &MongoUserStore{coll: nil}
    }
    return &MongoUserStore{coll: client.Database(dbName).Collection(collName)}
}

func (m *MongoUserStore) CreateUser(ctx context.Context, username, password string, roles []string) error {
    h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    return m.CreateUserWithHash(ctx, username, string(h), roles)
}

func (m *MongoUserStore) CreateUserWithHash(ctx context.Context, username, passwordHash string, roles []string) error {
    if m.coll == nil {
        return nil
    }
    doc := bson.M{"username": username, "password_hash": passwordHash, "roles": roles}
    // upsert so creating same user updates roles/hash
    opts := options.Update().SetUpsert(true)
    _, err := m.coll.UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": doc}, opts)
    return err
}

func (m *MongoUserStore) ValidateCredentials(ctx context.Context, username, password string) (bool, []string, error) {
    if m.coll == nil {
        return false, nil, nil
    }
    var out struct {
        PasswordHash string   `bson:"password_hash"`
        Roles        []string `bson:"roles"`
    }
    err := m.coll.FindOne(ctx, bson.M{"username": username}).Decode(&out)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, nil, nil
        }
        return false, nil, err
    }
    if err := bcrypt.CompareHashAndPassword([]byte(out.PasswordHash), []byte(password)); err != nil {
        return false, nil, nil
    }
    return true, out.Roles, nil
}

