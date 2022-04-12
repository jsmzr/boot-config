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

var instance Configer
var resource gjson.Result

func Register(name string, adapter Adapter) {
	_, ok := adapters[name]
	if ok {
		panic(fmt.Errorf("适配器 [%s] 已经装载，请勿重复加载", name))
	}
	adapters[name] = adapter
}

func InitInstance(name string, filename string) error {
	adapter, ok := adapters[name]
	if !ok {
		return fmt.Errorf("找不到适配器 [%s] 请确定已经装载", name)
	}

	if instance != nil {
		fmt.Printf("配置适配器已经存在，将使用 [%s] 替换\n", name)
	}
	fmt.Printf("开始使用配置 [%s] 载入适配器\n", filename)
	if newInstance, err := adapter.Load(filename); err != nil {
		return err
	} else {
		instance = newInstance
		resource = gjson.Parse(instance.GetJson())
		return nil
	}
}

func Get(key string) (interface{}, bool) {
	result := resource.Get(key)
	if result.Exists() {
		return result.Value(), true
	}
	return nil, false
}

func Resolve(prefix string, p interface{}) error {
	result := resource.Get(prefix)
	if !result.Exists() {
		return fmt.Errorf("not found value by [%s]", prefix)
	}
	return json.Unmarshal([]byte(result.String()), p)
}
