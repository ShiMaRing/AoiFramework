package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
}
func genBindVars(num int) string {
	var args []string
	for i := 0; i < num; i++ {
		args = append(args, "?")
	}
	return strings.Join(args, ",")
}

func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _limit(values ...interface{}) (string, []interface{}) {
	return "LIMIT ?", values
}

func _select(values ...interface{}) (string, []interface{}) {
	var tableName = values[0].(string)
	var fields = strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT  %s  FROM %s", fields, tableName), []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var bindStr string
	var sql strings.Builder
	var args []interface{}
	sql.WriteString("VALUES")
	//获取各项参数，参数类型为[]interface
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v)) //所有的长度都是一样的，没必要每次调用
		}
		sql.WriteString(fmt.Sprintf("( %s )", bindStr))
		if i+1 != len(values) {
			//说明还没到最后
			sql.WriteString(",")
		}
		args = append(args, v...)
	}
	return sql.String(), args
}

func _insert(values ...interface{}) (string, []interface{}) {
	//第一个参数表名，第二个参数各个字段切片
	// INSERT INTO $tableName ($fields)
	var tableName = values[0].(string)
	var fields = strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}
