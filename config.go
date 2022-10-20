package config

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type Adapter interface {
	Load(name string) (Configer, error)
}

type Configer interface {
	GetJson() string
}

var adapters = make(map[string]Adapter)
var cache = make(map[string]*gjson.Result)

var instance Configer
var resource gjson.Result

func Register(name string, adapter Adapter) {
	_, ok := adapters[name]
	if ok {
		panic(fmt.Errorf("adapter [%s] already register", name))
	}
	adapters[name] = adapter
}

func InitInstance(name string, filename string) error {
	adapter, ok := adapters[name]
	if !ok {
		return fmt.Errorf("not found [%s] adapter", name)
	}

	if instance != nil {
		fmt.Printf("Instance already init, will be [%s] replaced. \n", name)
	}
	if newInstance, err := adapter.Load(filename); err != nil {
		return err
	} else {
		instance = newInstance
		resource = gjson.Parse(instance.GetJson())
		return nil
	}
}

func Get(key string) (*gjson.Result, bool) {
	if res := cache[key]; res != nil {
		return res, true
	}
	value := resource.Get(key)
	if value.Exists() {
		cache[key] = &value
		return &value, true
	}
	return nil, false
}

func Resolve(prefix string, p interface{}) error {
	if prefix == "" {
		return json.Unmarshal([]byte(resource.String()), p)
	}
	res := cache[prefix]
	if res == nil {
		tmp := resource.Get(prefix)
		if !tmp.Exists() {
			return fmt.Errorf("not found value by [%s]", prefix)
		}
		res = &tmp
		cache[prefix] = res
	}
	return json.Unmarshal([]byte(res.String()), p)
}
