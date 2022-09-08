package main

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

type FileManager struct {
	config               interface{}
	GetFunction          func(filter interface{}, config interface{}) ([]byte, error)
	UpdateFunction       func(filter interface{}, update interface{}, config interface{}) error
	UpdateAndGetFunction func(filter interface{}, update interface{}, config interface{}) ([]byte, error)
	InsertFunction       func(insert interface{}, config interface{}) error
	DeleteFunction       func(filter interface{}, config interface{}) error
	SendMessageFunction  func(message interface{}, configParams interface{}) error
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

func (f FileManager) SendMessage(message interface{}) error {
	err := f.SendMessageFunction(message, f.config)
	return err
}

func GetFileManagerDefaultInstace(conf interface{}) FileManager {
	fm := FileManager{
		config:               conf,
		GetFunction:          GetDoc,
		InsertFunction:       AddDoc,
		UpdateFunction:       SetDoc,
		UpdateAndGetFunction: SetGetDoc,
		DeleteFunction:       RemoveDoc,
		SendMessageFunction:  SendMessage,
	}
	return fm
}

func GetFileManagerOverloadInstace(
	conf interface{},
	getFunction func(filter interface{}, config interface{}) ([]byte, error),
	insertFunction func(intert interface{}, config interface{}) error,
	updateFunction func(filter interface{}, update interface{}, config interface{}) error,
	updateandGetFunction func(filter interface{}, update interface{}, config interface{}) ([]byte, error),
	deleteFunction func(filter interface{}, config interface{}) error,
	sendMessageFunction func(message interface{}, confParams interface{}) error,

) FileManager {
	fm := FileManager{
		config:               conf,
		GetFunction:          getFunction,
		InsertFunction:       insertFunction,
		UpdateFunction:       updateFunction,
		UpdateAndGetFunction: updateandGetFunction,
		DeleteFunction:       deleteFunction,
		SendMessageFunction:  sendMessageFunction,
	}
	return fm
}
