SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: build_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.build_status AS ENUM (
    'queue',
    'clone',
    'checkout',
    'build',
    'failure',
    'success'
);


--
-- Name: log_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.log_kind AS ENUM (
    'stdout',
    'stderr',
    'nix_log',
    'system_log'
);


--
-- Name: auto_row_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.auto_row_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    NEW.updated_at = clock_timestamp();
    RETURN NEW;
  END;
$$;


--
-- Name: mark_build_finished(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.mark_build_finished() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    IF NEW.status = 'failure' OR NEW.status = 'success' THEN
      NEW.finished_at = clock_timestamp();
    END IF;
    RETURN NEW;
  END;
$$;


--
-- Name: notify_logline_inserted(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.notify_logline_inserted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  DECLARE BEGIN
    PERFORM pg_notify('loglines'::text, row_to_json(NEW)::text);
    RETURN NEW;
  END;
$$;


--
-- Name: notify_queue_inserted(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.notify_queue_inserted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  DECLARE
  BEGIN
    PERFORM pg_notify(CAST('scylla_queue' AS TEXT), CAST(NEW.name AS text) || ' ' || CAST(NEW.id AS text));
    RETURN NEW;
  END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: builds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.builds (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone,
    status_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    finished_at timestamp with time zone,
    project_id integer NOT NULL,
    status public.build_status DEFAULT 'queue'::public.build_status NOT NULL,
    data jsonb NOT NULL
);


--
-- Name: builds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.builds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: builds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.builds_id_seq OWNED BY public.builds.id;


--
-- Name: loglines; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loglines (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    build_id integer NOT NULL,
    line text
);


--
-- Name: loglines_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.loglines_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: loglines_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.loglines_id_seq OWNED BY public.loglines.id;


--
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone,
    build_id integer NOT NULL,
    kind public.log_kind DEFAULT 'stdout'::public.log_kind NOT NULL,
    content text
);


--
-- Name: logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.logs_id_seq OWNED BY public.logs.id;


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone,
    name text NOT NULL
);


--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.projects_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- Name: queue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.queue (
    id bigint NOT NULL,
    name text DEFAULT 'default'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    run_at timestamp with time zone DEFAULT now() NOT NULL,
    args jsonb DEFAULT '{}'::json NOT NULL,
    errors text[] DEFAULT '{}'::text[]
);


--
-- Name: queue_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.queue_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: queue_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.queue_id_seq OWNED BY public.queue.id;


--
-- Name: results; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.results (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone,
    build_id integer NOT NULL,
    path text NOT NULL
);


--
-- Name: results_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.results_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: results_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.results_id_seq OWNED BY public.results.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: builds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.builds ALTER COLUMN id SET DEFAULT nextval('public.builds_id_seq'::regclass);


--
-- Name: loglines id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loglines ALTER COLUMN id SET DEFAULT nextval('public.loglines_id_seq'::regclass);


--
-- Name: logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs ALTER COLUMN id SET DEFAULT nextval('public.logs_id_seq'::regclass);


--
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- Name: queue id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue ALTER COLUMN id SET DEFAULT nextval('public.queue_id_seq'::regclass);


--
-- Name: results id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.results ALTER COLUMN id SET DEFAULT nextval('public.results_id_seq'::regclass);


--
-- Name: builds builds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.builds
    ADD CONSTRAINT builds_pkey PRIMARY KEY (id);


--
-- Name: loglines loglines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loglines
    ADD CONSTRAINT loglines_pkey PRIMARY KEY (id);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);


--
-- Name: projects projects_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_name_key UNIQUE (name);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: queue queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_pkey PRIMARY KEY (id);


--
-- Name: results results_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.results
    ADD CONSTRAINT results_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: logline_build_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX logline_build_id ON public.loglines USING btree (build_id);


--
-- Name: queue_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX queue_name ON public.queue USING btree (id, name);


--
-- Name: builds after_build; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER after_build BEFORE UPDATE ON public.builds FOR EACH ROW EXECUTE PROCEDURE public.mark_build_finished();


--
-- Name: loglines inserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER inserted AFTER INSERT ON public.loglines FOR EACH ROW EXECUTE PROCEDURE public.notify_logline_inserted();


--
-- Name: queue queue_insert_notify; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER queue_insert_notify AFTER INSERT ON public.queue FOR EACH ROW EXECUTE PROCEDURE public.notify_queue_inserted();


--
-- Name: projects updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER updated BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE PROCEDURE public.auto_row_updated_at();


--
-- Name: builds updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER updated BEFORE UPDATE ON public.builds FOR EACH ROW EXECUTE PROCEDURE public.auto_row_updated_at();


--
-- Name: logs updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER updated BEFORE UPDATE ON public.logs FOR EACH ROW EXECUTE PROCEDURE public.auto_row_updated_at();


--
-- Name: results updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER updated BEFORE UPDATE ON public.results FOR EACH ROW EXECUTE PROCEDURE public.auto_row_updated_at();


--
-- Name: builds builds_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.builds
    ADD CONSTRAINT builds_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: loglines loglines_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loglines
    ADD CONSTRAINT loglines_build_id_fkey FOREIGN KEY (build_id) REFERENCES public.builds(id);


--
-- Name: logs logs_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_build_id_fkey FOREIGN KEY (build_id) REFERENCES public.builds(id);


--
-- Name: results results_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.results
    ADD CONSTRAINT results_build_id_fkey FOREIGN KEY (build_id) REFERENCES public.builds(id);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20180927123207');
