package di

import (
	"fmt"
	"reflect"
)

// TypeInstanceToName converts an instance of a type to a unique name.
func TypeInstanceToName(v interface{}) string {
	t := reflect.TypeOf(v)

	if name := t.Name(); name != "" {
		// non-interface types
		return fmt.Sprintf("%s.%s", t.PkgPath(), name)
	}

	// interface types
	e := t.Elem()
	return fmt.Sprintf("%s.%s", e.PkgPath(), e.Name())
}
