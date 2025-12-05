package tracker

import (
	"fmt"
	"reflect"
	"strconv"
)

// toInterfaceSlice 将 []string 转换为 []interface{}
func toInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// getHeadersByTags 通过反射获取结构体字段的name标签值作为表头
func getHeadersByTags(v interface{}) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	headers := make([]string, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 获取name标签
		if nameTag := field.Tag.Get("name"); nameTag != "" {
			headers = append(headers, nameTag)
		}
	}

	return headers
}

// getValuesByTags 通过反射获取结构体字段的值作为表格行数据
func getValuesByTags(v interface{}) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	values := make([]string, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 只处理有name标签的字段
		if nameTag := field.Tag.Get("name"); nameTag != "" {
			fieldVal := val.Field(i)
			values = append(values, formatValue(fieldVal))
		}
	}

	return values
}

// formatValue 格式化字段值为字符串
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

