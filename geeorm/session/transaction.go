package session

import "geeorm/log"

// 封装可以统一打印日志，方便定位问题
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	// 调用s.db.Begin()得到*sql.Tx对象，赋值给s.tx
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) RollBack() (err error) {
	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	return
}