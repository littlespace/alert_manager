package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	tpl "github.com/mayuresh82/alert_manager/template"
)

// custom structs to allow for mocking
type Dbase interface {
	NewTx() Txn
	Close() error
}

type DB struct {
	*sqlx.DB
}

func (d *DB) NewTx() Txn {
	tx := d.DB.MustBegin()
	return &Tx{tx}
}

func NewDB(addr, username, password, dbName string, timeout int) Dbase {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		glog.Fatalf("Invalid DB addr: %s", addr)
	}
	if host == "" {
		host = "localhost"
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%d sslmode=disable lock_timeout=15000", host, port, username, password, dbName, timeout)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		glog.Fatalf("Cant open DB: %v", err)
	}
	db.MustExec(tpl.Schema)
	return &DB{db}
}

func NewPartition(team string) string {
	tmpl := `
    CREATE TABLE IF NOT EXISTS alerts_%[1]s PARTITION OF alerts FOR VALUES IN ('%[1]s');
  `
	return fmt.Sprintf(tmpl, team)
}

type Txn interface {
	InQuery(query string, arg ...interface{}) error
	InSelect(query string, to interface{}, arg ...interface{}) error
	UpdateAlert(alert *Alert) error
	NewInsert(query string, item interface{}) (int64, error)
	GetAlert(query string, args ...interface{}) (*Alert, error)
	SelectAlerts(query string, args ...interface{}) (Alerts, error)
	SelectAlertsWithHistory(query string, args ...interface{}) (Alerts, error)
	AddAlertHistory(alerts Alerts) error
	SelectRules(query string, args ...interface{}) (SuppRules, error)
	NewRecord(alertId int64, event string) (int64, error)
	SelectTeams(query string, args ...interface{}) (Teams, error)
	SelectUsers(query string, args ...interface{}) (Users, error)
	Rollback() error
	Commit() error
	Exec(query string, args ...interface{}) error
}

type Tx struct {
	*sqlx.Tx
}

func (tx *Tx) InQuery(query string, arg ...interface{}) error {
	query, args, err := sqlx.In(query, arg...)
	if err != nil {
		return err
	}
	query = tx.Rebind(query)
	return tx.Exec(query, args...)
}

func (tx *Tx) InSelect(query string, to interface{}, arg ...interface{}) error {
	query, args, err := sqlx.In(query, arg...)
	if err != nil {
		return err
	}
	query = tx.Rebind(query)
	return tx.Select(to, query, args...)
}

func (tx *Tx) Exec(query string, args ...interface{}) error {
	_, err := tx.Tx.Exec(query, args...)
	return err
}

// WithTx wraps a transaction around a function call.
func WithTx(ctx context.Context, tx Txn, cb func(ctx context.Context, tx Txn) error) error {
	err := cb(ctx, tx)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}

func (tx *Tx) NewInsert(query string, item interface{}) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return newId, err
	}
	err = stmt.Get(&newId, item)
	return newId, err
}

type MyTime struct {
	time.Time
}

func (t MyTime) Value() (driver.Value, error) {
	return driver.Value(t.Unix()), nil
}

func (t *MyTime) Scan(src interface{}) error {
	ns := sql.NullInt64{}
	if err := ns.Scan(src); err != nil {
		return err
	}

	if !ns.Valid {
		return fmt.Errorf("MyTime.Scan: column is not nullable")
	}
	*t = MyTime{time.Unix(ns.Int64, 0)}
	return nil
}
