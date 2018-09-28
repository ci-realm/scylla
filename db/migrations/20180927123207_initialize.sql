-- migrate:up

CREATE TABLE projects (
  id              SERIAL                   NOT NULL UNIQUE PRIMARY KEY,
  created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
  updated_at      TIMESTAMP WITH TIME ZONE,
  name            TEXT                     NOT NULL UNIQUE
);

CREATE TYPE build_status as enum (
  'queue',
  'clone',
  'checkout',
  'build',
  'failure',
  'success'
);

CREATE TABLE builds (
  id              SERIAL                   NOT NULL UNIQUE PRIMARY KEY,
  created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
  updated_at      TIMESTAMP WITH TIME ZONE,
  project_id      INT                      NOT NULL REFERENCES projects(id),
  status          build_status             NOT NULL DEFAULT 'queue',
  data            JSONB                    NOT NULL
);

CREATE TYPE log_kind AS ENUM (
  'stdout',
  'stderr',
  'nix_log',
  'system_log'
);

CREATE TABLE logs (
  id              SERIAL                   NOT NULL UNIQUE PRIMARY KEY,
  created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
  updated_at      TIMESTAMP WITH TIME ZONE,
  build_id        INT                      NOT NULL REFERENCES builds(id),
  kind            log_kind                 NOT NULL DEFAULT 'stdout',
  content         TEXT
);

CREATE TABLE results (
  id              SERIAL                   NOT NULL UNIQUE PRIMARY KEY,
  created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
  updated_at      TIMESTAMP WITH TIME ZONE,
  build_id        INT                      NOT NULL REFERENCES builds(id),
  path            TEXT                     NOT NULL
);

CREATE TABLE que_jobs (
  priority    smallint    NOT NULL DEFAULT 100,
  run_at      timestamptz NOT NULL DEFAULT now(),
  job_id      bigserial   NOT NULL,
  job_class   text        NOT NULL,
  args        json        NOT NULL DEFAULT '[]'::json,
  error_count integer     NOT NULL DEFAULT 0,
  last_error  text,
  queue       text        NOT NULL DEFAULT '',
  CONSTRAINT que_jobs_pkey PRIMARY KEY (queue, priority, run_at, job_id)
);

CREATE FUNCTION auto_row_updated_at() RETURNS TRIGGER AS
$$
  BEGIN
    NEW.updated_at = clock_timestamp();
    RETURN NEW;
  END;
$$
language 'plpgsql';

CREATE TRIGGER auto_update_trigger BEFORE UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER auto_update_trigger BEFORE UPDATE ON builds FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER auto_update_trigger BEFORE UPDATE ON logs FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER auto_update_trigger BEFORE UPDATE ON results FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();

-- migrate:down

DROP FUNCTION IF EXISTS auto_row_updated_at CASCADE;
DROP TABLE IF EXISTS builds CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS logs CASCADE;
DROP TABLE IF EXISTS results CASCADE;
DROP TABLE IF EXISTS que_jobs CASCADE;
DROP TYPE IF EXISTS build_status;
DROP TYPE IF EXISTS log_kind;
