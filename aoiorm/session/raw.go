package session

import (
	"AoiFramework/aoiorm/dialect"
	"AoiFramework/aoiorm/olog"
	"AoiFramework/aoiorm/schema"
	"database/sql"
	"strings"
)

type Session struct {
	db   *sql.DB
	sql  strings.Builder
	args []interface{}

	dia      dialect.Dialect
	refTable *schema.Schema
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
}

func (session *Session) DB() *sql.DB {
	return session.db
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

	if result, err = s.db.Exec(s.sql.String(), s.args...); err != nil {
		olog.Error(err)
	}
	return
}
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	olog.Info(s.sql.String(), s.args)
	return s.db.QueryRow(s.sql.String(), s.args...)
}

func (s *Session) QueryRows() (*sql.Rows, error) {
	defer s.Clear()
	olog.Info(s.sql.String(), s.args)
	query, err := s.db.Query(s.sql.String(), s.args...)
	if err != nil {
		olog.Error(err)
	}
	return query, err
}
