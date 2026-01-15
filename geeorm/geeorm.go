package geeorm

import (
	"database/sql"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"

	_ "github.com/mattn/go-sqlite3"
)

type Engine struct {
	db *sql.DB
	dialect dialect.Dialect
}

type TxFunc func(*session.Session) (interface{}, error)

// 用户只需要将所有的操作放到一个回调函数里, 作为入参传递给engine.Transaction()
// 发生任何错误, 自动回滚, 如果没有错误发生, 则提交
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.RollBack()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.RollBack() // err is non-nil; don't change it
		} else {
			err = s.Commit() // err is nil; if Commit returns update err
		}
	}()

	return f(s)
}

// NewEngine:
// 1. 连接数据库 返回 *sql.db
// 2. 调用db.Ping() 检查数据库是否能够正常连接
func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	// Send a ping to make sure the databases connection is alive.
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s not found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// 通过Engine实例创建会话 进而与数据库进行交互
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}
