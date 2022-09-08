package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDoc(filter interface{}, config interface{}) ([]byte, error) {
	dbconfig, ok := config.(commonconfig)
	if !ok {
		return nil, fmt.Errorf("config argument is not type of webconfig")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	configs, err := Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer configs.Close(ctx)
	result := []bson.M{}
	err = configs.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	cancel()
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func SetDoc(filter interface{}, update interface{}, config interface{}) error {
	dbconfig, ok := config.(commonconfig)
	if !ok {
		return fmt.Errorf("config argument is not type of webconfig")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	updateeResult, err := Collection.UpdateOne(
		ctx,
		filter,
		update,
	)
	if err != nil {
		return err
	}
	if updateeResult.MatchedCount == 0 {
		return fmt.Errorf("no document to update")
	}
	return nil
}

func SetGetDoc(filter interface{}, update interface{}, config interface{}) ([]byte, error) {
	dbconfig, ok := config.(commonconfig)
	if !ok {
		return nil, fmt.Errorf("config argument is not type of webconfig")
	}
	result := bson.M{}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	var ReturnDocument options.ReturnDocument = 1
	myOptions := options.FindOneAndUpdateOptions{ReturnDocument: &ReturnDocument}
	updateResult := Collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		&myOptions,
	)
	err = updateResult.Err()
	if err != nil {
		return nil, err
	}
	err = updateResult.Decode(result)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func AddDoc(document interface{}, config interface{}) error {
	dbconfig, ok := config.(commonconfig)
	if !ok {
		return fmt.Errorf("config argument is not type of webconfig")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	_, err = Collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}
	return nil
}

func RemoveDoc(filter interface{}, config interface{}) error {
	dbconfig, ok := config.(commonconfig)
	if !ok {
		return fmt.Errorf("config argument is not type of webconfig")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	delResult, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if delResult.DeletedCount == 0 {
		return errors.New("nothing to delete")
	}
	return nil
}

type DBConfig struct {
	Database         string
	Collection       string
	Connectionstring string
}

type Account struct {
	Team     string `bson:"Team" json:"Team"`
	Name     string `bson:"Name" json:"Name"`
	Password string `bson:"Password" json:"Password"`
}

func ValidateDBConfig(conf DBConfig) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(conf.Connectionstring))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	dbs, err := client.ListDatabaseNames(ctx, bson.M{"name": conf.Database})
	if err != nil {
		return err
	}
	dbfound := false
	for _, db := range dbs {
		if db == conf.Database {
			dbfound = true
			break
		}
	}
	if !dbfound {
		return fmt.Errorf("cannot find database: %s", conf.Database)
	}
	Database := client.Database(conf.Database)
	colls, err := Database.ListCollectionNames(ctx, bson.M{"name": conf.Collection})
	if err != nil {
		return err
	}
	colfound := false
	for _, col := range colls {
		if col == conf.Collection {
			colfound = true
			break
		}
	}
	if !colfound {
		return fmt.Errorf("cannot find collection: %s", conf.Collection)
	}
	return nil
}
