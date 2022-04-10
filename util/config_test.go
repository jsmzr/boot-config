package util

import (
	"reflect"
	"testing"
)

type Foo struct {
	Id    int
	Name  string
	Price float64
	State bool
	Tag   []string
	Other Bar `name:"bar"`
}

type Bar struct {
	Id    int    `default:"123"`
	Name  string `required:"true"`
	Price float64
	State bool
	Tag   []int
}

func TestResolve(t *testing.T) {
	dict := make(map[string]interface{})
	dict["foo.bar.name"] = "bar"
	dict["foo.bar.price"] = 88.8
	dict["foo.bar.state"] = false
	dict["foo.bar.tag"] = []int{1, 2, 3}

	dict["foo.tag"] = []string{"foo", "bar"}
	dict["foo.id"] = 1
	dict["foo.name"] = "foo"
	dict["foo.price"] = 2.33
	dict["foo.state"] = true
	var foo Foo
	if err := ResolveStruct(&dict, "foo", &foo); err != nil {
		t.Fatal(err)
	}
	if foo.Id != dict["foo.id"] || foo.Name != dict["foo.name"] || foo.Price != dict["foo.price"] || foo.State != dict["foo.state"] || !reflect.DeepEqual(foo.Tag, dict["foo.tag"]) {
		// || foo.Price != dict["foo.price"] || reflect.DeepEqual(foo.Tag, dict["foo.tag"])  {
		t.Fatal("注入的值不符合预期")
	}

	bar := foo.Other
	if bar.Id != 123 || bar.Name != dict["foo.bar.name"] || bar.Price != dict["foo.bar.price"] || bar.State != dict["foo.bar.state"] || !reflect.DeepEqual(bar.Tag, dict["foo.bar.tag"]) {
		t.Fatal("注入的值不符合预期")
	}
}
