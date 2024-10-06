--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3 (Debian 16.3-1.pgdg120+1)
-- Dumped by pg_dump version 16.3 (Debian 16.3-1.pgdg120+1)
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
SET default_tablespace = '';
SET default_table_access_method = heap;
--
-- Name: cookie; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.cookie (
    user_id integer NOT NULL,
    cookie_value text,
    description text
);
ALTER TABLE public.cookie OWNER TO tg;
--
-- Name: cookie_user_id_seq; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.cookie_user_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.cookie_user_id_seq OWNER TO tg;
--
-- Name: cookie_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.cookie_user_id_seq OWNED BY public.cookie.user_id;
--
-- Name: forward_to; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.forward_to (
    id bigint NOT NULL,
    user_id bigint,
    forward bigint,
    file_name character varying(40) NOT NULL
);
ALTER TABLE public.forward_to OWNER TO tg;
--
-- Name: forward_to_id_seq; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.forward_to_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.forward_to_id_seq OWNER TO tg;
--
-- Name: forward_to_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.forward_to_id_seq OWNED BY public.forward_to.id;
--
-- Name: media_files; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.media_files (
    id integer NOT NULL,
    social_network_id character varying(255) NOT NULL,
    social_network_name character varying(255) NOT NULL,
    file_id character varying(255) NOT NULL,
    file_type character varying(50) NOT NULL,
    bot character varying(50)
);
ALTER TABLE public.media_files OWNER TO tg;
--
-- Name: media_files_id_seq1; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.media_files_id_seq1 AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.media_files_id_seq1 OWNER TO tg;
--
-- Name: media_files_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.media_files_id_seq1 OWNED BY public.media_files.id;
--
-- Name: message_queue; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.message_queue (
    id integer NOT NULL,
    user_id bigint,
    message text NOT NULL,
    bot character varying(50),
    social_network_name character varying(255) NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now()
);
ALTER TABLE public.message_queue OWNER TO tg;
--
-- Name: message_queue_id_seq; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.message_queue_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.message_queue_id_seq OWNER TO tg;
--
-- Name: message_queue_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.message_queue_id_seq OWNED BY public.message_queue.id;
--
-- Name: statistics; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.statistics (
    id integer NOT NULL,
    date date DEFAULT CURRENT_DATE,
    user_count integer,
    media_file_count integer
);
ALTER TABLE public.statistics OWNER TO tg;
--
-- Name: statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.statistics_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.statistics_id_seq OWNER TO tg;
--
-- Name: statistics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.statistics_id_seq OWNED BY public.statistics.id;
--
-- Name: user_message; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.user_message (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    first_name character varying(255),
    last_name character varying(255),
    username character varying(255),
    language_code character varying(255),
    message text,
    created_at timestamp without time zone DEFAULT now()
);
ALTER TABLE public.user_message OWNER TO tg;
--
-- Name: user_message_id_seq1; Type: SEQUENCE; Schema: public; Owner: tg
--

CREATE SEQUENCE public.user_message_id_seq1 AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.user_message_id_seq1 OWNER TO tg;
--
-- Name: user_message_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: tg
--

ALTER SEQUENCE public.user_message_id_seq1 OWNED BY public.user_message.id;
--
-- Name: users; Type: TABLE; Schema: public; Owner: tg
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    is_bot boolean,
    first_name character varying(255),
    last_name character varying(255),
    username character varying(255),
    language_code character varying(255),
    can_join_groups boolean,
    can_read_all_group_messages boolean,
    supports_inline_queries boolean,
    update timestamp without time zone DEFAULT now(),
    bot character varying(50)
);
ALTER TABLE public.users OWNER TO tg;
--
-- Name: cookie user_id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.cookie
ALTER COLUMN user_id
SET DEFAULT nextval('public.cookie_user_id_seq'::regclass);
--
-- Name: forward_to id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.forward_to
ALTER COLUMN id
SET DEFAULT nextval('public.forward_to_id_seq'::regclass);
--
-- Name: media_files id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.media_files
ALTER COLUMN id
SET DEFAULT nextval('public.media_files_id_seq1'::regclass);
--
-- Name: message_queue id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.message_queue
ALTER COLUMN id
SET DEFAULT nextval('public.message_queue_id_seq'::regclass);
--
-- Name: statistics id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.statistics
ALTER COLUMN id
SET DEFAULT nextval('public.statistics_id_seq'::regclass);
--
-- Name: user_message id; Type: DEFAULT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.user_message
ALTER COLUMN id
SET DEFAULT nextval('public.user_message_id_seq1'::regclass);
--
-- Name: cookie cookie_pkey; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.cookie
ADD CONSTRAINT cookie_pkey PRIMARY KEY (user_id);
--
-- Name: forward_to forward_to_pkey; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.forward_to
ADD CONSTRAINT forward_to_pkey PRIMARY KEY (id);
--
-- Name: media_files media_files_pkey1; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.media_files
ADD CONSTRAINT media_files_pkey1 PRIMARY KEY (id);
--
-- Name: message_queue message_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.message_queue
ADD CONSTRAINT message_queue_pkey PRIMARY KEY (id);
--
-- Name: statistics statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.statistics
ADD CONSTRAINT statistics_pkey PRIMARY KEY (id);
--
-- Name: user_message user_message_pkey1; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.user_message
ADD CONSTRAINT user_message_pkey1 PRIMARY KEY (id);
--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.users
ADD CONSTRAINT users_pkey PRIMARY KEY (id);
--
-- Name: idx_users_id; Type: INDEX; Schema: public; Owner: tg
--

CREATE INDEX idx_users_id ON public.users USING btree (id);
--
-- Name: forward_to forward_to_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.forward_to
ADD CONSTRAINT forward_to_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
--
-- Name: user_message user_message_user_id_fkey1; Type: FK CONSTRAINT; Schema: public; Owner: tg
--

ALTER TABLE ONLY public.user_message
ADD CONSTRAINT user_message_user_id_fkey1 FOREIGN KEY (user_id) REFERENCES public.users(id);
--
-- PostgreSQL database dump complete