package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const KEY_NAME = "name"
const KEY_DEFAULT_VALUE = "default"
const KEY_REQUIRED = "required"

// 获取属性名称
func getFieldName(field *reflect.StructField) string {
	name := field.Tag.Get(KEY_NAME)
	if name == "_" {
		return name
	}
	// 优先使用 tag 中的名称
	if name != "" {
		return name
	}
	name = field.Name
	// 转换为小写字母开头
	return strings.ToLower(name[0:1]) + name[1:]
}

// 转换为 int64
func convertInt64(i interface{}) (int64, error) {
	switch t := i.(type) {
	case int64:
		return t, nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case string:
		tmp, err := strconv.Atoi(t)
		if err != nil {
			return 0, err
		}
		return int64(tmp), nil
	default:
		return 0, fmt.Errorf("不能转换 %v 为 int64", i)
	}
}

// 转换为 float64
func convertFloat64(i interface{}) (float64, error) {
	switch t := i.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("不能转换 %v 为 float64", i)
	}
}

func convertBool(i interface{}) (bool, error) {
	switch t := i.(type) {
	case bool:
		return t, nil
	case string:
		return t == "true", nil
	default:
		return false, fmt.Errorf("不能转换 %v 为 bool", i)
	}
}

// 设置属性值
func setFieldValue(fieldValue *reflect.Value, value interface{}) error {
	switch fieldValue.Type().Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		if converted, err := convertInt64(value); err != nil {
			return err
		} else {
			fieldValue.SetInt(converted)
		}
	case reflect.Float32, reflect.Float64:
		if converted, err := convertFloat64(value); err != nil {
			return err
		} else {
			fieldValue.SetFloat(converted)
		}
	case reflect.String:
		if converted, ok := value.(string); ok {
			fieldValue.SetString(converted)
		} else {
			fieldValue.SetString(fmt.Sprintf("%v", value))
		}
	case reflect.Bool:
		if converted, err := convertBool(value); err != nil {
			return err
		} else {
			fieldValue.SetBool(converted)
		}
	default:
		return fmt.Errorf("不支持该类型变量")
	}
	return nil
}

// 基础类型数组解析
// []struct 类型的解析暂不支持
func resolveArrayByReflect(dict *map[string]interface{}, prefix string, structValue *reflect.Value, tag reflect.StructTag) error {
	values, ok := (*dict)[prefix]
	required := tag.Get(KEY_REQUIRED)
	if !ok {
		if required == "true" {
			return fmt.Errorf("未配置 %s", prefix)
		}
		return nil
	}
	reflectArr := reflect.ValueOf(values)
	if reflectArr.Kind() != reflect.Slice {
		if required == "true" {
			return fmt.Errorf("配置的 %s 不是数组类型", prefix)
		}
		return nil
	}
	arrLen := reflectArr.Len()

	reflectValues := make([]reflect.Value, arrLen)
	valueType := structValue.Type().Elem()
	for i := 0; i < arrLen; i++ {
		item := reflectArr.Index(i).Interface()
		itemValue := reflect.New(valueType).Elem()
		setFieldValue(&itemValue, item)
		reflectValues[i] = itemValue
	}
	result := reflect.Append(*structValue, reflectValues...)
	structValue.Set(result)
	return nil
}

// 通过反射解析结构体
func resolveStructByReflect(dict *map[string]interface{}, prefix string, structType reflect.Type, structValue *reflect.Value) error {
	count := structType.NumField()
	if count == 0 {
		return nil
	}
	if prefix != "" {
		prefix += "."
	}
	for index := 0; index < count; index++ {
		field := structType.Field(index)
		name := getFieldName(&field)
		if name == "_" {
			continue
		}
		key := prefix + name
		fieldValue := structValue.Field(index)
		switch field.Type.Kind() {
		case reflect.Struct:
			resolveStructByReflect(dict, key, field.Type, &fieldValue)
			continue
		case reflect.Slice:
			resolveArrayByReflect(dict, key, &fieldValue, field.Tag)
			continue
		}
		value, ok := (*dict)[key]
		if ok {
			setFieldValue(&fieldValue, value)
			continue
		}
		// 未配置时从 tag 取默认配置
		tag := field.Tag
		defaultValue := tag.Get(KEY_DEFAULT_VALUE)
		if defaultValue != "" {
			setFieldValue(&fieldValue, defaultValue)
			continue
		}
		// 无默认配置时检查配置项是否是必须的，如果是必须的则返回异常
		requiredFlag := tag.Get(KEY_REQUIRED)
		if requiredFlag == "true" {
			return fmt.Errorf("未设置配置 [%s] 的值", key)
		}
	}
	return nil
}

// 解析结构体
func ResolveStruct(dict *map[string]interface{}, prefix string, p interface{}) error {
	pt := reflect.TypeOf(p)
	if pt.Kind() != reflect.Ptr {
		return fmt.Errorf("需要使用结构体地址作为参数")
	}
	structType := pt.Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("需要使用结构体地址作为参数")
	}
	pv := reflect.ValueOf(p).Elem()
	return resolveStructByReflect(dict, prefix, structType, &pv)
}

/**
 * 扁平化键值对
 * 如：{"aa": {"bb": "123", "cc": "234"}} 将被转换为 {"aa.bb": "123", "aa.cc": "234"}
 */
func FlatMap(dict *map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range *dict {
		if reflect.TypeOf(value).Kind() == reflect.Map {
			tmpValue := value.(map[interface{}]interface{})
			for k, v := range FlatMap(&tmpValue) {
				result[fmt.Sprintf("%s.%s", key, k)] = v
			}
		} else {
			result[fmt.Sprintf("%s", key)] = value
		}
	}
	return result
}

/**
 * 扁平化键值对
 * 如：{"aa": {"bb": "123", "cc": "234"}} 将被转换为 {"aa.bb": "123", "aa.cc": "234"}
 */
func FlatStringMap(dict *map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range *dict {
		if reflect.TypeOf(value).Kind() == reflect.Map {
			tmpValue := value.(map[string]interface{})
			for k, v := range FlatStringMap(&tmpValue) {
				result[fmt.Sprintf("%s.%s", key, k)] = v
			}
		} else {
			result[key] = value
		}
	}
	return result
}
