package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func GetDocument(filter interface{}, dbconfig DBConfig) ([]bson.M, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	result := []bson.M{}
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return result, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	configs, err := Collection.Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer configs.Close(ctx)
	err = configs.All(ctx, &result)
	if err != nil {
		return result, err
	}
	cancel()
	return result, nil
}

func GetSingleDocument(filter interface{}, dbconfig DBConfig) (bson.M, error) {
	result := bson.M{}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return result, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	configs := Collection.FindOne(ctx, filter)
	if err != nil {
		return result, err
	}
	err = configs.Err()
	if err != nil {
		return result, err
	}
	err = configs.Decode(result)
	if err != nil {
		return result, err
	}
	cancel()
	return result, nil
}

func SetDocument(filter interface{}, update interface{}, dbconfig DBConfig) error {
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

func SetGetDocument(filter interface{}, update interface{}, dbconfig DBConfig) (bson.M, error) {
	result := bson.M{}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return result, err
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
		return result, err
	}
	err = updateResult.Decode(result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func AddDocument(document interface{}, dbconfig DBConfig) error {
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

func RemoveDocument(filter interface{}, dbconfig DBConfig) error {
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
