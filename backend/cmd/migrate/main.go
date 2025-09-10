package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	db := client.Database("hris")
	coll := db.Collection("employees")

	// 1) migrate legacy fields
	filter := bson.M{"$or": []bson.M{{"employeeid": bson.M{"$exists": true}}, {"legalname": bson.M{"$exists": true}}, {"hiredate": bson.M{"$exists": true}}, {"preferredname": bson.M{"$exists": true}}}}
	cur, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatalf("find failed: %v", err)
	}
	defer cur.Close(ctx)
	migrated := 0
	for cur.Next(ctx) {
		var d bson.M
		if err := cur.Decode(&d); err != nil {
			continue
		}
		upd := bson.M{}
		unset := bson.M{}
		if v, ok := d["employeeid"]; ok {
			if _, exists := d["employee_id"]; !exists {
				upd["employee_id"] = v
			}
			unset["employeeid"] = ""
		}
		if v, ok := d["legalname"]; ok {
			if _, exists := d["legal_name"]; !exists {
				upd["legal_name"] = v
			}
			unset["legalname"] = ""
		}
		if v, ok := d["hiredate"]; ok {
			if _, exists := d["hire_date"]; !exists {
				upd["hire_date"] = v
			}
			unset["hiredate"] = ""
		}
		if v, ok := d["preferredname"]; ok {
			if _, exists := d["preferred_name"]; !exists {
				upd["preferred_name"] = v
			}
			unset["preferredname"] = ""
		}
		ops := bson.M{}
		if len(upd) > 0 {
			ops["$set"] = upd
		}
		if len(unset) > 0 {
			ops["$unset"] = unset
		}
		if len(ops) > 0 {
			id := d["_id"]
			if _, err := coll.UpdateOne(ctx, bson.M{"_id": id}, ops); err == nil {
				migrated++
			}
		}
	}
	fmt.Printf("migrated %d documents\n", migrated)

	// 2) ensure employee_id exists for all docs: fill from employeeid or generate
	filterMissing := bson.M{"$or": []bson.M{{"employee_id": bson.M{"$exists": false}}, {"employee_id": nil}}}
	cur2, err := coll.Find(ctx, filterMissing)
	if err != nil {
		log.Fatalf("find missing failed: %v", err)
	}
	defer cur2.Close(ctx)
	filled := 0
	for cur2.Next(ctx) {
		var d bson.M
		if err := cur2.Decode(&d); err != nil {
			continue
		}
		if v, ok := d["employeeid"]; ok {
			if _, exists := d["employee_id"]; !exists {
				if _, err := coll.UpdateOne(ctx, bson.M{"_id": d["_id"]}, bson.M{"$set": bson.M{"employee_id": v}, "$unset": bson.M{"employeeid": ""}}); err == nil {
					filled++
				}
			}
			continue
		}
		// generate
		gen := fmt.Sprintf("emp-%d-%s", time.Now().UnixNano(), fmt.Sprintf("%v", d["_id"]))
		if _, err := coll.UpdateOne(ctx, bson.M{"_id": d["_id"]}, bson.M{"$set": bson.M{"employee_id": gen}}); err == nil {
			filled++
		}
	}
	fmt.Printf("filled %d missing employee_id\n", filled)

	// 3) create index (non-partial for better portability); duplicates should be resolved by previous steps
	idxModel := mongo.IndexModel{Keys: bson.D{bson.E{Key: "employee_id", Value: 1}}, Options: options.Index().SetUnique(true)}
	if _, err := coll.Indexes().CreateOne(ctx, idxModel); err != nil {
		log.Printf("index create warning: %v", err)
	} else {
		log.Printf("created index on employee_id")
	}

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("disconnect warning: %v", err)
	}
}
