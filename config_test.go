package main

import (
	"testing"
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
