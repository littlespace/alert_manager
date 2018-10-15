package template

var Schema = `
CREATE TABLE IF NOT EXISTS alerts (
  id SERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  description TEXT NOT NULL,
  entity VARCHAR(128) NOT NULL,
  external_id VARCHAR(32) NOT NULL,
  source VARCHAR(16) NOT NULL,
  device VARCHAR(16),
  site VARCHAR(16),
  owner VARCHAR(16),
  team VARCHAR(16),
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
  scope VARCHAR(16));

CREATE TABLE IF NOT EXISTS suppression_rules (
  id SERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  entities JSON NOT NULL,
  rtype SMALLINT  NOT NULL,
  created_at BIGINT NOT NULL,
  duration INT NOT NULL,
  reason TEXT,
  creator varchar(64) NOT NULL);
`