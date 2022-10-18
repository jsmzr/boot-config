package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

type TestConfigAdapter struct{}
type TestConfig struct{}

type Test2ConfigAdapter struct{}
type Test2Config struct{}

type Test3ConfigAdapter struct{}
type Test3Config struct{}

type Cc struct {
	Dd int `json:"dd"`
}
type Bb struct {
	Bb int64 `json:"bb"`
	Cc Cc    `json:"cc"`
}
type TestData struct {
	Foo string `json:"foo"`
	Arr []int  `json:"arr"`
	Aa  Bb     `json:"aa"`
}

var testData0 = TestData{Foo: "boo", Arr: []int{1, 2, 3}, Aa: Bb{123, Cc{456}}}
var testData1 = TestData{Foo: "foo", Arr: []int{4, 5, 6}, Aa: Bb{321, Cc{123}}}

func (t *TestConfigAdapter) Load(name string) (Configer, error) {
	return &TestConfig{}, nil
}

func (t *TestConfig) GetJson() string {
	b, err := json.Marshal(testData0)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

func (t *Test2ConfigAdapter) Load(name string) (Configer, error) {
	return &Test2Config{}, nil
}

func (t *Test2Config) GetJson() string {
	b, err := json.Marshal(testData1)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}
func (t *Test3ConfigAdapter) Load(name string) (Configer, error) {
	return nil, errors.New("load error")
}

func (t *Test3Config) GetJson() string {
	return ""
}

func TestRegister(t *testing.T) {
	adapters = make(map[string]Adapter)
	Register("test", &TestConfigAdapter{})
	if v, ok := adapters["test"]; !ok || v == nil {
		t.Fatal("register failed")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("register 2")
		}
	}()

	Register("test", &TestConfigAdapter{})
}

func TestInitInstance(t *testing.T) {
	adapters = make(map[string]Adapter)
	instance = nil
	Register("test", &TestConfigAdapter{})
	if err := InitInstance("no-test", "file_1"); err == nil {
		t.Fatal("no-test isn't register")
	}
	if err := InitInstance("test", "file_2"); err != nil {
		t.Fatal(err.Error())
	}
	if instance == nil {
		t.Fatal("load config failed")
	}
	fmt.Printf("old instance: %T, old resource: %v\n", instance, resource)
	oldInstance := instance
	oldResource := resource
	// replace
	Register("test2", &Test2ConfigAdapter{})
	if err := InitInstance("test2", "any_file3"); err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("old instance: %T, old resource: %v\n", instance, resource)
	if oldInstance == instance || oldResource.Raw == resource.Raw {
		t.Fatal("instance not reload")
	}

	Register("test3", &Test3ConfigAdapter{})
	if err := InitInstance("test3", "any"); err == nil {
		t.Fatal("load adapter should be failed")
	}
}

func TestGet(t *testing.T) {
	adapters = make(map[string]Adapter)
	instance = nil
	cache = make(map[string]*gjson.Result)
	Register("test", &TestConfigAdapter{})
	if err := InitInstance("test", "file_2"); err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(resource.Raw)
	foo, ok := Get("foo")
	fmt.Printf("foo: %v, ok: %v\n", foo, ok)
	if !ok || foo.Str != testData0.Foo {
		t.Fatal("get config value error")
	}
	ab, ok := Get("aa.bb")
	fmt.Printf("aa.bb: %v, ok: %v\n", ab, ok)
	if !ok || ab.Int() != testData0.Aa.Bb {
		t.Fatal("get config value error")
	}
	arr, ok := Get("arr")
	fmt.Printf("arr: %v, ok: %v\n", arr, ok)
	if !ok || !arr.IsArray() {
		t.Fatal("get config value error")
	}
	arrObj := arr.Array()
	if arrObj[0].Int() != int64(testData0.Arr[0]) ||
		arrObj[1].Int() != int64(testData0.Arr[1]) ||
		arrObj[2].Int() != int64(testData0.Arr[2]) {
		t.Fatal("get config array error")
	}

	if cache["foo"] == nil {
		t.Fatal("cache is failed")
	}
	foo1, ok := Get("foo")
	if !ok || foo1.Str != testData0.Foo {
		t.Fatal("Get value is failed from cache")
	}

	if _, ok := Get("notFound"); ok {
		t.Fatal("should be not found")
	}
}

func TestResolve(t *testing.T) {
	adapters = make(map[string]Adapter)
	instance = nil
	cache = make(map[string]*gjson.Result)
	Register("test", &TestConfigAdapter{})
	if err := InitInstance("test", "file_2"); err != nil {
		t.Fatal(err.Error())
	}
	var obj TestData
	if err := Resolve("", &obj); err != nil {
		t.Fatal(err.Error())
	}
	if obj.Foo != testData0.Foo ||
		obj.Aa.Bb != testData0.Aa.Bb ||
		!reflect.DeepEqual(obj.Arr, testData0.Arr) {
		t.Fatal("resolve data failed")
	}
	var b Bb
	if err := Resolve("aa", &b); err != nil {
		t.Fatal(err.Error())
	}
	if b.Bb != testData0.Aa.Bb {
		t.Fatal("resolve data failed")
	}

	var b1 Bb
	if err := Resolve("aa", &b1); err != nil {
		t.Fatal(err.Error())
	}
	if b1.Bb != testData0.Aa.Bb {
		t.Fatal("resolve data failed")
	}

	var c Cc
	if err := Resolve("aa.cc", &c); err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(c.Dd)
	if c.Dd != testData0.Aa.Cc.Dd {
		t.Fatal("resolve data failed")
	}
	var c1 Cc
	if err := Resolve("aabb", &c1); err == nil {
		t.Fatal("value should be not found")
	}

}
