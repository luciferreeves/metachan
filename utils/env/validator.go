package env

import (
	"fmt"
	"reflect"
)

func validateConfigInput(config any) (reflect.Value, reflect.Type, error) {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("config must be a pointer to struct")
	}
	elem := v.Elem()
	return elem, elem.Type(), nil
}
