package main

import "testing"

func TestDivideSuccess(t *testing.T) {
	result, err := divide(10, 2)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := 5.0
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestDivideByZero(t *testing.T) {
	result, err := divide(10, 0)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != 0 {
		t.Errorf("expected result 0, got %v", result)
	}
}
