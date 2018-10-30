package models

import (
	"github.com/golang/glog"
	"time"
)

var (
	QueryInsertNewRecord = `INSERT INTO alert_history (
		alert_id, timestamp, event
	) VALUES (:alert_id, :timestamp, :event) RETURNING id`

	QueryAlertHistory = "SELECT * from alert_history WHERE alert_id IN (?) ORDER BY alert_id, id"
)

type Record struct {
	Id        int64
	AlertId   int64 `db:"alert_id"`
	Timestamp MyTime
	Event     string
}

func NewRecord(alertId int64, event string) *Record {
	return &Record{
		AlertId: alertId, Event: event, Timestamp: MyTime{time.Now()},
	}
}

func (tx *Tx) NewRecord(alertId int64, event string) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertNewRecord)
	if err != nil {
		glog.Errorf("Failed to create record: %v", err)
		return newId, err
	}
	err = stmt.Get(&newId, NewRecord(alertId, event))
	if err != nil {
		glog.Errorf("Failed to create record: %v", err)
	}
	return newId, err
}

func (tx *Tx) AddAlertHistory(alerts Alerts) error {
	var allRecords []*Record
	var ids []int64
	idsToAlert := make(map[int64]*Alert)
	for _, a := range alerts {
		ids = append(ids, a.Id)
		idsToAlert[a.Id] = a
	}
	if err := tx.InSelect(QueryAlertHistory, &allRecords, ids); err != nil {
		return err
	}
	for _, rec := range allRecords {
		idsToAlert[rec.AlertId].History = append(idsToAlert[rec.AlertId].History, rec)
	}
	return nil
}
