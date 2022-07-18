package main

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGetHash(t *testing.T) {
	hashed := GetHash("1Log2Notification!")
	if hashed != "1f8fc71151f40c365769b8b37d230916cb9dc487e02752b262c2d5d3cd564e3a" {
		t.Errorf("Have to return 1f8fc71151f40c365769b8b37d230916cb9dc487e02752b262c2d5d3cd564e3a instead of %s", hashed)
	}
}

func TestPasswordComplexityCheck(t *testing.T) {
	if PasswordComplexityCheck("1") {
		t.Error("Have to return false for a single digit")
	}
	if PasswordComplexityCheck("A") {
		t.Error("Have to return false for a single capital letter")
	}
	if PasswordComplexityCheck("z") {
		t.Error("Have to return false for a single small letter")
	}
	if PasswordComplexityCheck("1Az") {
		t.Error("Have to return false for a length mismatch")
	}
	if PasswordComplexityCheck("AaAaAaAa") {
		t.Error("Have to return false for a missing digit")
	}
	if PasswordComplexityCheck("A1A1A1A1") {
		t.Error("Have to return false for a missing small letter")
	}
	if PasswordComplexityCheck("a1a1a1a1") {
		t.Error("Have to return false for a missing capital letter")
	}
	if !PasswordComplexityCheck("A1aA1a1a") {
		t.Error("Have to return true for a complexity")
	}
}

var FM FileManager = FileManagerCreate(ConfigDBConf)

func MockGet(filter interface{}, config interface{}) ([]byte, error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	mid := bson.M{}
	err = json.Unmarshal(bytes, &mid)
	if err != nil {
		return nil, err
	}
	mid["LastName"] = "Hewlett"
	mid["Age"] = 35
	mid["School"] = "High"
	array := []bson.M{}
	array = append(array, mid)
	bytes, err = json.Marshal(array)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func MockUpdateAndGet(filter interface{}, update interface{}, config interface{}) ([]byte, error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	mid := bson.M{}
	err = json.Unmarshal(bytes, &mid)
	if err != nil {
		return nil, err
	}
	mid["LastName"] = "Hewlett"
	bytes, err = json.Marshal(mid)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func TestGetDocument(t *testing.T) {
	FM.GetFunction = MockGet
	result, err := FM.Get(bson.M{"FirstName": "Packard"})
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
	if val, ok := result[0]["FirstName"]; val != "Packard" || !ok {
		t.Errorf("Didn't get expected values; FistName: %s, ok: %v", val, ok)
	}
}

func TestGetOneDocument(t *testing.T) {
	FM.GetFunction = MockGet
	result, err := FM.GetOne(bson.M{"FirstName": "Packard"})
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
	if val, ok := result["FirstName"]; val != "Packard" || !ok {
		t.Errorf("Didn't get expected values; FistName: %s, ok: %v", val, ok)
	}
}

func TestUpdateAndGetDocument(t *testing.T) {
	FM.UpdateAndGetFunction = MockUpdateAndGet
	result, err := FM.UpdateAndGet(bson.M{"FirstName": "Packard"}, bson.M{"LastName": "Hewlett"})
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
	if val, ok := result["LastName"]; val != "Hewlett" || !ok {
		t.Errorf("Didn't get expected values; LastName: %s, ok: %v", val, ok)
	}
}
