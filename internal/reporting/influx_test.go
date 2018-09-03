package reporting

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockTime() time.Time {
	return time.Now()
}

var now = mockTime()
var nowTs = now.UnixNano()

var testDatas = map[*Datapoint]string{
	&Datapoint{
		Measurement: "test_ms1",
		Tags:        map[string]string{"Tag1": "Value1"},
		Fields:      map[string]interface{}{"Field1": 100},
		TimeStamp:   now,
	}: fmt.Sprintf("test_ms1,Tag1=Value1 Field1=100 %d", nowTs),
	&Datapoint{
		Measurement: "test_ms2_no_tags",
		Fields:      map[string]interface{}{"Field1": 100},
		TimeStamp:   now,
	}: fmt.Sprintf("test_ms2_no_tags Field1=100 %d", nowTs),
}

func TestFormatLineProtocol(t *testing.T) {

	for dp, str := range testDatas {
		formatted, err := dp.formatLineProtocol()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, formatted, str)
	}
}

func TestInfluxReporter(t *testing.T) {

	var body []byte
	var err error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	r := &InfluxReporter{Url: ts.URL}

	var out []string
	for dp, str := range testDatas {
		out = append(out, str)
		data, _ := dp.formatLineProtocol()
		r.addToBuffer(data)
	}
	expected := strings.Join(out, "\n")
	r.flush()
	assert.Equal(t, string(body), expected)
}
