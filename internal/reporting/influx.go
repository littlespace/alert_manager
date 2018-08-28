package reporting

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Datapoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	TimeStamp   time.Time
}

var DataChan = make(chan *Datapoint)

func (d *Datapoint) formatLineProtocol() (string, error) {
	var tags, fields []string
	for tk, tv := range d.Tags {
		tv = strings.Replace(tv, " ", "_", -1)
		tags = append(tags, fmt.Sprintf("%s=%s", tk, tv))
	}
	for fk, fv := range d.Fields {
		switch fv.(type) {
		case int:
			fv = fv.(int)
		case int64:
			fv = fv.(int64)
		case float64:
			fv = fv.(float64)
		default:
			return "", fmt.Errorf("Cant format field %s: Need numeric value", fk)
		}
		fields = append(fields, fmt.Sprintf("%s=%v", fk, fv))
	}

	if len(d.Tags) == 0 {
		return fmt.Sprintf("%s %s %v", d.Measurement, strings.Join(fields, ","), d.TimeStamp.UnixNano()), nil
	}
	return fmt.Sprintf("%s,%s %s %v",
		d.Measurement,
		strings.Join(tags, ","),
		strings.Join(fields, ","),
		d.TimeStamp.UnixNano()), nil
}

type InfluxReporter struct {
	Url           string
	FlushInterval time.Duration `mapstructure:"flush_interval"`

	buffer []string
	sync.Mutex
}

func (n *InfluxReporter) addToBuffer(data string) {
	n.Lock()
	defer n.Unlock()
	n.buffer = append(n.buffer, data)
}

func (n *InfluxReporter) flush() {
	n.Lock()
	defer n.Unlock()
	if len(n.buffer) == 0 {
		return
	}
	data := strings.Join(n.buffer, "\n")
	if n.Url == "stdout" {
		fmt.Println(data)
		n.buffer = n.buffer[:0]
		return
	}
	resp, err := http.Post(n.Url, "binary/octet-stream", bytes.NewBuffer([]byte(data)))
	if err != nil {
		//n.statsPostError.Add(1)
		glog.Errorf("Output: Unable to post to influx: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		//n.statsPostError.Add(1)
		glog.Errorf("Output: Unable to post to influx: Got HTTP %d", resp.StatusCode)
	}
	//n.statPostsSent.Add(1)
	n.buffer = n.buffer[:0]
}

func (n *InfluxReporter) handleFlush() {
	t := time.NewTicker(n.FlushInterval)
	for {
		<-t.C
		n.flush()
	}
}

func (n *InfluxReporter) Start(ctx context.Context) {
	go n.handleFlush()
	for {
		select {
		case datapoint := <-DataChan:
			data, err := datapoint.formatLineProtocol()
			if err != nil {
				glog.Errorf("%v", err)
				break
			}
			n.addToBuffer(data)
		case <-ctx.Done():
			return
		}
	}
}
