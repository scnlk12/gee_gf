package geeorm

import (
	"database/sql"
	"fmt"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Engine struct {
	db      *sql.DB
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

// difference returns a - b
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// Migrate table
func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			log.Info("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("select * from %s limit 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		// difference用于计算前后两个字段的差集
		// 新表 - 旧表 = 新增字段
		// 旧表 - 新表 = 删除字段
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, delete cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("alter table %s add column %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("create table %s as select %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("drop table %s;", table.Name))
		s.Raw(fmt.Sprintf("alter table %s rename to %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
