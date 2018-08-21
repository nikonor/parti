--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.8
-- Dumped by pg_dump version 10.5 (Ubuntu 10.5-1.pgdg16.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: go_template; Type: DATABASE; Schema: -; Owner: apps
--

CREATE DATABASE go_template WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'ru_RU.UTF-8' LC_CTYPE = 'ru_RU.UTF-8';


ALTER DATABASE go_template OWNER TO apps;

\connect go_template

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: f_set_only_one_active_config(); Type: FUNCTION; Schema: public; Owner: apps
--

CREATE FUNCTION public.f_set_only_one_active_config() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF upper(TG_OP)='INSERT' THEN
    UPDATE config SET is_active='f' WHERE name!=NEW.name AND is_active='t';
  ELSIF NEW.is_active AND OLD.is_active != NEW.is_active THEN
    UPDATE config SET is_active='f' WHERE name!=NEW.name AND is_active='t';
  END IF;
  RETURN NEW;
END;
$$;


ALTER FUNCTION public.f_set_only_one_active_config() OWNER TO apps;

--
-- Name: set_create_time(); Type: FUNCTION; Schema: public; Owner: apps
--

CREATE FUNCTION public.set_create_time() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    UPDATE_STR text;
BEGIN
    NEW.create_time := now();
    NEW.update_time := now();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_create_time() OWNER TO apps;

--
-- Name: set_update_time(); Type: FUNCTION; Schema: public; Owner: apps
--

CREATE FUNCTION public.set_update_time() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    UPDATE_STR text;
BEGIN
    NEW.update_time := now();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_update_time() OWNER TO apps;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: config; Type: TABLE; Schema: public; Owner: apps
--

CREATE TABLE public.config (
    name character varying(20) NOT NULL,
    data jsonb,
    create_time timestamp with time zone DEFAULT now(),
    update_time timestamp with time zone DEFAULT now(),
    is_active boolean DEFAULT true,
    description character varying(50)
);


ALTER TABLE public.config OWNER TO apps;

--
-- Data for Name: config; Type: TABLE DATA; Schema: public; Owner: apps
--

INSERT INTO public.config VALUES ('develop', '{"module": {"http": {"port": ":8087", "title": "Настройки Web сервиса", "socket": "/var/run/emp/__APP_NAME__.socket", "useSocket": false, "adminkaTimeOut": 0, "cancelingTimeOut": 15}, "config": {"updated": "2017-10-09 19:38:01", "revision": 1}, "instance": {"timeout": 1, "attempts": 3, "maxCount": 0, "loginTimeout": 5, "bgInstancesCount": 0}}, "functional": {}, "environment": {"rmq": {"mail": {"queue": {"name": "svc.com.email.bti", "durable": true, "exclusive": false, "autoDelete": false}, "title": "Очередь для отправки почты", "exchange": {"name": "amq.topic", "type": "topic", "durable": true, "autoDelete": false}, "routingKey": "svc.com.email.bti", "connectConf": {"addr": "10.250.27.11", "port": "5672", "login": "emp", "password": "emp"}, "rmqPoolSize": 1, "connectToQueue": true}, "black_hole": {"queue": {"name": "black.hole", "durable": true, "exclusive": false, "autoDelete": false}, "title": "Очередь для отправки в никуда", "routingKey": "black.hole", "connectConf": {"addr": "rabbitmq", "port": "5672", "login": "emp", "password": "emp"}, "rmqPoolSize": 1, "connectToQueue": true}}, "redis": {"db": 0, "host": "redis", "port": ":6379", "password": ""}}, "integration": {}}', '2018-02-02 16:50:11.840166+03', '2018-02-05 16:13:08.46233+03', true, NULL);


--
-- Name: config config_name_key; Type: CONSTRAINT; Schema: public; Owner: apps
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_name_key UNIQUE (name);


--
-- Name: config config_pkey; Type: CONSTRAINT; Schema: public; Owner: apps
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_pkey PRIMARY KEY (name);


--
-- Name: config t_config_set_create_time_bi; Type: TRIGGER; Schema: public; Owner: apps
--

CREATE TRIGGER t_config_set_create_time_bi BEFORE INSERT ON public.config FOR EACH ROW EXECUTE PROCEDURE public.set_create_time();


--
-- Name: config t_config_set_update_time_bu; Type: TRIGGER; Schema: public; Owner: apps
--

CREATE TRIGGER t_config_set_update_time_bu BEFORE UPDATE ON public.config FOR EACH ROW EXECUTE PROCEDURE public.set_update_time();


--
-- Name: config t_set_only_one_active_config_biu; Type: TRIGGER; Schema: public; Owner: apps
--

CREATE TRIGGER t_set_only_one_active_config_biu BEFORE INSERT OR UPDATE ON public.config FOR EACH ROW EXECUTE PROCEDURE public.f_set_only_one_active_config();


--
-- PostgreSQL database dump complete
--

