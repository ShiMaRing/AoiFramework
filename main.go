package main

import (
	"fmt"
	"reflect"
)

type A struct {
	name string
	Age  int
	sex  bool
}

func main() {
	fmt.Println(difference(nil, nil) == nil)
	/*var a = A{
		name: "hello",
		Age:  102,
		sex:  true,
	}
	hello(a)*/
}

func hello(a interface{}) {
	t := reflect.Indirect(reflect.ValueOf(a))
	fmt.Printf("%s \n", t.FieldByName("name"))
	ageValue := t.FieldByName("Age")
	fmt.Println(ageValue.Interface().(int))
}

func hello2(a interface{}) {
	t := reflect.ValueOf(a).Elem().Type()
	fmt.Println(t.NumField())
}

//difference 返回 a-b
func difference(a, b []string) (diff []string) {

	return
}
