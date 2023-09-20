// Code generated by go generate. DO NOT EDIT.

package opt

import (
	"encoding/json"
	"reflect"
)

// CustomNormalizationOption is a wrapper for an CustomNormalization option parameter. It holds
// the actual value of the option that can be accessed by calling Get.
type CustomNormalizationOption struct {
	value map[string]map[string]string
}

// CustomNormalization wraps the given value into a CustomNormalizationOption.
func CustomNormalization(v map[string]map[string]string) *CustomNormalizationOption {
	return &CustomNormalizationOption{v}
}

// Get retrieves the actual value of the option parameter.
func (o *CustomNormalizationOption) Get() map[string]map[string]string {
	if o == nil {
		return map[string]map[string]string{}
	}
	return o.value
}

// MarshalJSON implements the json.Marshaler interface for
// CustomNormalizationOption.
func (o CustomNormalizationOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface for
// CustomNormalizationOption.
func (o *CustomNormalizationOption) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.value = map[string]map[string]string{}
		return nil
	}
	return json.Unmarshal(data, &o.value)
}

// Equal returns true if the given option is equal to the instance one. In case
// the given option is nil, we checked the instance one is set to the default
// value of the option.
func (o *CustomNormalizationOption) Equal(o2 *CustomNormalizationOption) bool {
	if o == nil {
		return o2 == nil || reflect.DeepEqual(o2.value, map[string]map[string]string{})
	}
	if o2 == nil {
		return o == nil || reflect.DeepEqual(o.value, map[string]map[string]string{})
	}
	return reflect.DeepEqual(o.value, o2.value)
}

// CustomNormalizationEqual returns true if the two options are equal.
// In case of one option being nil, the value of the other must be nil as well
// or be set to the default value of this option.
func CustomNormalizationEqual(o1, o2 *CustomNormalizationOption) bool {
	return o1.Equal(o2)
}
