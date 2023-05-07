package api

import (
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"reflect"
	"time"
)

func ValidateWithTags(s interface{}, prefix interface{}) error {
	// Get the type of the struct
	if reflect.ValueOf(s).Kind() == reflect.Ptr {
		v := reflect.ValueOf(s)
		if v.IsNil() {
			return fmt.Errorf("request body is empty")
		}
		s = v.Elem().Interface()
	}
	t := reflect.TypeOf(s)

	// Iterate over each field in the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag

		// Get the "validate" tag value
		validateTag := tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Get the field value
		value := reflect.ValueOf(s).FieldByName(field.Name)

		switch validateTag {
		case "required":
			// Check if the value is zero or empty
			if (value.Kind() == reflect.Ptr && value.IsNil()) || !value.IsValid() {
				return fmt.Errorf("%v%s field is required", prefix, field.Name)
			}
		case "range":
			if value.Kind() == reflect.Slice && value.Len() == 0 {
				return fmt.Errorf("%v%s field must not be empty", prefix, field.Name)
			}
			if val, ok := value.Interface().([]int32); !ok ||
				value.Len() != 2 ||
				val[0] > val[1] {
				return fmt.Errorf("%v%s field must be a range e.g [1,2]", prefix, field.Name)
			}
		case "json_date":
			if value.Kind() == reflect.Struct && value.Type() == reflect.TypeOf(time.Time{}) {
				zeroTime := time.Time{}
				if value.Interface() == zeroTime {
					return fmt.Errorf("%v%s field is required", prefix, field.Name)
				}
			}
			date := time.Time(value.Interface().(models.JSONDate))
			if date.Before(time.Now()) {
				return fmt.Errorf("%s field must be date after current date", date.Format(time.DateOnly))
			}
		}
	}

	return nil
}
