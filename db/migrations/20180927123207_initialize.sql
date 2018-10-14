-- migrate:up

CREATE TABLE queue(
	id          BIGSERIAL   NOT NULL UNIQUE PRIMARY KEY,
	name        TEXT        NOT NULL DEFAULT 'default',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	run_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
	args        JSONB       NOT NULL DEFAULT '{}'::json,
	errors      TEXT[]      DEFAULT '{}'
);

CREATE UNIQUE INDEX queue_name ON queue (id, name);
CREATE OR REPLACE FUNCTION notify_queue_inserted() RETURNS trigger AS
$$
  DECLARE
  BEGIN
    PERFORM pg_notify(CAST('scylla_queue' AS TEXT), CAST(NEW.name AS text) || ' ' || CAST(NEW.id AS text));
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER queue_insert_notify
  AFTER INSERT ON queue
  FOR EACH ROW EXECUTE PROCEDURE notify_queue_inserted();

CREATE TABLE projects (
  id         SERIAL      NOT NULL UNIQUE PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  updated_at TIMESTAMPTZ,
  name       TEXT        NOT NULL UNIQUE
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
  id          SERIAL       NOT NULL UNIQUE PRIMARY KEY,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT clock_timestamp(),
  updated_at  TIMESTAMPTZ,
  status_at   TIMESTAMPTZ  NOT NULL DEFAULT clock_timestamp(),
  finished_at TIMESTAMPTZ,
  project_id  INT          NOT NULL REFERENCES projects(id),
  status      build_status NOT NULL DEFAULT 'queue',
  data        JSONB        NOT NULL
);

CREATE TYPE log_kind AS ENUM (
  'stdout',
  'stderr',
  'nix_log',
  'system_log'
);

CREATE TABLE logs (
  id         SERIAL      NOT NULL UNIQUE PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  updated_at TIMESTAMPTZ,
  build_id   INT         NOT NULL REFERENCES builds(id),
  kind       log_kind    NOT NULL DEFAULT 'stdout',
  content    TEXT
);

CREATE TABLE loglines (
  id         SERIAL      NOT NULL UNIQUE PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  build_id   INT         NOT NULL REFERENCES builds(id),
  line       TEXT
);

CREATE INDEX logline_build_id ON loglines (build_id);

CREATE TABLE results (
  id         SERIAL      NOT NULL UNIQUE PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  updated_at TIMESTAMPTZ,
  build_id   INT         NOT NULL REFERENCES builds(id),
  path       TEXT        NOT NULL
);

CREATE FUNCTION auto_row_updated_at() RETURNS TRIGGER AS
$$
  BEGIN
    NEW.updated_at = clock_timestamp();
    RETURN NEW;
  END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER updated BEFORE UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER updated BEFORE UPDATE ON builds   FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER updated BEFORE UPDATE ON logs     FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();
CREATE TRIGGER updated BEFORE UPDATE ON results  FOR EACH ROW EXECUTE PROCEDURE auto_row_updated_at();

CREATE FUNCTION mark_build_finished() RETURNS TRIGGER AS
$$
  BEGIN
    IF NEW.status = 'failure' OR NEW.status = 'success' THEN
      NEW.finished_at = clock_timestamp();
    END IF;
    RETURN NEW;
  END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER after_build BEFORE UPDATE ON builds FOR EACH ROW EXECUTE PROCEDURE mark_build_finished();

CREATE FUNCTION notify_logline_inserted() RETURNS trigger AS
$$
  DECLARE BEGIN
    PERFORM pg_notify('loglines'::text, row_to_json(NEW)::text);
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER inserted AFTER INSERT ON loglines FOR EACH ROW EXECUTE PROCEDURE notify_logline_inserted();

-- migrate:down

DROP FUNCTION IF EXISTS auto_row_updated_at CASCADE;
DROP FUNCTION IF EXISTS notify_queue_inserted CASCADE;
DROP FUNCTION IF EXISTS notify_logline_inserted CASCADE;
DROP FUNCTION IF EXISTS mark_build_finished CASCADE;

DROP TABLE IF EXISTS builds CASCADE;
DROP TABLE IF EXISTS loglines CASCADE;
DROP TABLE IF EXISTS logs CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS queue CASCADE;
DROP TABLE IF EXISTS results CASCADE;

DROP TYPE IF EXISTS build_status;
DROP TYPE IF EXISTS log_kind;
