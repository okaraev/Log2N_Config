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

type FileManager struct {
	config               DBConfig
	GetFunction          func(filter interface{}, config interface{}) ([]byte, error)
	UpdateFunction       func(filter interface{}, update interface{}, config interface{}) error
	UpdateAndGetFunction func(filter interface{}, update interface{}, config interface{}) ([]byte, error)
	InsertFunction       func(insert interface{}, config interface{}) error
	DeleteFunction       func(filter interface{}, config interface{}) error
}

func (f FileManager) Get(filter interface{}) ([]bson.M, error) {
	bytes, err := f.GetFunction(filter, f.config)
	if err != nil {
		return nil, err
	}
	result := []bson.M{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f FileManager) GetOne(filter interface{}) (bson.M, error) {
	bytes, err := f.GetFunction(filter, f.config)
	if err != nil {
		return nil, err
	}
	result := []bson.M{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

func (f FileManager) Update(filter interface{}, update interface{}) error {
	err := f.UpdateFunction(filter, update, f.config)
	if err != nil {
		return err
	}
	return nil
}

func (f FileManager) Insert(insert interface{}) error {
	err := f.InsertFunction(insert, f.config)
	if err != nil {
		return err
	}
	return nil
}

func (f FileManager) Delete(filter interface{}) error {
	err := f.DeleteFunction(filter, f.config)
	if err != nil {
		return err
	}
	return nil
}

func (f FileManager) UpdateAndGet(filter interface{}, update interface{}) (bson.M, error) {
	bytes, err := f.UpdateAndGetFunction(filter, update, f.config)
	if err != nil {
		return nil, err
	}
	result := bson.M{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ConfigFrom(in interface{}) (DBConfig, error) {
	conf := DBConfig{}
	bytes, err := json.Marshal(in)
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func GetDoc(filter interface{}, config interface{}) ([]byte, error) {
	dbconfig, err := ConfigFrom(config)
	if err != nil {
		return nil, err
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
	dbconfig, err := ConfigFrom(config)
	if err != nil {
		return err
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
	dbconfig, err := ConfigFrom(config)
	if err != nil {
		return nil, err
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
	dbconfig, err := ConfigFrom(config)
	if err != nil {
		return err
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
	dbconfig, err := ConfigFrom(config)
	if err != nil {
		return err
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

func FileManagerCreate(conf DBConfig) FileManager {
	fm := FileManager{
		config:               conf,
		GetFunction:          GetDoc,
		InsertFunction:       AddDoc,
		UpdateFunction:       SetDoc,
		UpdateAndGetFunction: SetGetDoc,
		DeleteFunction:       RemoveDoc,
	}
	return fm
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
