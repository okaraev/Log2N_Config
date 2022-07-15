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

func GetTeamConfig(team string, config DBConfig) ([]TeamConfig, error) {
	filter := bson.M{}
	if team != "%" {
		filter = bson.M{"Team": team}
	}
	docs, err := GetMongoDoc(filter, config)
	if err != nil {
		return nil, err
	}
	slice := []TeamConfig{}
	for _, doc := range docs {
		bytes, err := bson.Marshal(doc)
		if err != nil {
			return nil, err
		}
		structure := TeamConfig{}
		err = bson.Unmarshal(bytes, &structure)
		if err != nil {
			return nil, err
		}
		slice = append(slice, structure)
	}
	return slice, nil
}

func GetSingleTeamConfig(team string, configName string, config DBConfig) (TeamConfig, error) {
	myconfig := TeamConfig{}
	filter := bson.M{
		"Team": team,
		"Name": configName,
	}
	doc, err := GetMongoDoc(filter, config)
	if err != nil {
		return myconfig, err
	}
	bytes, err := bson.Marshal(doc[0])
	if err != nil {
		return myconfig, err
	}
	err = bson.Unmarshal(bytes, &myconfig)
	if err != nil {
		return myconfig, err
	}
	return myconfig, nil
}

func AddTeamConfig(tc TeamConfig, dbconfig DBConfig) error {
	configbytes, err := bson.Marshal(tc)
	if err != nil {
		return err
	}
	myConfig := bson.M{}
	err = bson.Unmarshal(configbytes, &myConfig)
	if err != nil {
		return err
	}
	err = AddMongoDoc(myConfig, dbconfig)
	if err != nil {
		return err
	}
	return nil
}

func SetTeamConfig(TC bson.M, config DBConfig) error {
	err := SetMongoDoc(TC, config)
	if err != nil {
		return err
	}
	return nil
}

func RemoveTeamConfig(configName string, config DBConfig) error {
	filter := bson.M{"Name": configName}
	err := RemoveMongoDoc(filter, config)
	if err != nil {
		return err
	}
	return nil
}

func GetTeamUser(Name string, dbconfig DBConfig) (Account, error) {
	filter := bson.M{"Name": Name}
	docs, err := GetMongoDoc(filter, dbconfig)
	if err != nil {
		return Account{}, err
	}
	structure := Account{}
	for _, doc := range docs {
		docbytes, err := bson.Marshal(doc)
		if err != nil {
			return Account{}, err
		}
		err = bson.Unmarshal(docbytes, &structure)
		if err != nil {
			return Account{}, err
		}
	}
	return structure, nil
}

func AddTeamUser(user Account, dbconfig DBConfig) error {
	userbytes, err := bson.Marshal(user)
	if err != nil {
		return err
	}
	myUser := bson.M{}
	err = bson.Unmarshal(userbytes, &myUser)
	if err != nil {
		return err
	}
	err = AddMongoDoc(myUser, dbconfig)
	if err != nil {
		return err
	}
	return nil
}

func SetTeamUser(TC bson.M, config DBConfig) error {
	err := SetMongoDoc(TC, config)
	if err != nil {
		return err
	}
	return nil
}

func RemoveTeamUser(userName string, config DBConfig) error {
	filter := bson.M{"Name": userName}
	err := RemoveMongoDoc(filter, config)
	if err != nil {
		return err
	}
	return nil
}

func GetMongoDoc(filter bson.M, dbconfig DBConfig) (bson.A, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return bson.A{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return bson.A{}, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	configs, err := Collection.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	defer configs.Close(ctx)
	returnObject := bson.A{}
	err = configs.All(ctx, &returnObject)
	if err != nil {
		return bson.A{}, err
	}
	cancel()
	return returnObject, nil
}

func SetMongoDoc(doc bson.M, config DBConfig) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(config.Connectionstring))
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
	Database := client.Database(config.Database)
	Collection := Database.Collection(config.Collection)
	filter := bson.M{"Name": doc["Name"]}
	updateeResult, err := Collection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: doc},
		},
	)
	if err != nil {
		return err
	}
	if updateeResult.MatchedCount == 0 {
		return errors.New("no document found to update")
	}
	return nil
}

func AddMongoDoc(m bson.M, dbconfig DBConfig) error {
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
	_, err = Collection.InsertOne(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func RemoveMongoDoc(filter bson.M, dbconfig DBConfig) error {
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

func GetDocument(filter interface{}, dbconfig DBConfig) (bson.A, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbconfig.Connectionstring))
	if err != nil {
		return bson.A{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return bson.A{}, err
	}
	defer client.Disconnect(ctx)
	Database := client.Database(dbconfig.Database)
	Collection := Database.Collection(dbconfig.Collection)
	configs, err := Collection.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	defer configs.Close(ctx)
	returnObject := bson.A{}
	err = configs.All(ctx, &returnObject)
	if err != nil {
		return bson.A{}, err
	}
	cancel()
	return returnObject, nil
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
