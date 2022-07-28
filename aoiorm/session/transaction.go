package session

import "AoiFramework/aoiorm/olog"

//封装开始结束回滚

func (s *Session) Begin() (err error) {
	olog.Info("transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		olog.Error(err)
		return err
	}
	return nil
}

func (s *Session) Commit() (err error) {
	olog.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		olog.Error(err)
		return err
	}
	return
}

func (s *Session) Rollback() (err error) {
	olog.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		olog.Error(err)
	}
	return
}
