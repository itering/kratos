package config

import "testing"

type testApply struct {
	Name string `json:"name"`
}

func TestApplyJSON(t *testing.T) {
	b := []byte(`{"name": "kratos"}`)
	a := new(testApply)
	if err := ApplyJSON(b, a); err != nil {
		t.Fatal(err)
	}
	if a.Name != "kratos" {
		t.Fatalf("name is invalid %s", a.Name)
	}
}

func TestApplyYAML(t *testing.T) {
	b := []byte(`name: kratos`)
	a := new(testApply)
	if err := ApplyYAML(b, a); err != nil {
		t.Fatal(err)
	}
	if a.Name != "kratos" {
		t.Fatalf("name is invalid %s", a.Name)
	}
}

func TestApplyTOML(t *testing.T) {
	b := []byte(`name = "kratos"`)
	a := new(testApply)
	if err := ApplyTOML(b, a); err != nil {
		t.Fatal(err)
	}
	if a.Name != "kratos" {
		t.Fatalf("name is invalid %s", a.Name)
	}
}
