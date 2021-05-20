package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cortexproject/cortex/pkg/cortex"
)

func main() {
	x := cortex.Config{}
	traverse(reflect.TypeOf(x), 0)
}

func traverse(t reflect.Type, depth int) {

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Tag != "" && f.Tag.Get("yaml") != "-" {

				var typ string = f.Type.Name()

				// Get underlying type of the slice.
				if f.Type.Kind() == reflect.Slice {
					// TODO: handle ptr and struct
					typ = "[]" + f.Type.Elem().Kind().String()
				}

				// Convert float64 to json.Number because Kubernetes API
				// conventions doesn't allow floats.
				if typ == "float64" {
					typ = "json.Number"
				}
				// Parse urls as strings.
				// TODO: add validation here
				if typ == "URLValue" {
					typ = "string"
				}
				// Duration type requires the package name.
				if typ == "Duration" {
					typ = "time.Duration"
				}

				fmt.Printf("%s %-40s %-30s `json:\"%s\"`\n", strings.Repeat("  ", depth+1), f.Name, typ, f.Tag.Get("yaml"))
				traverse(f.Type, depth+1)
			}
		}
	}
}
