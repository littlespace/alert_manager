package template

import "fmt"

var Schema = `
CREATE TABLE IF NOT EXISTS alerts (
  id SERIAL,
  name VARCHAR(128) NOT NULL,
  description TEXT NOT NULL,
  entity VARCHAR(128) NOT NULL,
  external_id VARCHAR(32) NOT NULL,
  source VARCHAR(64) NOT NULL,
  device VARCHAR(64),
  site VARCHAR(64),
  owner VARCHAR(64),
  team VARCHAR(64) NOT NULL,
  tags VARCHAR(64)[] DEFAULT array[]::varchar[],
  start_time BIGINT NOT NULL,
  agg_id INT,
  auto_expire BOOLEAN NOT NULL,
  auto_clear BOOLEAN NOT NULL,
  is_aggregate BOOLEAN NOT NULL,
  expire_after INT,
  severity SMALLINT NOT NULL,
  status SMALLINT NOT NULL,
  labels JSON,
  last_active BIGINT NOT NULL,
  scope VARCHAR(16)
  ) PARTITION BY LIST(team);

CREATE TABLE IF NOT EXISTS suppression_rules (
  id SERIAL,
  name VARCHAR(128) NOT NULL,
  entities JSON NOT NULL,
  mcond SMALLINT  NOT NULL,
  created_at BIGINT NOT NULL,
  duration INT NOT NULL,
  reason TEXT,
  creator varchar(64) NOT NULL,
  team VARCHAR(64) NOT NULL
  ) PARTITION BY LIST(team);

CREATE TABLE IF NOT EXISTS alert_history (
  id SERIAL PRIMARY KEY,
  timestamp BIGINT NOT NULL,
  alert_id INT NOT NULL,
  event TEXT NOT NULL);
`

func Partition(team string) string {
	tmpl := `
    CREATE TABLE IF NOT EXISTS alerts_%[1]s PARTITION OF alerts FOR VALUES IN ('%[1]s');
    CREATE TABLE IF NOT EXISTS suppression_rules_%[1]s PARTITION OF suppression_rules FOR VALUES IN ('%[1]s');
    CREATE INDEX ON alerts (id);
    CREATE INDEX ON suppression_rules (id);
    CREATE INDEX ON alert_history (id, alert_id);
  `
	return fmt.Sprintf(tmpl, team)
}
