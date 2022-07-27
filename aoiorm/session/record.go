package session

import (
	"AoiFramework/aoiorm/clause"
	"reflect"
)

//增删查改

// Insert 增加方法，参数为各个结构体指针
func (s *Session) Insert(values ...interface{}) (int64, error) {
	var args []interface{}
	for _, value := range values {
		table := s.Model(value).refTable //获取表数据，利用获得的field反射提取字段数据
		s.cla.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues := table.RecordValues(value)
		args = append(args, recordValues) //提取出的数据设置到args中
	}
	//开始拼凑sql
	s.cla.Set(clause.VALUES, args...)
	build, i := s.cla.Build(clause.INSERT, clause.VALUES)
	exec, err := s.Raw(build, i...).Exec()
	if err != nil {
		return 0, err
	}
	return exec.RowsAffected()
}

func (s *Session) Find(values interface{}) error {
	//传入的是一个切片地址
	dstSlice := reflect.Indirect(reflect.ValueOf(values))
	dstType := dstSlice.Type().Elem() //容器的elem方法返回具体的类型
	tmp := reflect.New(dstType).Elem().Interface()
	table := s.Model(tmp).refTable //获取table

	//传入参数，构筑sql语句
	s.cla.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.cla.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	//传入参数进行查询
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	for rows.Next() {
		//创建实例
		dst := reflect.New(dstType).Elem()
		var values []interface{}
		//根据得到的列名称进行选择
		for _, name := range table.FieldNames {
			values = append(values, dst.FieldByName(name).Addr().Interface())
		}
		//得到每个字段关联的实例对象
		if err := rows.Scan(values...); err != nil {
			return err
		}
		dstSlice.Set(reflect.Append(dstSlice, dst))
	}
	return rows.Close()
}
