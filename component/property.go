package component

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// PropertySetter handles setting properties on any Go struct using reflection.
type PropertySetter struct{}

// SetProperty sets a property on an object by name.
// It tries multiple strategies:
// 1. Setter method call (Set~PropertyName~)
// 2. Direct field set (if exported and settable)
func (ps *PropertySetter) SetProperty(obj interface{}, prop string, value interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("property setter: expected pointer, got %T", obj)
	}

	// Strategy 1: Try setter method first (most common pattern in bubbles)
	if err := ps.setByMethod(v, prop, value); err == nil {
		return nil
	}

	// Strategy 2: Try direct field
	if err := ps.setByField(v, prop, value); err == nil {
		return nil
	}

	return fmt.Errorf("property %q not found on type %T", prop, obj)
}

// setByMethod tries to call Set~PropertyName~(value).
func (ps *PropertySetter) setByMethod(v reflect.Value, prop string, value interface{}) error {
	methodName := "Set" + strings.ToUpper(prop[:1]) + prop[1:]
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method %s not found", methodName)
	}

	// Convert value to method parameter type
	params := make([]reflect.Value, 1)
	paramType := method.Type().In(0)
	converted, err := ps.convertValue(value, paramType)
	if err != nil {
		return fmt.Errorf("cannot convert %T to %v for method %s", value, paramType, methodName)
	}
	params[0] = converted

	method.Call(params)
	return nil
}

// setByField sets an exported field directly.
func (ps *PropertySetter) setByField(v reflect.Value, prop string, value interface{}) error {
	fieldName := strings.ToUpper(prop[:1]) + prop[1:]
	field := v.Elem().FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("field %s not found or not settable", fieldName)
	}

	converted, err := ps.convertValue(value, field.Type())
	if err != nil {
		return err
	}
	field.Set(converted)
	return nil
}

// convertValue converts an interface{} value to the target type.
func (ps *PropertySetter) convertValue(value interface{}, target reflect.Type) (reflect.Value, error) {
	// Direct type match
	if reflect.TypeOf(value).ConvertibleTo(target) {
		return reflect.ValueOf(value).Convert(target), nil
	}

	// Handle common types
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ps.convertToInt(value, target)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ps.convertToUint(value, target)
	case reflect.Float32, reflect.Float64:
		return ps.convertToFloat(value, target)
	case reflect.Bool:
		return ps.convertToBool(value, target)
	case reflect.String:
		return ps.convertToString(value, target)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", value, target)
	}
}

func (ps *PropertySetter) convertToInt(value interface{}, target reflect.Type) (reflect.Value, error) {
	var i int64
	switch v := value.(type) {
	case int:
		i = int64(v)
	case int8:
		i = int64(v)
	case int16:
		i = int64(v)
	case int32:
		i = int64(v)
	case int64:
		i = v
	case uint:
		i = int64(v)
	case uint8:
		i = int64(v)
	case uint16:
		i = int64(v)
	case uint32:
		i = int64(v)
	case uint64:
		i = int64(v)
	case float32:
		i = int64(v)
	case float64:
		i = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as integer", v)
		}
		i = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to integer", value)
	}
	return reflect.ValueOf(i).Convert(target), nil
}

func (ps *PropertySetter) convertToUint(value interface{}, target reflect.Type) (reflect.Value, error) {
	var u uint64
	switch v := value.(type) {
	case int:
		u = uint64(v)
	case int8:
		u = uint64(v)
	case int16:
		u = uint64(v)
	case int32:
		u = uint64(v)
	case int64:
		u = uint64(v)
	case uint:
		u = uint64(v)
	case uint8:
		u = uint64(v)
	case uint16:
		u = uint64(v)
	case uint32:
		u = uint64(v)
	case uint64:
		u = v
	case float32:
		u = uint64(v)
	case float64:
		u = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as unsigned integer", v)
		}
		u = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to unsigned integer", value)
	}
	return reflect.ValueOf(u).Convert(target), nil
}

func (ps *PropertySetter) convertToFloat(value interface{}, target reflect.Type) (reflect.Value, error) {
	var f float64
	switch v := value.(type) {
	case float32:
		f = float64(v)
	case float64:
		f = v
	case int:
		f = float64(v)
	case int8:
		f = float64(v)
	case int16:
		f = float64(v)
	case int32:
		f = float64(v)
	case int64:
		f = float64(v)
	case uint:
		f = float64(v)
	case uint8:
		f = float64(v)
	case uint16:
		f = float64(v)
	case uint32:
		f = float64(v)
	case uint64:
		f = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as float", v)
		}
		f = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to float", value)
	}
	return reflect.ValueOf(f).Convert(target), nil
}

func (ps *PropertySetter) convertToBool(value interface{}, target reflect.Type) (reflect.Value, error) {
	var b bool
	switch v := value.(type) {
	case bool:
		b = v
	case string:
		switch v {
		case "true", "1", "yes":
			b = true
		case "false", "0", "no":
			b = false
		default:
			parsed, err := strconv.ParseBool(v)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse %q as boolean", v)
			}
			b = parsed
		}
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to boolean", value)
	}
	return reflect.ValueOf(b).Convert(target), nil
}

func (ps *PropertySetter) convertToString(value interface{}, target reflect.Type) (reflect.Value, error) {
	switch v := value.(type) {
	case string:
		return reflect.ValueOf(v), nil
	default:
		return reflect.ValueOf(fmt.Sprintf("%v", value)), nil
	}
}
