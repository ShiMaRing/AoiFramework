package session

import (
	"AoiFramework/aoiorm/olog"
	"AoiFramework/aoiorm/schema"
	"fmt"
	"reflect"
	"strings"
)

// Model 操作的是一张表
func (s *Session) Model(value interface{}) *Session {
	//为空或者操作的元素不相等
	if s.refTable == nil || reflect.ValueOf(value) != reflect.ValueOf(s.refTable.Module) {
		s.refTable = schema.Parse(value, s.dia)
	}
	return s
}

func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		olog.Error("Model is not set")
	}
	return s.refTable
}

/*数据库表的创建、删除和判断是否存在的功能。三个方法的实现逻辑是相似的，
利用 RefTable() 返回的数据库表和字段的信息，拼接出 SQL 语句，调用原生 SQL 接口执行。*/

// CreateTable 创建表
func (s *Session) CreateTable() error {
	table := s.refTable
	var columns []string //拼凑
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	//创建sql
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("Create table %s (%s);", table.Name, desc)).Exec()
	return err
}

//删除表
func (s *Session) DropTable() (err error) {
	_, err = s.Raw(fmt.Sprintf("drop table if exists %s", s.refTable.Name)).Exec()
	return
}

//检查是否存在表
func (s *Session) HasTable() bool {
	sql, values := s.dia.TableExistSQL(s.refTable.Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.refTable.Name //检查表明

}
