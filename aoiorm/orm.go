package aoiorm

import (
	"AoiFramework/aoiorm/dialect"
	"AoiFramework/aoiorm/olog"
	"AoiFramework/aoiorm/session"
	"database/sql"
	"fmt"
	"strings"
)

//提供交互接口
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		olog.Error(err)
		return
	}
	err = db.Ping()
	if err != nil {
		olog.Error(err)
		return
	}
	getDialect, b := dialect.GetDialect(driver)
	if !b {
		olog.Errorf("dialect %s is not founded ", driver)
		return
	}
	olog.Info("connect database success!!")
	return &Engine{db: db, dialect: getDialect}, nil
}

func (engine *Engine) Close() {
	err := engine.db.Close()
	if err != nil {
		olog.Error("close database fail")
		return
	}
	olog.Info("close database success")

}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

type TxFunc func(session *session.Session) (result interface{}, err error)

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {

	newSession := engine.NewSession()
	if err = newSession.Begin(); err != nil {
		olog.Error(err)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			newSession.Rollback()
			panic(p)
		} else if err != nil {
			newSession.Rollback()
		} else {
			newSession.Commit()
		}
	}()
	return f(newSession)
}

//difference 返回 a-b
func difference(a, b []string) (diff []string) {
	mapB := make(map[string]struct{})
	for _, v := range b {
		mapB[v] = struct{}{}
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

// Migrate 数据库迁移操作，将原来的表迁移至新的映射表
func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			//如果没有的话，直接创建即可
			olog.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		//检查表名与当前模型字段
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		colums, _ := rows.Columns()
		addCols := difference(table.FieldNames, colums)
		delCols := difference(colums, table.FieldNames)

		olog.Infof("added cols %v, deleted cols %v", addCols, delCols)
		for _, col := range addCols {
			field := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, col, field.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}
		//没有需要删除的字段
		if len(delCols) == 0 {
			return
		}

		tmp := "tmp_" + table.Name

		fieldStr := strings.Join(table.FieldNames, ", ")

		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))

		_, err = s.Exec()
		return
	})
	return err
}
