package main

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/AlexMickh/proj-user/internal/config"
)

func main() {
	cfg := config.MustLoad()

	file, err := os.Create(".env")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	v := reflect.ValueOf(*cfg)

	processStruct(file, v)
}

func processStruct(w io.Writer, v reflect.Value) {
	t := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)

		if field.Kind() == reflect.Struct {
			processStruct(w, field)
		} else {
			fieldValue := v.Field(i)
			fieldType := t.Field(i)
			tag := fieldType.Tag.Get("env")

			_, err := fmt.Fprintf(w, "%s=%s\n", tag, fmt.Sprint(fieldValue))
			if err != nil {
				panic(err)
			}
		}
	}
}
