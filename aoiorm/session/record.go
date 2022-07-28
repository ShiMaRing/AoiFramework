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
		s.CallMethod(BeforeInsert, value)
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
		s.CallMethod(AfterQuery, dst.Addr().Interface())

		dstSlice.Set(reflect.Append(dstSlice, dst))
	}
	return rows.Close()
}

// support map[string]interface{}
// also support kv list: "Name", "Tom", "Age", 18, ....
func (s *Session) Update(kv ...interface{}) (int64, error) {
	//先进行类型转化
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		//说明是列表形式
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	//开始拼凑sql
	s.cla.Set(clause.UPDATE, s.refTable.Name, m)
	build, i := s.cla.Build(clause.UPDATE, clause.WHERE)
	exec, err := s.Raw(build, i...).Exec()
	if err != nil {
		return 0, err
	}
	return exec.RowsAffected()
}

// Delete records with where clause
func (s *Session) Delete() (int64, error) {
	//只要表名
	s.cla.Set(clause.DELETE, s.refTable.Name)
	build, i := s.cla.Build(clause.DELETE, clause.WHERE)
	exec, err := s.Raw(build, i...).Exec()
	if err != nil {
		return 0, err
	}
	return exec.RowsAffected()
}
func (s *Session) Count() (int64, error) {
	s.cla.Set(clause.COUNT, s.refTable.Name)
	build, i := s.cla.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(build, i...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

// Limit 添加链式调用
func (s *Session) Limit(num int) *Session {
	s.cla.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	s.cla.Set(clause.WHERE, append([]interface{}{desc}, args...)...)
	return s
}

// OrderBy adds order by condition to clause
func (s *Session) OrderBy(desc string) *Session {
	s.cla.Set(clause.ORDERBY, desc)
	return s
}
