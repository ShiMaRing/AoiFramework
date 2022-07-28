package session

import (
	"AoiFramework/aoiorm/clause"
	"AoiFramework/aoiorm/dialect"
	"AoiFramework/aoiorm/olog"
	"AoiFramework/aoiorm/schema"
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Session struct {
	db *sql.DB

	sql  strings.Builder
	args []interface{}

	dia      dialect.Dialect
	refTable *schema.Schema

	cla clause.Clause

	tx *sql.Tx
}

// New 传入翻译器和数据源
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:  db,
		dia: dialect,
	}
}

func (session *Session) Clear() {
	session.sql.Reset()
	session.args = []interface{}{}
}

func (s *Session) DB() CommonDB {
	if s.tx == nil {
		return s.db
	}
	return s.tx
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.args = append(s.args, values...)
	return s
}

func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	olog.Info(s.sql.String(), s.args)

	if result, err = s.DB().Exec(s.sql.String(), s.args...); err != nil {
		olog.Error(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	olog.Info(s.sql.String(), s.args)
	return s.DB().QueryRow(s.sql.String(), s.args...)
}

func (s *Session) QueryRows() (*sql.Rows, error) {
	defer s.Clear()
	olog.Info(s.sql.String(), s.args)
	query, err := s.DB().Query(s.sql.String(), s.args...)
	if err != nil {
		olog.Error(err)
	}
	return query, err
}

func (s *Session) First(value interface{}) error {

	indirect := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(indirect.Type())).Elem() //生成切片元素
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}

	indirect.Set(destSlice.Index(0))
	return nil
}
