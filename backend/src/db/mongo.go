package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Connect establishes a MongoDB client using the provided URI and returns it.
// Caller is responsible for calling CloseClient when done.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
    if uri == "" {
        return nil, nil
    }
    clientOpts := options.Client().ApplyURI(uri)
    // allow a small timeout if caller provided a background context
    ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    c, err := mongo.Connect(ctx2, clientOpts)
    if err != nil {
        return nil, err
    }
    if err := c.Ping(ctx2, nil); err != nil {
        return nil, err
    }
    return c, nil
}

// CloseClient disconnects the provided Mongo client.
func CloseClient(ctx context.Context, c *mongo.Client) error {
    if c == nil {
        return nil
    }
    ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return c.Disconnect(ctx2)
}

// GetCollection returns a collection handle for the given database and collection names.
func GetCollection(c *mongo.Client, dbName, collName string) *mongo.Collection {
    if c == nil {
        return nil
    }
    return c.Database(dbName).Collection(collName)
}

