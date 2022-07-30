package aoirpc

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

//利用反射注册服务

//代指对应的相关方法
type methodType struct {
	method    reflect.Method //方法名等类型
	ArgType   reflect.Type   //传入参数类型
	ReplyType reflect.Type   //返回值类型
	numCalls  uint64         //调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

//根据方法参数类型创建对应的实例对象
func (m *methodType) newArgv() reflect.Value {
	// arg may be a pointer type, or a value type
	//传入的第一个参数可能是指针,或者是值
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Pointer {
		//获取指针指向的实例对象的value
		argv = reflect.New(m.ArgType.Elem()) //获取的也是指针类型指向创建的空值
	} else {
		argv = reflect.New(m.ArgType).Elem() //获取的是根据类型创建的值类型
	}
	return argv
}

//返回对应的返回值类型实例
func (m *methodType) newReplyv() reflect.Value {
	//一定是指针类型
	//replyv是一个指针类型的value，指向一个m.ArgType.Elem()实例
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

//服务实例
type service struct {
	name     string                 //服务的类型名称
	typ      reflect.Type           //类型
	receiver reflect.Value          //接收者实例
	methods  map[string]*methodType //包含的函数集
}

func newService(rcvr interface{}) *service {
	s := new(service)
	s.receiver = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.receiver).Type().Name() //不论是否是指针
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	//得到s注册的所有方法
	s.methods = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		//要对方法进行判断
		t := method.Type
		//获取函数类型
		if t.NumIn() != 3 || t.NumOut() != 1 {
			continue
		}
		if t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			fmt.Println(t.Out(0))
			fmt.Println(reflect.TypeOf((*error)(nil)))
			continue
		}

		argType, rplyType := t.In(1), t.In(2)

		if !ast.IsExported(argType.Name()) && argType.PkgPath() != "" {
			continue
		}

		if !ast.IsExported(rplyType.Name()) && rplyType.PkgPath() != "" {
			continue
		}

		//必须要是指针
		if rplyType.Kind() != reflect.Pointer {
			continue
		}

		s.methods[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: rplyType,
		}

		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	//获取函数
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.receiver, argv, replyv})

	if err := returnValues[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}
