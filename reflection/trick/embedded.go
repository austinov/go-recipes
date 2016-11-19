package trick

import "reflect"

// some example types to demonstrate
// extracting embedded type from struct.
type (
	Response struct {
		Code int
		Desc string
	}

	Message struct {
		Response
		Text string
	}

	People struct {
		Name     string
		Birthday string
	}
)

// getEmbedded returns embedded struct Response in any interface if it is.
// It's useful when we have few types with anonymous field
// and we need to extract it.
func GetEmbedded(v interface{}) (Response, bool) {
	if vals, ok := getValue(v); !ok {
		return Response{}, false
	} else {
		return extract(vals)
	}
}

// getValue returns reflec.Value from interface if it is a struct.
func getValue(s interface{}) (reflect.Value, bool) {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	return v, true
}

// extract returns anonymous field Response from value if it has.
func extract(v reflect.Value) (Response, bool) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous && field.Name == "Response" {
			iface := v.FieldByName(field.Name).Interface()
			r, ok := iface.(Response)
			return r, ok
		}
	}
	return Response{}, false
}
