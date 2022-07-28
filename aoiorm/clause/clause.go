package clause

import "strings"

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

func (c *Clause) Set(name Type, vars ...interface{}) {
	//根据传入的名称选择对应的类型生成器，生成后添加到对用的结构中
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	gen := generators[name]
	s, args := gen(vars...) //此处必要展开，不然只有一个元素
	c.sql[name] = s
	c.sqlVars[name] = args
}

func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var args []interface{}
	//orders为排序类型
	for _, order := range orders {
		sqls = append(sqls, c.sql[order])
		args = append(args, c.sqlVars[order]...)
	}

	//完成之后需要进行清空
	defer func() {
		c.sql = nil
		c.sqlVars = nil
	}()

	return strings.Join(sqls, " "), args
}
