package schema

import (
	"AoiFramework/aoiorm/dialect"
	"go/ast"
	"reflect"
)

// Field 代表某一列
type Field struct {
	Name string
	Type string
	Tag  string //约束条件
}

// Schema 代表一张表
type Schema struct {
	Module     interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	//先获取传入参数的类型数据
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Module:   dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}
	//根据类型进行生成
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		//获取各个字段，要求P不是嵌套结构体并且是暴露的
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DtaTypeof(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("aoiorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, field.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	//获取目的dst的value对象
	indirect := reflect.Indirect(reflect.ValueOf(dest))
	var args []interface{}
	//根据保存的内容获取各个元素
	for _, field := range schema.Fields { //从保存的模式中读取数据
		args = append(args, indirect.FieldByName(field.Name).Interface())
	}
	return args
}
