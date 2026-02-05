package configutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	tele "social-network/shared/go/telemetry"
	"strconv"
	"strings"
)

var (
	ErrBadArgument     = errors.New("argument must be pointer to struct")
	ErrUnsettableField = errors.New("field cannot be set (must be exported)")
	ErrMissingEnv      = errors.New("required env var missing")
	ErrBadConversion   = errors.New("env var conversion failed")
	ErrNoTaggedFields  = errors.New("no env-tagged fields found")
)

//TODO alex, make it return which ones were failed

// LoadConfigs loads configuration values from environment variables into the provided struct.
// It returns a boolean indicating whether any struct fields were not replaced using environment variables.
// Useful for ensuring default values do not accidentally end up in production.
func LoadConfigs(localConfig any) (bool, error) {
	tele.Debug(context.Background(), "before: @1", "configs", localConfig)
	reflectVal := reflect.ValueOf(localConfig)
	if reflectVal.Kind() != reflect.Pointer || reflectVal.Elem().Kind() != reflect.Struct {
		return false, ErrBadArgument
	}

	strctVal := reflectVal.Elem()
	strctType := strctVal.Type()

	notSwapped := false

	for i := 0; i < strctVal.NumField(); i++ {
		valField := strctVal.Field(i)
		typeField := strctType.Field(i)

		tagVal := typeField.Tag.Get("env")
		if tagVal == "" {
			continue
		}

		if !valField.CanSet() {
			return false, fmt.Errorf("%w: %s", ErrUnsettableField, typeField.Name)
		}

		envVal, ok := os.LookupEnv(tagVal)
		if !ok {
			notSwapped = true
			continue
		}

		switch valField.Kind() {
		case reflect.Int:
			v, err := strconv.ParseInt(envVal, 10, 64)
			if err != nil {
				return false, fmt.Errorf("%w (%s): %v", ErrBadConversion, tagVal, err)
			}
			valField.SetInt(v)

		case reflect.String:
			valField.SetString(envVal)

		case reflect.Float64:
			v, err := strconv.ParseFloat(envVal, 64)
			if err != nil {
				return false, fmt.Errorf("%w (%s): %v", ErrBadConversion, tagVal, err)
			}
			valField.SetFloat(v)
		case reflect.Bool:
			v, err := strconv.ParseBool(envVal)
			if err != nil {
				return false, fmt.Errorf("%w (%s): %v", ErrBadConversion, tagVal, err)
			}
			valField.SetBool(v)
		case reflect.Slice:
			parts := strings.Split(envVal, ",")
			if valField.Type().Elem().Kind() == reflect.String {
				slice := reflect.MakeSlice(valField.Type(), len(parts), len(parts))
				for i, s := range parts {
					slice.Index(i).SetString(strings.TrimSpace(s))
				}
				valField.Set(slice)
				continue
			}
			if valField.Type().Elem().Kind() == reflect.Int {
				slice := reflect.MakeSlice(valField.Type(), len(parts), len(parts))
				for i, part := range parts {
					val, err := strconv.Atoi(strings.TrimSpace(part))
					if err != nil {
						return false, fmt.Errorf("bad int in int slice")
					}
					slice.Index(i).SetInt(int64(val))
				}
				valField.Set(slice)
				continue
			}
			return false, fmt.Errorf("unsupported slice kind %s on field %s", valField.Kind(), typeField.Name)

		default:
			return false, fmt.Errorf("unsupported kind %s on field %s", valField.Kind(), typeField.Name)
		}

	}
	tele.Debug(context.Background(), "after: @1", "configs", localConfig)
	return notSwapped, nil
}
