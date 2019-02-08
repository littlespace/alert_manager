package template

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
  id SERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  entities JSON NOT NULL,
  mcond SMALLINT  NOT NULL,
  created_at BIGINT NOT NULL,
  duration INT NOT NULL,
  reason TEXT,
  creator varchar(64) NOT NULL);

CREATE TABLE IF NOT EXISTS alert_history (
  id SERIAL PRIMARY KEY,
  timestamp BIGINT NOT NULL,
  alert_id INT NOT NULL,
  event TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS teams (
  id SERIAL PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  organization VARCHAR(64));

CREATE TABLE IF NOT EXISTS users (
  id SERIAL,
  name VARCHAR(64) NOT NULL,
  team_id INT REFERENCES teams(id),
  PRIMARY KEY (id, team_id));

CREATE INDEX ON alerts (id);
CREATE INDEX ON alert_history (alert_id);
`
