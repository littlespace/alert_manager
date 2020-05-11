package parsers

import (
	"testing"

	"github.com/mayuresh82/alert_manager/listener"
	"github.com/stretchr/testify/assert"
)

var testDatas = map[string]struct {
	raw    string
	parsed []*listener.WebHookAlert
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
		parsed: []*listener.WebHookAlert{&listener.WebHookAlert{
			Id:      "1",
			Name:    "Neteng BB Input Errors Test",
			Details: "A BB Link is experiencing > 1000pps input errors over 15 minutes!\nMetric: input_errors, Value: 1222\n",
			Device:  "br1-sjc1",
			Entity:  "et-0/0/3:0",
			Status:  listener.Status_ALERTING,
			Source:  "grafana",
		}},
	},
	"observium": {
		raw: `{"ALERT_STATE":"ALERTING","ALERT_URL":"http:blah","ALERT_UNIXTIME":1535139630,"ALERT_TIMESTAMP":"2018-08-24 19:40:30 +00:00","ALERT_TIMESTAMP_RFC2822":"Fri, 24 Aug 2018 19:40:30 +0000","ALERT_TIMESTAMP_RFC3339":"2018-08-24T19:40:30+00:00","ALERT_ID":"106678","ALERT_MESSAGE":"Neteng BGP Down","CONDITIONS":"bgpPeerState notequals established (idle)","METRICS":"bgpPeerState = idle","DURATION":"2m 11s (2018-08-24 19:38:19)","ENTITY_LINK":"<a href=\"http:\/\/rp-lutil9.roblox.local\/device\/device=529\/tab=routing\/proto=bgp\/\" class=\"entity-popup \" data-eid=\"4278\" data-etype=\"bgp_peer\">AS65101 10.130.249.121<\/a>","ENTITY_NAME":"AS65101 10.130.249.121","ENTITY_ID":"4278","ENTITY_TYPE":"bgp_peer","ENTITY_DESCRIPTION":"Provider X","DEVICE_HOSTNAME":"bs1-ash1","DEVICE_SYSNAME":"bs1-ash1","DEVICE_ID":"627","DEVICE_LINK":"<a href=\"http:\/\/rp-lutil9.roblox.local\/device\/device=627\/\" class=\"entity-popup \" data-eid=\"627\" data-etype=\"device\">bs1-ash1<\/a>","DEVICE_HARDWARE":"QFX10002-72Q","DEVICE_OS":"Juniper JunOS 17.3R2-S2.1 (Ethernet Switch)","DEVICE_LOCATION":"44274 Round Table Plaza, Building L Suite 112, Ashburn, VA 20147","DEVICE_UPTIME":"56 days, 22h 26m 23s","DEVICE_REBOOTED":"2018-06-28 21:13:44","TITLE":"ALERT: [bs1-ash1] [bgp_peer] [AS65101 10.130.249.121] Neteng BGP Down"}`,
		parsed: []*listener.WebHookAlert{&listener.WebHookAlert{
			Id:      "106678",
			Name:    "Neteng BGP Down",
			Details: "ALERT: [bs1-ash1] [bgp_peer] [AS65101 10.130.249.121] Neteng BGP Down / Provider X",
			Device:  "bs1-ash1",
			Entity:  "AS65101 10.130.249.121",
			Status:  listener.Status_ALERTING,
			Source:  "observium",
		}},
	},
	"kapacitor": {
		raw: `{"id":"Neteng Transit Util Out","message":"br2-lhr1:et-0/0/9:1","details":"Transit Util exceeds 0.4265067984","time":"2018-11-08T00:20:00Z","duration":0,"level":"WARNING","data":{"series":[{"name":"jnpr_interface_stat","tags":{"device":"br2-lhr1","entity":"et-0/0/9:1"},"columns":["time","sigma","stat"],"values":[["2018-11-08T00:20:00Z",0,0.4265067984]]}]},"previousLevel":"WARNING"}`,
		parsed: []*listener.WebHookAlert{&listener.WebHookAlert{
			Id:      "",
			Name:    "Neteng Transit Util Out",
			Details: "br2-lhr1:et-0/0/9:1\nTransit Util exceeds 0.4265067984",
			Device:  "br2-lhr1",
			Entity:  "et-0/0/9:1",
			Status:  listener.Status_ALERTING,
			Source:  "kapacitor",
		}},
	},
	"ns1": {
		raw: `{"job": {"notify_delay": 0,"job_type": "tcp","last_log_removal": 1548139477,"frequency": 20, "rapid_recheck": false, "region_scope": "fixed","id": "5c46bbc1a632f60001fwerwer", "notify_repeat": 0,"notify_regional": false,"regions": ["ams","dal","lga","sin","sjc"],"policy": "all","config": {"host": "1.6.7.35","port": 443},"status": {},"notify_failback": true,"rules": [],"v2": true,"active": true,"name": "monitornames","notes": null},"region": "global","since": 1548139477,"state": "down"}`,
		parsed: []*listener.WebHookAlert{&listener.WebHookAlert{
			Id:      "5c46bbc1a632f60001fwerwer",
			Name:    "Neteng DNS Monitor Down",
			Details: "monitornames",
			Entity:  "monitornames",
			Status:  listener.Status_ALERTING,
			Source:  "ns1",
		}},
	},
	"prom": {
		raw: `{
			"status": "firing",
			"alerts": [
			  {
				"status": "firing",
				"labels": {
				  "alertname": "Device down",
				  "description": "Device foo is down",
				  "device": "foo",
				  "entity": "foo",
				  "role": "rack-switch",
				  "severity": "warning",
				  "site": "ord1",
				  "source": "prometheus"
				},
				"annotations": {},
				"startsAt": "2019-04-02T22:28:12.208437528Z",
				"endsAt": "0001-01-01T00:00:00Z"
			  },
			  {
				"status": "firing",
				"labels": {
				  "alertname": "Device down",
				  "description": "Device bar is down",
				  "device": "bar",
				  "entity": "bar",
				  "role": "rack-switch",
				  "severity": "warning",
				  "site": "ord1",
				  "source": "prometheus"
				},
				"annotations": {},
				"startsAt": "2019-04-02T22:28:12.208437528Z",
				"endsAt": "0001-01-01T00:00:00Z"
			  }
			],
			"commonLabels": {
				"alertname": "Device down",
				"role": "rack-switch",
				"severity": "warning",
				"site": "ord1",
				"source": "prometheus"
			},
			"externalURL": "http://8fd5083ae377:9093",
			"version": "4"
		  }`,
		parsed: []*listener.WebHookAlert{
			&listener.WebHookAlert{
				Id:      "None",
				Name:    "Device down",
				Details: "Device foo is down",
				Device:  "foo",
				Entity:  "foo",
				Status:  listener.Status_ALERTING,
				Source:  "prometheus",
			},
			&listener.WebHookAlert{
				Id:      "None",
				Name:    "Device down",
				Details: "Device bar is down",
				Device:  "bar",
				Entity:  "bar",
				Status:  listener.Status_ALERTING,
				Source:  "prometheus",
			}},
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
		case "ns1":
			parser = &Ns1Parser{name: "ns1"}
		case "prom":
			parser = &PromParser{name: "prom"}
		}
		result, err := parser.Parse([]byte(data.raw))
		if err != nil {
			t.Fatal(err)
		}
		for _, alertData := range result.Alerts {
			var match bool
			for _, parsed := range data.parsed {
				if alertData.Device == parsed.Device {
					match = true
					assert.Equal(t, alertData.Id, parsed.Id)
					assert.Equal(t, alertData.Name, parsed.Name)
					assert.Equal(t, alertData.Details, parsed.Details)
					assert.Equal(t, alertData.Device, parsed.Device)
					assert.Equal(t, alertData.Entity, parsed.Entity)
					assert.Equal(t, alertData.Status, parsed.Status)
					assert.Equal(t, alertData.Source, parsed.Source)
				}
			}
			assert.Equal(t, match, true)
		}

	}
}

func TestParsingGeneric(t *testing.T) {
	raw := `{"id": 1, "name": "Generic JSON alert", "entity": "ent1", "device": "dev1", "description": "its down", "timestamp": "2018-12-04T14:57:34-06:00", "severity": "info", "status": "alerting"}`
	parser := &GenericParser{name: "generic"}
	result, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	alertData := result.Alerts[0]
	assert.Equal(t, alertData.Id, "1")
	assert.Equal(t, alertData.Name, "Generic JSON alert")
	assert.Equal(t, alertData.Entity, "ent1")
	assert.Equal(t, alertData.Device, "dev1")
	assert.Equal(t, alertData.Details, "its down")
	assert.Equal(t, alertData.Status, listener.Status_ALERTING)
	assert.Equal(t, alertData.Level, "INFO")

	raw = `{"id": 1, "name": "Generic JSON alert", "entity": "ent1"}`
	result, err = parser.Parse([]byte(raw))
	assert.Error(t, err)
}
