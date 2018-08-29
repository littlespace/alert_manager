package parsers

import (
	"github.com/mayuresh82/alert_manager/listener"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testDatas = map[string]struct {
	raw    string
	parsed *listener.WebHookAlertData
}{
	"grafana": {
		raw: `{
      "title": "My First Test Alert",
      "ruleId": 1,
      "ruleName": "Neteng BB Input Errors Test",
      "ruleUrl": "http://url.to.grafana/db/dashboard/my_dashboard?panelId=2",
      "state": "alerting",
      "imageUrl": "http://s3.image.url",
      "message": "A BB Link is experiencing > 1000pps input errors over 15 minutes!",
      "evalMatches": [
        {
          "metric": "input_errors",
          "tags": {"device": "br1-sjc1", "interface": "et-0/0/3:0"},
          "value": 1222
        }
      ]
    }`,
		parsed: &listener.WebHookAlertData{
			Id:      "1",
			Name:    "Neteng BB Input Errors Test",
			Details: "A BB Link is experiencing > 1000pps input errors over 15 minutes!\nMetric: input_errors, Value: 1222\n",
			Device:  "br1-sjc1",
			Entity:  "et-0/0/3:0",
			Status:  listener.Status_ALERTING,
			Source:  "grafana",
		},
	},
	"observium": {
		raw: `{"ALERT_STATE":"ALERTING","ALERT_URL":"http:blah","ALERT_UNIXTIME":1535139630,"ALERT_TIMESTAMP":"2018-08-24 19:40:30 +00:00","ALERT_TIMESTAMP_RFC2822":"Fri, 24 Aug 2018 19:40:30 +0000","ALERT_TIMESTAMP_RFC3339":"2018-08-24T19:40:30+00:00","ALERT_ID":"106678","ALERT_MESSAGE":"Neteng BGP Down","CONDITIONS":"bgpPeerState notequals established (idle)","METRICS":"bgpPeerState = idle","DURATION":"2m 11s (2018-08-24 19:38:19)","ENTITY_LINK":"<a href=\"http:\/\/rp-lutil9.roblox.local\/device\/device=529\/tab=routing\/proto=bgp\/\" class=\"entity-popup \" data-eid=\"4278\" data-etype=\"bgp_peer\">AS65101 10.130.249.121<\/a>","ENTITY_NAME":"AS65101 10.130.249.121","ENTITY_ID":"4278","ENTITY_TYPE":"bgp_peer","ENTITY_DESCRIPTION":"","DEVICE_HOSTNAME":"bs1-ash1","DEVICE_SYSNAME":"bs1-ash1","DEVICE_ID":"627","DEVICE_LINK":"<a href=\"http:\/\/rp-lutil9.roblox.local\/device\/device=627\/\" class=\"entity-popup \" data-eid=\"627\" data-etype=\"device\">bs1-ash1<\/a>","DEVICE_HARDWARE":"QFX10002-72Q","DEVICE_OS":"Juniper JunOS 17.3R2-S2.1 (Ethernet Switch)","DEVICE_LOCATION":"44274 Round Table Plaza, Building L Suite 112, Ashburn, VA 20147","DEVICE_UPTIME":"56 days, 22h 26m 23s","DEVICE_REBOOTED":"2018-06-28 21:13:44","TITLE":"ALERT: [bs1-ash1] [bgp_peer] [AS65101 10.130.249.121] Neteng BGP Down"}`,
		parsed: &listener.WebHookAlertData{
			Id:      "106678",
			Name:    "Neteng BGP Down",
			Details: "ALERT: [bs1-ash1] [bgp_peer] [AS65101 10.130.249.121] Neteng BGP Down Url: http:blah",
			Device:  "bs1-ash1",
			Entity:  "AS65101 10.130.249.121",
			Status:  listener.Status_ALERTING,
			Source:  "observium",
		},
	},
}

func TestParsing(t *testing.T) {

	for name, data := range testDatas {
		var parser listener.Parser
		switch name {
		case "grafana":
			parser = &GrafanaParser{name: "grafana"}
		case "observium":
			parser = &ObserviumParser{name: "observium"}
		case "kapacitor":
			parser = &KapacitorParser{name: "kapacitor"}
		}
		result, err := parser.Parse([]byte(data.raw))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, result.Id, data.parsed.Id)
		assert.Equal(t, result.Name, data.parsed.Name)
		assert.Equal(t, result.Details, data.parsed.Details)
		assert.Equal(t, result.Device, data.parsed.Device)
		assert.Equal(t, result.Entity, data.parsed.Entity)
		assert.Equal(t, result.Status, data.parsed.Status)
		assert.Equal(t, result.Source, data.parsed.Source)
	}
}
