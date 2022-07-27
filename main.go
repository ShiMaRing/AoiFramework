package main

import (
	"fmt"
	"reflect"
)

type A struct {
	name string
	age  int
	sex  bool
}

func main() {
	var a = A{
		name: "hello",
	}
	hello(a)
}

func hello(a interface{}) {
	t := reflect.Indirect(reflect.ValueOf(a)).Type()
	fmt.Println(t.NumField())
}

func hello2(a interface{}) {
	t := reflect.ValueOf(a).Elem().Type()
	fmt.Println(t.NumField())
}
