package aoiorm

import (
	"AoiFramework/aoiorm/dialect"
	"AoiFramework/aoiorm/olog"
	"AoiFramework/aoiorm/session"
	"database/sql"
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
		olog.Error("dialect %s is not founded ", driver)
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
