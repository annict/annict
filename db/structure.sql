SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS '';


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pg_stat_statements; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_statements; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_stat_statements IS 'track execution statistics of all SQL statements executed';


--
-- Name: activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    recipient_id bigint,
    recipient_type character varying(510),
    trackable_id bigint NOT NULL,
    trackable_type character varying(510) NOT NULL,
    action character varying(510),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    work_id bigint,
    episode_id bigint,
    status_id bigint,
    episode_record_id bigint,
    multiple_episode_record_id bigint,
    work_record_id bigint,
    activity_group_id bigint NOT NULL,
    migrated_at timestamp without time zone,
    mer_processed_at timestamp without time zone
);


--
-- Name: activity_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activity_groups (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    itemable_type character varying NOT NULL,
    single boolean DEFAULT false NOT NULL,
    activities_count integer DEFAULT 0 NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: activity_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.activity_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: activity_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.activity_groups_id_seq OWNED BY public.activity_groups.id;


--
-- Name: ar_internal_metadata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_internal_metadata (
    key character varying NOT NULL,
    value character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: casts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.casts (
    id bigint NOT NULL,
    person_id bigint NOT NULL,
    work_id bigint NOT NULL,
    name character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    character_id bigint NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: casts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.casts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: casts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.casts_id_seq OWNED BY public.casts.id;


--
-- Name: channel_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.channel_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: channel_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.channel_groups (
    id bigint DEFAULT nextval('public.channel_groups_id_seq'::regclass) NOT NULL,
    sc_chgid character varying(510),
    name character varying(510) NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: channel_works_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.channel_works_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: channel_works; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.channel_works (
    id bigint DEFAULT nextval('public.channel_works_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    channel_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: channels_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.channels_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: channels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.channels (
    id bigint DEFAULT nextval('public.channels_id_seq'::regclass) NOT NULL,
    channel_group_id bigint NOT NULL,
    sc_chid integer,
    name character varying NOT NULL COLLATE pg_catalog."C",
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    vod boolean DEFAULT false,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    deleted_at timestamp without time zone,
    name_alter character varying DEFAULT ''::character varying NOT NULL,
    unpublished_at timestamp without time zone
);


--
-- Name: character_favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.character_favorites (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    character_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: character_favorites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.character_favorites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: character_favorites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.character_favorites_id_seq OWNED BY public.character_favorites.id;


--
-- Name: character_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.character_images (
    id bigint NOT NULL,
    character_id bigint NOT NULL,
    user_id bigint NOT NULL,
    attachment_file_name character varying NOT NULL,
    attachment_file_size integer NOT NULL,
    attachment_content_type character varying NOT NULL,
    attachment_updated_at timestamp without time zone NOT NULL,
    copyright character varying DEFAULT ''::character varying NOT NULL,
    asin character varying DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: character_images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.character_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: character_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.character_images_id_seq OWNED BY public.character_images.id;


--
-- Name: characters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.characters (
    id bigint NOT NULL,
    name character varying NOT NULL,
    name_kana character varying DEFAULT ''::character varying NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    nickname character varying DEFAULT ''::character varying NOT NULL,
    nickname_en character varying DEFAULT ''::character varying NOT NULL,
    birthday character varying DEFAULT ''::character varying NOT NULL,
    birthday_en character varying DEFAULT ''::character varying NOT NULL,
    age character varying DEFAULT ''::character varying NOT NULL,
    age_en character varying DEFAULT ''::character varying NOT NULL,
    blood_type character varying DEFAULT ''::character varying NOT NULL,
    blood_type_en character varying DEFAULT ''::character varying NOT NULL,
    height character varying DEFAULT ''::character varying NOT NULL,
    height_en character varying DEFAULT ''::character varying NOT NULL,
    weight character varying DEFAULT ''::character varying NOT NULL,
    weight_en character varying DEFAULT ''::character varying NOT NULL,
    nationality character varying DEFAULT ''::character varying NOT NULL,
    nationality_en character varying DEFAULT ''::character varying NOT NULL,
    occupation character varying DEFAULT ''::character varying NOT NULL,
    occupation_en character varying DEFAULT ''::character varying NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    description_en text DEFAULT ''::text NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    description_source character varying DEFAULT ''::character varying NOT NULL,
    description_source_en character varying DEFAULT ''::character varying NOT NULL,
    favorite_users_count integer DEFAULT 0 NOT NULL,
    series_id bigint,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: characters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.characters_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: characters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.characters_id_seq OWNED BY public.characters.id;


--
-- Name: collection_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collection_items (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    collection_id bigint NOT NULL,
    work_id bigint NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: collection_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.collection_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: collection_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.collection_items_id_seq OWNED BY public.collection_items.id;


--
-- Name: collections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collections (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    name character varying NOT NULL,
    description character varying DEFAULT ''::character varying NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    collection_items_count integer DEFAULT 0 NOT NULL
);


--
-- Name: collections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: collections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.collections_id_seq OWNED BY public.collections.id;


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id bigint DEFAULT nextval('public.comments_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    episode_record_id bigint NOT NULL,
    body text NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    work_id bigint,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: cover_images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cover_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: db_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.db_activities (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    trackable_id bigint NOT NULL,
    trackable_type character varying NOT NULL,
    action character varying NOT NULL,
    parameters json,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    root_resource_id bigint,
    root_resource_type character varying,
    object_id bigint,
    object_type character varying
);


--
-- Name: db_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.db_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: db_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.db_activities_id_seq OWNED BY public.db_activities.id;


--
-- Name: db_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.db_comments (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    resource_id bigint NOT NULL,
    resource_type character varying NOT NULL,
    body text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: db_comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.db_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: db_comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.db_comments_id_seq OWNED BY public.db_comments.id;


--
-- Name: delayed_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.delayed_jobs (
    id bigint NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    handler text NOT NULL,
    last_error text,
    run_at timestamp without time zone,
    locked_at timestamp without time zone,
    failed_at timestamp without time zone,
    locked_by character varying,
    queue character varying,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: delayed_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.delayed_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: delayed_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.delayed_jobs_id_seq OWNED BY public.delayed_jobs.id;


--
-- Name: email_confirmations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_confirmations (
    id bigint NOT NULL,
    user_id bigint,
    email public.citext NOT NULL,
    event character varying NOT NULL,
    token character varying NOT NULL,
    back character varying,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: email_confirmations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.email_confirmations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: email_confirmations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.email_confirmations_id_seq OWNED BY public.email_confirmations.id;


--
-- Name: email_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_notifications (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    unsubscription_key character varying NOT NULL,
    event_followed_user boolean DEFAULT true NOT NULL,
    event_liked_episode_record boolean DEFAULT true NOT NULL,
    event_friends_joined boolean DEFAULT true NOT NULL,
    event_next_season_came boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    event_favorite_works_added boolean DEFAULT true NOT NULL,
    event_related_works_added boolean DEFAULT true NOT NULL
);


--
-- Name: email_notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.email_notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: email_notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.email_notifications_id_seq OWNED BY public.email_notifications.id;


--
-- Name: episode_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.episode_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: episode_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.episode_records (
    id bigint DEFAULT nextval('public.episode_records_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    episode_id bigint NOT NULL,
    body text,
    modify_body boolean DEFAULT false NOT NULL,
    twitter_url_hash character varying(510) DEFAULT NULL::character varying,
    facebook_url_hash character varying(510) DEFAULT NULL::character varying,
    twitter_click_count integer DEFAULT 0 NOT NULL,
    facebook_click_count integer DEFAULT 0 NOT NULL,
    comments_count integer DEFAULT 0 NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    shared_twitter boolean DEFAULT false NOT NULL,
    shared_facebook boolean DEFAULT false NOT NULL,
    work_id bigint NOT NULL,
    rating double precision,
    multiple_episode_record_id bigint,
    oauth_application_id bigint,
    rating_state character varying,
    review_id bigint,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL,
    record_id bigint NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: episodes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.episodes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: episodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.episodes (
    id bigint DEFAULT nextval('public.episodes_id_seq'::regclass) NOT NULL,
    work_id bigint NOT NULL,
    number character varying(510) DEFAULT NULL::character varying,
    sort_number integer DEFAULT 0 NOT NULL,
    sc_count integer,
    title character varying(510) DEFAULT NULL::character varying,
    episode_records_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    prev_episode_id bigint,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    fetch_syobocal boolean DEFAULT false NOT NULL,
    raw_number double precision,
    title_ro character varying DEFAULT ''::character varying NOT NULL,
    title_en character varying DEFAULT ''::character varying NOT NULL,
    episode_record_bodies_count integer DEFAULT 0 NOT NULL,
    score double precision,
    ratings_count integer DEFAULT 0 NOT NULL,
    satisfaction_rate double precision,
    number_en character varying DEFAULT ''::character varying NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: faq_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.faq_categories (
    id bigint NOT NULL,
    name character varying NOT NULL,
    locale character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: faq_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.faq_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: faq_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.faq_categories_id_seq OWNED BY public.faq_categories.id;


--
-- Name: faq_contents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.faq_contents (
    id bigint NOT NULL,
    faq_category_id bigint NOT NULL,
    question character varying NOT NULL,
    answer text NOT NULL,
    locale character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: faq_contents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.faq_contents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: faq_contents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.faq_contents_id_seq OWNED BY public.faq_contents.id;


--
-- Name: finished_tips; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.finished_tips (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    tip_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: finished_tips_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.finished_tips_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finished_tips_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.finished_tips_id_seq OWNED BY public.finished_tips.id;


--
-- Name: flashes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.flashes (
    id bigint NOT NULL,
    client_uuid character varying NOT NULL,
    data json,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: flashes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.flashes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: flashes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.flashes_id_seq OWNED BY public.flashes.id;


--
-- Name: follows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.follows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: follows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.follows (
    id bigint DEFAULT nextval('public.follows_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    following_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: forum_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forum_categories (
    id bigint NOT NULL,
    slug character varying NOT NULL,
    name character varying NOT NULL,
    name_en character varying NOT NULL,
    description character varying NOT NULL,
    description_en character varying NOT NULL,
    postable_role character varying NOT NULL,
    sort_number integer NOT NULL,
    forum_posts_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: forum_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.forum_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: forum_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.forum_categories_id_seq OWNED BY public.forum_categories.id;


--
-- Name: forum_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forum_comments (
    id bigint NOT NULL,
    user_id bigint,
    forum_post_id bigint NOT NULL,
    body text NOT NULL,
    edited_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: COLUMN forum_comments.edited_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.forum_comments.edited_at IS 'The datetime which user has changed body.';


--
-- Name: forum_comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.forum_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: forum_comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.forum_comments_id_seq OWNED BY public.forum_comments.id;


--
-- Name: forum_post_participants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forum_post_participants (
    id bigint NOT NULL,
    forum_post_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: forum_post_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.forum_post_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: forum_post_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.forum_post_participants_id_seq OWNED BY public.forum_post_participants.id;


--
-- Name: forum_posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forum_posts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    forum_category_id bigint NOT NULL,
    title character varying NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    forum_comments_count integer DEFAULT 0 NOT NULL,
    edited_at timestamp without time zone,
    last_commented_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: COLUMN forum_posts.edited_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.forum_posts.edited_at IS 'The datetime which user has changed title, body and so on.';


--
-- Name: forum_posts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.forum_posts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: forum_posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.forum_posts_id_seq OWNED BY public.forum_posts.id;


--
-- Name: gumroad_subscribers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gumroad_subscribers (
    id bigint NOT NULL,
    gumroad_id character varying NOT NULL,
    gumroad_product_id character varying NOT NULL,
    gumroad_product_name character varying NOT NULL,
    gumroad_user_id character varying,
    gumroad_user_email character varying,
    gumroad_purchase_ids character varying[] NOT NULL,
    gumroad_created_at timestamp without time zone NOT NULL,
    gumroad_cancelled_at timestamp without time zone,
    gumroad_user_requested_cancellation_at timestamp without time zone,
    gumroad_charge_occurrence_count timestamp without time zone,
    gumroad_ended_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: gumroad_subscribers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gumroad_subscribers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gumroad_subscribers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gumroad_subscribers_id_seq OWNED BY public.gumroad_subscribers.id;


--
-- Name: internal_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_statistics (
    id bigint NOT NULL,
    key character varying NOT NULL,
    value double precision NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: internal_statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.internal_statistics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: internal_statistics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.internal_statistics_id_seq OWNED BY public.internal_statistics.id;


--
-- Name: library_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.library_entries (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    next_episode_id bigint,
    kind integer,
    watched_episode_ids bigint[] DEFAULT '{}'::integer[] NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    status_id bigint,
    program_id bigint,
    next_slot_id bigint,
    note text DEFAULT ''::text NOT NULL
);


--
-- Name: library_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.library_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: library_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.library_entries_id_seq OWNED BY public.library_entries.id;


--
-- Name: likes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.likes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: likes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.likes (
    id bigint DEFAULT nextval('public.likes_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    recipient_id bigint NOT NULL,
    recipient_type character varying(510) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: multiple_episode_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.multiple_episode_records (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: multiple_episode_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.multiple_episode_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: multiple_episode_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.multiple_episode_records_id_seq OWNED BY public.multiple_episode_records.id;


--
-- Name: mute_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mute_users (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    muted_user_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: mute_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.mute_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mute_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.mute_users_id_seq OWNED BY public.mute_users.id;


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    action_user_id bigint NOT NULL,
    trackable_id bigint NOT NULL,
    trackable_type character varying(510) NOT NULL,
    action character varying(510) NOT NULL,
    read boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: number_formats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.number_formats (
    id bigint NOT NULL,
    name character varying NOT NULL,
    data character varying[] DEFAULT '{}'::character varying[] NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    format character varying DEFAULT ''::character varying NOT NULL
);


--
-- Name: number_formats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.number_formats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: number_formats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.number_formats_id_seq OWNED BY public.number_formats.id;


--
-- Name: oauth_access_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_access_grants (
    id bigint NOT NULL,
    resource_owner_id bigint NOT NULL,
    application_id bigint NOT NULL,
    token character varying NOT NULL,
    expires_in integer NOT NULL,
    redirect_uri text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    revoked_at timestamp without time zone,
    scopes character varying
);


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_access_grants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_access_grants_id_seq OWNED BY public.oauth_access_grants.id;


--
-- Name: oauth_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_access_tokens (
    id bigint NOT NULL,
    resource_owner_id bigint NOT NULL,
    application_id bigint,
    token character varying NOT NULL,
    refresh_token character varying,
    expires_in integer,
    revoked_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    scopes character varying,
    previous_refresh_token character varying DEFAULT ''::character varying NOT NULL,
    description character varying DEFAULT ''::character varying NOT NULL
);


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_access_tokens_id_seq OWNED BY public.oauth_access_tokens.id;


--
-- Name: oauth_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_applications (
    id bigint NOT NULL,
    name character varying NOT NULL,
    uid character varying NOT NULL,
    secret character varying NOT NULL,
    redirect_uri text NOT NULL,
    scopes character varying DEFAULT ''::character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    owner_id bigint,
    owner_type character varying,
    confidential boolean DEFAULT true NOT NULL,
    deleted_at timestamp without time zone,
    hide_social_login boolean DEFAULT false NOT NULL
);


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_applications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_applications_id_seq OWNED BY public.oauth_applications.id;


--
-- Name: organization_favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_favorites (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    organization_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    watched_works_count integer DEFAULT 0 NOT NULL
);


--
-- Name: organization_favorites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.organization_favorites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: organization_favorites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.organization_favorites_id_seq OWNED BY public.organization_favorites.id;


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id bigint NOT NULL,
    name character varying NOT NULL,
    url character varying,
    wikipedia_url character varying,
    twitter_username character varying,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    name_kana character varying DEFAULT ''::character varying NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    url_en character varying DEFAULT ''::character varying NOT NULL,
    wikipedia_url_en character varying DEFAULT ''::character varying NOT NULL,
    twitter_username_en character varying DEFAULT ''::character varying NOT NULL,
    favorite_users_count integer DEFAULT 0 NOT NULL,
    staffs_count integer DEFAULT 0 NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.organizations_id_seq OWNED BY public.organizations.id;


--
-- Name: people; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.people (
    id bigint NOT NULL,
    prefecture_id bigint,
    name character varying NOT NULL,
    name_kana character varying DEFAULT ''::character varying NOT NULL,
    nickname character varying,
    gender character varying,
    url character varying,
    wikipedia_url character varying,
    twitter_username character varying,
    birthday date,
    blood_type character varying,
    height integer,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    nickname_en character varying DEFAULT ''::character varying NOT NULL,
    url_en character varying DEFAULT ''::character varying NOT NULL,
    wikipedia_url_en character varying DEFAULT ''::character varying NOT NULL,
    twitter_username_en character varying DEFAULT ''::character varying NOT NULL,
    favorite_users_count integer DEFAULT 0 NOT NULL,
    casts_count integer DEFAULT 0 NOT NULL,
    staffs_count integer DEFAULT 0 NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: people_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.people_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: people_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.people_id_seq OWNED BY public.people.id;


--
-- Name: person_favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.person_favorites (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    person_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    watched_works_count integer DEFAULT 0 NOT NULL
);


--
-- Name: person_favorites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.person_favorites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: person_favorites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.person_favorites_id_seq OWNED BY public.person_favorites.id;


--
-- Name: prefectures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prefectures (
    id bigint NOT NULL,
    name character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: prefectures_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.prefectures_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: prefectures_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.prefectures_id_seq OWNED BY public.prefectures.id;


--
-- Name: profiles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profiles (
    id bigint DEFAULT nextval('public.profiles_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    name character varying(510) DEFAULT ''::character varying NOT NULL,
    description character varying(510) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    background_image_animated boolean DEFAULT false NOT NULL,
    tombo_avatar_file_name character varying,
    tombo_avatar_content_type character varying,
    tombo_avatar_file_size integer,
    tombo_avatar_updated_at timestamp without time zone,
    tombo_background_image_file_name character varying,
    tombo_background_image_content_type character varying,
    tombo_background_image_file_size integer,
    tombo_background_image_updated_at timestamp without time zone,
    url character varying,
    image_data text,
    background_image_data text
);


--
-- Name: programs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.programs (
    id bigint NOT NULL,
    channel_id bigint NOT NULL,
    work_id bigint NOT NULL,
    url character varying,
    started_at timestamp without time zone,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    vod_title_code character varying DEFAULT ''::character varying NOT NULL,
    vod_title_name character varying DEFAULT ''::character varying NOT NULL,
    rebroadcast boolean DEFAULT false NOT NULL,
    minimum_episode_generatable_number integer DEFAULT 1 NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: programs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.programs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: programs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.programs_id_seq OWNED BY public.programs.id;


--
-- Name: providers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.providers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providers (
    id bigint DEFAULT nextval('public.providers_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    name character varying(510) NOT NULL,
    uid character varying(510) NOT NULL,
    token character varying(510) NOT NULL,
    token_expires_at integer,
    token_secret character varying(510) DEFAULT NULL::character varying,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: reactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reactions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    target_user_id bigint NOT NULL,
    kind character varying NOT NULL,
    collection_item_id bigint,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reactions_id_seq OWNED BY public.reactions.id;


--
-- Name: receptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.receptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: receptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.receptions (
    id bigint DEFAULT nextval('public.receptions_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    channel_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.records (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    impressions_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    watched_at timestamp without time zone NOT NULL
);


--
-- Name: records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.records_id_seq OWNED BY public.records.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(510) NOT NULL
);


--
-- Name: seasons_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.seasons_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: seasons; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seasons (
    id bigint DEFAULT nextval('public.seasons_id_seq'::regclass) NOT NULL,
    name character varying(510) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    sort_number integer NOT NULL,
    year integer NOT NULL
);


--
-- Name: series; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.series (
    id bigint NOT NULL,
    name character varying NOT NULL,
    name_ro character varying DEFAULT ''::character varying NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    series_works_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone,
    name_alter character varying DEFAULT ''::character varying NOT NULL,
    name_alter_en character varying DEFAULT ''::character varying NOT NULL
);


--
-- Name: series_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.series_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: series_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.series_id_seq OWNED BY public.series.id;


--
-- Name: series_works; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.series_works (
    id bigint NOT NULL,
    series_id bigint NOT NULL,
    work_id bigint NOT NULL,
    summary character varying DEFAULT ''::character varying NOT NULL,
    summary_en character varying DEFAULT ''::character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: series_works_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.series_works_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: series_works_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.series_works_id_seq OWNED BY public.series_works.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id bigint NOT NULL,
    session_id character varying NOT NULL,
    data jsonb DEFAULT '"{}"'::jsonb NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sessions_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sessions_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sessions_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sessions_id_seq1 OWNED BY public.sessions.id;


--
-- Name: settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.settings (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    hide_record_body boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    share_record_to_twitter boolean DEFAULT false,
    share_record_to_facebook boolean DEFAULT false,
    slots_sort_type character varying DEFAULT ''::character varying NOT NULL,
    display_option_work_list character varying DEFAULT 'list_detailed'::character varying NOT NULL,
    display_option_user_work_list character varying DEFAULT 'grid_detailed'::character varying NOT NULL,
    records_sort_type character varying DEFAULT 'created_at_desc'::character varying NOT NULL,
    display_option_record_list character varying DEFAULT 'all_comments'::character varying NOT NULL,
    share_review_to_twitter boolean DEFAULT false NOT NULL,
    share_review_to_facebook boolean DEFAULT false NOT NULL,
    hide_supporter_badge boolean DEFAULT false NOT NULL,
    share_status_to_twitter boolean DEFAULT false NOT NULL,
    share_status_to_facebook boolean DEFAULT false NOT NULL,
    privacy_policy_agreed boolean DEFAULT false NOT NULL,
    timeline_mode character varying DEFAULT 'following'::character varying NOT NULL
);


--
-- Name: settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.settings_id_seq OWNED BY public.settings.id;


--
-- Name: slots_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.slots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: slots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slots (
    id bigint DEFAULT nextval('public.slots_id_seq'::regclass) NOT NULL,
    channel_id bigint NOT NULL,
    episode_id bigint,
    work_id bigint NOT NULL,
    started_at timestamp with time zone NOT NULL,
    sc_last_update timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    sc_pid integer,
    rebroadcast boolean DEFAULT false NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    program_id bigint,
    number integer,
    irregular boolean DEFAULT false NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: staffs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.staffs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: staffs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.staffs (
    id bigint DEFAULT nextval('public.staffs_id_seq'::regclass) NOT NULL,
    work_id bigint NOT NULL,
    name character varying NOT NULL,
    role character varying NOT NULL,
    role_other character varying,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    resource_id bigint NOT NULL,
    resource_type character varying NOT NULL,
    name_en character varying DEFAULT ''::character varying NOT NULL,
    role_other_en character varying DEFAULT ''::character varying NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: staffs_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.staffs_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: staffs_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.staffs_id_seq1 OWNED BY public.staffs.id;


--
-- Name: statuses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.statuses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.statuses (
    id bigint DEFAULT nextval('public.statuses_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    kind integer NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    oauth_application_id bigint
);


--
-- Name: syobocal_alerts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.syobocal_alerts (
    id bigint NOT NULL,
    work_id bigint,
    kind integer NOT NULL,
    sc_prog_item_id integer,
    sc_sub_title character varying(255),
    sc_prog_comment character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: syobocal_alerts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.syobocal_alerts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: syobocal_alerts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.syobocal_alerts_id_seq OWNED BY public.syobocal_alerts.id;


--
-- Name: tips; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tips (
    id bigint NOT NULL,
    target integer NOT NULL,
    slug character varying(255) NOT NULL,
    title character varying(255) NOT NULL,
    icon_name character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: tips_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tips_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tips_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tips_id_seq OWNED BY public.tips.id;


--
-- Name: trailers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trailers (
    id bigint NOT NULL,
    work_id bigint NOT NULL,
    url character varying NOT NULL,
    title character varying NOT NULL,
    thumbnail_file_name character varying,
    thumbnail_content_type character varying,
    thumbnail_file_size integer,
    thumbnail_updated_at timestamp without time zone,
    sort_number integer DEFAULT 0 NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    title_en character varying DEFAULT ''::character varying NOT NULL,
    image_data text,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: trailers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trailers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trailers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trailers_id_seq OWNED BY public.trailers.id;


--
-- Name: twitter_bots_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.twitter_bots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: twitter_bots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.twitter_bots (
    id bigint DEFAULT nextval('public.twitter_bots_id_seq'::regclass) NOT NULL,
    name character varying(510) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: twitter_watching_lists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.twitter_watching_lists (
    id bigint NOT NULL,
    username character varying NOT NULL,
    name character varying NOT NULL,
    since_id character varying,
    discord_webhook_url character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: twitter_watching_lists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.twitter_watching_lists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: twitter_watching_lists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.twitter_watching_lists_id_seq OWNED BY public.twitter_watching_lists.id;


--
-- Name: userland_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.userland_categories (
    id bigint NOT NULL,
    name character varying NOT NULL,
    name_en character varying NOT NULL,
    sort_number integer DEFAULT 0 NOT NULL,
    userland_projects_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: userland_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.userland_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: userland_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.userland_categories_id_seq OWNED BY public.userland_categories.id;


--
-- Name: userland_project_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.userland_project_members (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    userland_project_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: userland_project_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.userland_project_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: userland_project_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.userland_project_members_id_seq OWNED BY public.userland_project_members.id;


--
-- Name: userland_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.userland_projects (
    id bigint NOT NULL,
    userland_category_id bigint NOT NULL,
    name character varying NOT NULL,
    summary character varying NOT NULL,
    description text NOT NULL,
    url character varying NOT NULL,
    icon_file_name character varying,
    icon_content_type character varying,
    icon_file_size integer,
    icon_updated_at timestamp without time zone,
    available boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL,
    image_data text
);


--
-- Name: userland_projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.userland_projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: userland_projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.userland_projects_id_seq OWNED BY public.userland_projects.id;


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint DEFAULT nextval('public.users_id_seq'::regclass) NOT NULL,
    username public.citext NOT NULL,
    email public.citext NOT NULL,
    role integer NOT NULL,
    encrypted_password character varying(510) DEFAULT ''::character varying NOT NULL,
    remember_created_at timestamp with time zone,
    sign_in_count integer DEFAULT 0 NOT NULL,
    current_sign_in_at timestamp with time zone,
    last_sign_in_at timestamp with time zone,
    current_sign_in_ip character varying(510) DEFAULT NULL::character varying,
    last_sign_in_ip character varying(510) DEFAULT NULL::character varying,
    confirmation_token character varying(510) DEFAULT NULL::character varying,
    confirmed_at timestamp with time zone,
    confirmation_sent_at timestamp with time zone,
    unconfirmed_email character varying(510) DEFAULT NULL::character varying,
    episode_records_count integer DEFAULT 0 NOT NULL,
    notifications_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    time_zone character varying NOT NULL,
    locale character varying NOT NULL,
    reset_password_token character varying,
    reset_password_sent_at timestamp without time zone,
    record_cache_expired_at timestamp without time zone,
    status_cache_expired_at timestamp without time zone,
    work_tag_cache_expired_at timestamp without time zone,
    work_comment_cache_expired_at timestamp without time zone,
    gumroad_subscriber_id bigint,
    allowed_locales character varying[],
    records_count integer DEFAULT 0 NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    deleted_at timestamp without time zone,
    character_favorites_count integer DEFAULT 0 NOT NULL,
    person_favorites_count integer DEFAULT 0 NOT NULL,
    organization_favorites_count integer DEFAULT 0 NOT NULL,
    plan_to_watch_works_count integer DEFAULT 0 NOT NULL,
    watching_works_count integer DEFAULT 0 NOT NULL,
    completed_works_count integer DEFAULT 0 NOT NULL,
    on_hold_works_count integer DEFAULT 0 NOT NULL,
    dropped_works_count integer DEFAULT 0 NOT NULL,
    following_count integer DEFAULT 0 NOT NULL,
    followers_count integer DEFAULT 0 NOT NULL
);


--
-- Name: versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.versions (
    id bigint DEFAULT nextval('public.versions_id_seq'::regclass) NOT NULL,
    item_type character varying(510) NOT NULL,
    item_id integer NOT NULL,
    event character varying(510) NOT NULL,
    whodunnit character varying(510) DEFAULT NULL::character varying,
    object text,
    created_at timestamp with time zone
);


--
-- Name: vod_titles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vod_titles (
    id bigint NOT NULL,
    channel_id bigint NOT NULL,
    work_id bigint,
    code character varying NOT NULL,
    name character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    mail_sent_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone
);


--
-- Name: vod_titles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.vod_titles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: vod_titles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.vod_titles_id_seq OWNED BY public.vod_titles.id;


--
-- Name: work_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_comments (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    body character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: work_comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_comments_id_seq OWNED BY public.work_comments.id;


--
-- Name: work_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_images (
    id bigint NOT NULL,
    work_id bigint NOT NULL,
    user_id bigint NOT NULL,
    attachment_file_name character varying,
    attachment_file_size integer,
    attachment_content_type character varying,
    attachment_updated_at timestamp without time zone,
    copyright character varying DEFAULT ''::character varying NOT NULL,
    asin character varying DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    color_rgb character varying DEFAULT '255,255,255'::character varying NOT NULL,
    image_data text NOT NULL
);


--
-- Name: work_images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_images_id_seq OWNED BY public.work_images.id;


--
-- Name: work_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_records (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    title character varying DEFAULT ''::character varying,
    body text NOT NULL,
    rating_animation_state character varying,
    rating_music_state character varying,
    rating_story_state character varying,
    rating_character_state character varying,
    rating_overall_state character varying,
    likes_count integer DEFAULT 0 NOT NULL,
    impressions_count integer DEFAULT 0 NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    modified_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    oauth_application_id bigint,
    locale character varying DEFAULT 'other'::character varying NOT NULL,
    record_id bigint NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: work_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_records_id_seq OWNED BY public.work_records.id;


--
-- Name: work_taggables; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_taggables (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_tag_id bigint NOT NULL,
    description character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL
);


--
-- Name: work_taggables_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_taggables_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_taggables_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_taggables_id_seq OWNED BY public.work_taggables.id;


--
-- Name: work_taggings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_taggings (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_id bigint NOT NULL,
    work_tag_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: work_taggings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_taggings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_taggings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_taggings_id_seq OWNED BY public.work_taggings.id;


--
-- Name: work_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_tags (
    id bigint NOT NULL,
    name character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    work_taggings_count integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    locale character varying DEFAULT 'other'::character varying NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: work_tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_tags_id_seq OWNED BY public.work_tags.id;


--
-- Name: works_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.works_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: works; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.works (
    id bigint DEFAULT nextval('public.works_id_seq'::regclass) NOT NULL,
    season_id bigint,
    sc_tid integer,
    title character varying(510) NOT NULL,
    media integer NOT NULL,
    official_site_url character varying(510) DEFAULT ''::character varying NOT NULL,
    wikipedia_url character varying(510) DEFAULT ''::character varying NOT NULL,
    episodes_count integer DEFAULT 0 NOT NULL,
    watchers_count integer DEFAULT 0 NOT NULL,
    released_at date,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    twitter_username character varying(510) DEFAULT NULL::character varying,
    twitter_hashtag character varying(510) DEFAULT NULL::character varying,
    released_at_about character varying,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    number_format_id bigint,
    title_kana character varying DEFAULT ''::character varying NOT NULL,
    title_ro character varying DEFAULT ''::character varying NOT NULL,
    title_en character varying DEFAULT ''::character varying NOT NULL,
    official_site_url_en character varying DEFAULT ''::character varying NOT NULL,
    wikipedia_url_en character varying DEFAULT ''::character varying NOT NULL,
    synopsis text DEFAULT ''::text NOT NULL,
    synopsis_en text DEFAULT ''::text NOT NULL,
    synopsis_source character varying DEFAULT ''::character varying NOT NULL,
    synopsis_source_en character varying DEFAULT ''::character varying NOT NULL,
    mal_anime_id integer,
    facebook_og_image_url character varying DEFAULT ''::character varying NOT NULL,
    twitter_image_url character varying DEFAULT ''::character varying NOT NULL,
    recommended_image_url character varying DEFAULT ''::character varying NOT NULL,
    season_year integer,
    season_name integer,
    key_pv_id bigint,
    manual_episodes_count integer,
    no_episodes boolean DEFAULT false NOT NULL,
    work_records_count integer DEFAULT 0 NOT NULL,
    started_on date,
    ended_on date,
    score double precision,
    ratings_count integer DEFAULT 0 NOT NULL,
    satisfaction_rate double precision,
    records_count integer DEFAULT 0 NOT NULL,
    work_records_with_body_count integer DEFAULT 0 NOT NULL,
    start_episode_raw_number double precision DEFAULT 1.0 NOT NULL,
    deleted_at timestamp without time zone,
    title_alter character varying DEFAULT ''::character varying NOT NULL,
    title_alter_en character varying DEFAULT ''::character varying NOT NULL,
    unpublished_at timestamp without time zone
);


--
-- Name: activity_groups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activity_groups ALTER COLUMN id SET DEFAULT nextval('public.activity_groups_id_seq'::regclass);


--
-- Name: casts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casts ALTER COLUMN id SET DEFAULT nextval('public.casts_id_seq'::regclass);


--
-- Name: character_favorites id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_favorites ALTER COLUMN id SET DEFAULT nextval('public.character_favorites_id_seq'::regclass);


--
-- Name: character_images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_images ALTER COLUMN id SET DEFAULT nextval('public.character_images_id_seq'::regclass);


--
-- Name: characters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters ALTER COLUMN id SET DEFAULT nextval('public.characters_id_seq'::regclass);


--
-- Name: collection_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_items ALTER COLUMN id SET DEFAULT nextval('public.collection_items_id_seq'::regclass);


--
-- Name: collections id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections ALTER COLUMN id SET DEFAULT nextval('public.collections_id_seq'::regclass);


--
-- Name: db_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_activities ALTER COLUMN id SET DEFAULT nextval('public.db_activities_id_seq'::regclass);


--
-- Name: db_comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_comments ALTER COLUMN id SET DEFAULT nextval('public.db_comments_id_seq'::regclass);


--
-- Name: delayed_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.delayed_jobs ALTER COLUMN id SET DEFAULT nextval('public.delayed_jobs_id_seq'::regclass);


--
-- Name: email_confirmations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_confirmations ALTER COLUMN id SET DEFAULT nextval('public.email_confirmations_id_seq'::regclass);


--
-- Name: email_notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_notifications ALTER COLUMN id SET DEFAULT nextval('public.email_notifications_id_seq'::regclass);


--
-- Name: faq_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faq_categories ALTER COLUMN id SET DEFAULT nextval('public.faq_categories_id_seq'::regclass);


--
-- Name: faq_contents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faq_contents ALTER COLUMN id SET DEFAULT nextval('public.faq_contents_id_seq'::regclass);


--
-- Name: finished_tips id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.finished_tips ALTER COLUMN id SET DEFAULT nextval('public.finished_tips_id_seq'::regclass);


--
-- Name: flashes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.flashes ALTER COLUMN id SET DEFAULT nextval('public.flashes_id_seq'::regclass);


--
-- Name: forum_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_categories ALTER COLUMN id SET DEFAULT nextval('public.forum_categories_id_seq'::regclass);


--
-- Name: forum_comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_comments ALTER COLUMN id SET DEFAULT nextval('public.forum_comments_id_seq'::regclass);


--
-- Name: forum_post_participants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_post_participants ALTER COLUMN id SET DEFAULT nextval('public.forum_post_participants_id_seq'::regclass);


--
-- Name: forum_posts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_posts ALTER COLUMN id SET DEFAULT nextval('public.forum_posts_id_seq'::regclass);


--
-- Name: gumroad_subscribers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gumroad_subscribers ALTER COLUMN id SET DEFAULT nextval('public.gumroad_subscribers_id_seq'::regclass);


--
-- Name: internal_statistics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_statistics ALTER COLUMN id SET DEFAULT nextval('public.internal_statistics_id_seq'::regclass);


--
-- Name: library_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries ALTER COLUMN id SET DEFAULT nextval('public.library_entries_id_seq'::regclass);


--
-- Name: multiple_episode_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.multiple_episode_records ALTER COLUMN id SET DEFAULT nextval('public.multiple_episode_records_id_seq'::regclass);


--
-- Name: mute_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mute_users ALTER COLUMN id SET DEFAULT nextval('public.mute_users_id_seq'::regclass);


--
-- Name: number_formats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.number_formats ALTER COLUMN id SET DEFAULT nextval('public.number_formats_id_seq'::regclass);


--
-- Name: oauth_access_grants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants ALTER COLUMN id SET DEFAULT nextval('public.oauth_access_grants_id_seq'::regclass);


--
-- Name: oauth_access_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.oauth_access_tokens_id_seq'::regclass);


--
-- Name: oauth_applications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_applications ALTER COLUMN id SET DEFAULT nextval('public.oauth_applications_id_seq'::regclass);


--
-- Name: organization_favorites id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_favorites ALTER COLUMN id SET DEFAULT nextval('public.organization_favorites_id_seq'::regclass);


--
-- Name: organizations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations ALTER COLUMN id SET DEFAULT nextval('public.organizations_id_seq'::regclass);


--
-- Name: people id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.people ALTER COLUMN id SET DEFAULT nextval('public.people_id_seq'::regclass);


--
-- Name: person_favorites id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person_favorites ALTER COLUMN id SET DEFAULT nextval('public.person_favorites_id_seq'::regclass);


--
-- Name: prefectures id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prefectures ALTER COLUMN id SET DEFAULT nextval('public.prefectures_id_seq'::regclass);


--
-- Name: programs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs ALTER COLUMN id SET DEFAULT nextval('public.programs_id_seq'::regclass);


--
-- Name: reactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions ALTER COLUMN id SET DEFAULT nextval('public.reactions_id_seq'::regclass);


--
-- Name: records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records ALTER COLUMN id SET DEFAULT nextval('public.records_id_seq'::regclass);


--
-- Name: series id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series ALTER COLUMN id SET DEFAULT nextval('public.series_id_seq'::regclass);


--
-- Name: series_works id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_works ALTER COLUMN id SET DEFAULT nextval('public.series_works_id_seq'::regclass);


--
-- Name: sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions ALTER COLUMN id SET DEFAULT nextval('public.sessions_id_seq1'::regclass);


--
-- Name: settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settings ALTER COLUMN id SET DEFAULT nextval('public.settings_id_seq'::regclass);


--
-- Name: syobocal_alerts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.syobocal_alerts ALTER COLUMN id SET DEFAULT nextval('public.syobocal_alerts_id_seq'::regclass);


--
-- Name: tips id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tips ALTER COLUMN id SET DEFAULT nextval('public.tips_id_seq'::regclass);


--
-- Name: trailers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trailers ALTER COLUMN id SET DEFAULT nextval('public.trailers_id_seq'::regclass);


--
-- Name: twitter_watching_lists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.twitter_watching_lists ALTER COLUMN id SET DEFAULT nextval('public.twitter_watching_lists_id_seq'::regclass);


--
-- Name: userland_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_categories ALTER COLUMN id SET DEFAULT nextval('public.userland_categories_id_seq'::regclass);


--
-- Name: userland_project_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_project_members ALTER COLUMN id SET DEFAULT nextval('public.userland_project_members_id_seq'::regclass);


--
-- Name: userland_projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_projects ALTER COLUMN id SET DEFAULT nextval('public.userland_projects_id_seq'::regclass);


--
-- Name: vod_titles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vod_titles ALTER COLUMN id SET DEFAULT nextval('public.vod_titles_id_seq'::regclass);


--
-- Name: work_comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_comments ALTER COLUMN id SET DEFAULT nextval('public.work_comments_id_seq'::regclass);


--
-- Name: work_images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_images ALTER COLUMN id SET DEFAULT nextval('public.work_images_id_seq'::regclass);


--
-- Name: work_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records ALTER COLUMN id SET DEFAULT nextval('public.work_records_id_seq'::regclass);


--
-- Name: work_taggables id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggables ALTER COLUMN id SET DEFAULT nextval('public.work_taggables_id_seq'::regclass);


--
-- Name: work_taggings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggings ALTER COLUMN id SET DEFAULT nextval('public.work_taggings_id_seq'::regclass);


--
-- Name: work_tags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_tags ALTER COLUMN id SET DEFAULT nextval('public.work_tags_id_seq'::regclass);


--
-- Name: activities activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);


--
-- Name: activity_groups activity_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activity_groups
    ADD CONSTRAINT activity_groups_pkey PRIMARY KEY (id);


--
-- Name: ar_internal_metadata ar_internal_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ar_internal_metadata
    ADD CONSTRAINT ar_internal_metadata_pkey PRIMARY KEY (key);


--
-- Name: casts casts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casts
    ADD CONSTRAINT casts_pkey PRIMARY KEY (id);


--
-- Name: channel_groups channel_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_groups
    ADD CONSTRAINT channel_groups_pkey PRIMARY KEY (id);


--
-- Name: channel_groups channel_groups_sc_chgid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_groups
    ADD CONSTRAINT channel_groups_sc_chgid_key UNIQUE (sc_chgid);


--
-- Name: channel_works channel_works_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_works
    ADD CONSTRAINT channel_works_pkey PRIMARY KEY (id);


--
-- Name: channel_works channel_works_user_id_work_id_channel_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_works
    ADD CONSTRAINT channel_works_user_id_work_id_channel_id_key UNIQUE (user_id, work_id, channel_id);


--
-- Name: channels channels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (id);


--
-- Name: channels channels_sc_chid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_sc_chid_key UNIQUE (sc_chid);


--
-- Name: character_favorites character_favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_favorites
    ADD CONSTRAINT character_favorites_pkey PRIMARY KEY (id);


--
-- Name: character_images character_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_images
    ADD CONSTRAINT character_images_pkey PRIMARY KEY (id);


--
-- Name: characters characters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters
    ADD CONSTRAINT characters_pkey PRIMARY KEY (id);


--
-- Name: episode_records checkins_facebook_url_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT checkins_facebook_url_hash_key UNIQUE (facebook_url_hash);


--
-- Name: episode_records checkins_twitter_url_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT checkins_twitter_url_hash_key UNIQUE (twitter_url_hash);


--
-- Name: collection_items collection_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT collection_items_pkey PRIMARY KEY (id);


--
-- Name: collections collections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: db_activities db_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_activities
    ADD CONSTRAINT db_activities_pkey PRIMARY KEY (id);


--
-- Name: db_comments db_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_comments
    ADD CONSTRAINT db_comments_pkey PRIMARY KEY (id);


--
-- Name: delayed_jobs delayed_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.delayed_jobs
    ADD CONSTRAINT delayed_jobs_pkey PRIMARY KEY (id);


--
-- Name: email_confirmations email_confirmations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_confirmations
    ADD CONSTRAINT email_confirmations_pkey PRIMARY KEY (id);


--
-- Name: email_notifications email_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_notifications
    ADD CONSTRAINT email_notifications_pkey PRIMARY KEY (id);


--
-- Name: episode_records episode_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT episode_records_pkey PRIMARY KEY (id);


--
-- Name: episodes episodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_pkey PRIMARY KEY (id);


--
-- Name: episodes episodes_work_id_sc_count_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_work_id_sc_count_key UNIQUE (work_id, sc_count);


--
-- Name: faq_categories faq_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faq_categories
    ADD CONSTRAINT faq_categories_pkey PRIMARY KEY (id);


--
-- Name: faq_contents faq_contents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faq_contents
    ADD CONSTRAINT faq_contents_pkey PRIMARY KEY (id);


--
-- Name: finished_tips finished_tips_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.finished_tips
    ADD CONSTRAINT finished_tips_pkey PRIMARY KEY (id);


--
-- Name: flashes flashes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.flashes
    ADD CONSTRAINT flashes_pkey PRIMARY KEY (id);


--
-- Name: follows follows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_pkey PRIMARY KEY (id);


--
-- Name: follows follows_user_id_following_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_user_id_following_id_key UNIQUE (user_id, following_id);


--
-- Name: forum_categories forum_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_categories
    ADD CONSTRAINT forum_categories_pkey PRIMARY KEY (id);


--
-- Name: forum_comments forum_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_comments
    ADD CONSTRAINT forum_comments_pkey PRIMARY KEY (id);


--
-- Name: forum_post_participants forum_post_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_post_participants
    ADD CONSTRAINT forum_post_participants_pkey PRIMARY KEY (id);


--
-- Name: forum_posts forum_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_posts
    ADD CONSTRAINT forum_posts_pkey PRIMARY KEY (id);


--
-- Name: gumroad_subscribers gumroad_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gumroad_subscribers
    ADD CONSTRAINT gumroad_subscribers_pkey PRIMARY KEY (id);


--
-- Name: internal_statistics internal_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_statistics
    ADD CONSTRAINT internal_statistics_pkey PRIMARY KEY (id);


--
-- Name: library_entries library_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT library_entries_pkey PRIMARY KEY (id);


--
-- Name: likes likes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_pkey PRIMARY KEY (id);


--
-- Name: multiple_episode_records multiple_episode_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.multiple_episode_records
    ADD CONSTRAINT multiple_episode_records_pkey PRIMARY KEY (id);


--
-- Name: mute_users mute_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mute_users
    ADD CONSTRAINT mute_users_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: number_formats number_formats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.number_formats
    ADD CONSTRAINT number_formats_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_grants oauth_access_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants
    ADD CONSTRAINT oauth_access_grants_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_tokens oauth_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens
    ADD CONSTRAINT oauth_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: oauth_applications oauth_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_applications
    ADD CONSTRAINT oauth_applications_pkey PRIMARY KEY (id);


--
-- Name: organization_favorites organization_favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_favorites
    ADD CONSTRAINT organization_favorites_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: people people_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.people
    ADD CONSTRAINT people_pkey PRIMARY KEY (id);


--
-- Name: person_favorites person_favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person_favorites
    ADD CONSTRAINT person_favorites_pkey PRIMARY KEY (id);


--
-- Name: prefectures prefectures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prefectures
    ADD CONSTRAINT prefectures_pkey PRIMARY KEY (id);


--
-- Name: profiles profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_pkey PRIMARY KEY (id);


--
-- Name: profiles profiles_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_user_id_key UNIQUE (user_id);


--
-- Name: programs programs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs
    ADD CONSTRAINT programs_pkey PRIMARY KEY (id);


--
-- Name: providers providers_name_uid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_name_uid_key UNIQUE (name, uid);


--
-- Name: providers providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_pkey PRIMARY KEY (id);


--
-- Name: reactions reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_pkey PRIMARY KEY (id);


--
-- Name: receptions receptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.receptions
    ADD CONSTRAINT receptions_pkey PRIMARY KEY (id);


--
-- Name: receptions receptions_user_id_channel_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.receptions
    ADD CONSTRAINT receptions_user_id_channel_id_key UNIQUE (user_id, channel_id);


--
-- Name: records records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_version_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_version_key UNIQUE (version);


--
-- Name: seasons seasons_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seasons
    ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);


--
-- Name: series series_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series
    ADD CONSTRAINT series_pkey PRIMARY KEY (id);


--
-- Name: series_works series_works_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_works
    ADD CONSTRAINT series_works_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: settings settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT settings_pkey PRIMARY KEY (id);


--
-- Name: slots slots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slots
    ADD CONSTRAINT slots_pkey PRIMARY KEY (id);


--
-- Name: staffs staffs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.staffs
    ADD CONSTRAINT staffs_pkey PRIMARY KEY (id);


--
-- Name: statuses statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statuses
    ADD CONSTRAINT statuses_pkey PRIMARY KEY (id);


--
-- Name: syobocal_alerts syobocal_alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.syobocal_alerts
    ADD CONSTRAINT syobocal_alerts_pkey PRIMARY KEY (id);


--
-- Name: tips tips_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tips
    ADD CONSTRAINT tips_pkey PRIMARY KEY (id);


--
-- Name: trailers trailers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trailers
    ADD CONSTRAINT trailers_pkey PRIMARY KEY (id);


--
-- Name: twitter_bots twitter_bots_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.twitter_bots
    ADD CONSTRAINT twitter_bots_name_key UNIQUE (name);


--
-- Name: twitter_bots twitter_bots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.twitter_bots
    ADD CONSTRAINT twitter_bots_pkey PRIMARY KEY (id);


--
-- Name: twitter_watching_lists twitter_watching_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.twitter_watching_lists
    ADD CONSTRAINT twitter_watching_lists_pkey PRIMARY KEY (id);


--
-- Name: userland_categories userland_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_categories
    ADD CONSTRAINT userland_categories_pkey PRIMARY KEY (id);


--
-- Name: userland_project_members userland_project_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_project_members
    ADD CONSTRAINT userland_project_members_pkey PRIMARY KEY (id);


--
-- Name: userland_projects userland_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_projects
    ADD CONSTRAINT userland_projects_pkey PRIMARY KEY (id);


--
-- Name: users users_confirmation_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_confirmation_token_key UNIQUE (confirmation_token);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: versions versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.versions
    ADD CONSTRAINT versions_pkey PRIMARY KEY (id);


--
-- Name: vod_titles vod_titles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vod_titles
    ADD CONSTRAINT vod_titles_pkey PRIMARY KEY (id);


--
-- Name: work_comments work_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_comments
    ADD CONSTRAINT work_comments_pkey PRIMARY KEY (id);


--
-- Name: work_images work_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_images
    ADD CONSTRAINT work_images_pkey PRIMARY KEY (id);


--
-- Name: work_records work_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records
    ADD CONSTRAINT work_records_pkey PRIMARY KEY (id);


--
-- Name: work_taggables work_taggables_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggables
    ADD CONSTRAINT work_taggables_pkey PRIMARY KEY (id);


--
-- Name: work_taggings work_taggings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggings
    ADD CONSTRAINT work_taggings_pkey PRIMARY KEY (id);


--
-- Name: work_tags work_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_tags
    ADD CONSTRAINT work_tags_pkey PRIMARY KEY (id);


--
-- Name: works works_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.works
    ADD CONSTRAINT works_pkey PRIMARY KEY (id);


--
-- Name: activities_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_user_id_idx ON public.activities USING btree (user_id);


--
-- Name: channel_works_channel_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX channel_works_channel_id_idx ON public.channel_works USING btree (channel_id);


--
-- Name: channel_works_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX channel_works_user_id_idx ON public.channel_works USING btree (user_id);


--
-- Name: channel_works_work_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX channel_works_work_id_idx ON public.channel_works USING btree (work_id);


--
-- Name: channels_channel_group_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX channels_channel_group_id_idx ON public.channels USING btree (channel_group_id);


--
-- Name: checkins_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX checkins_user_id_idx ON public.episode_records USING btree (user_id);


--
-- Name: comments_checkin_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_checkin_id_idx ON public.comments USING btree (episode_record_id);


--
-- Name: comments_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_user_id_idx ON public.comments USING btree (user_id);


--
-- Name: delayed_jobs_priority; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX delayed_jobs_priority ON public.delayed_jobs USING btree (priority, run_at);


--
-- Name: episodes_work_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX episodes_work_id_idx ON public.episodes USING btree (work_id);


--
-- Name: follows_following_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX follows_following_id_idx ON public.follows USING btree (following_id);


--
-- Name: follows_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX follows_user_id_idx ON public.follows USING btree (user_id);


--
-- Name: index_activities_on_activity_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_activity_group_id ON public.activities USING btree (activity_group_id);


--
-- Name: index_activities_on_activity_group_id_and_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_activity_group_id_and_created_at ON public.activities USING btree (activity_group_id, created_at);


--
-- Name: index_activities_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_created_at ON public.activities USING btree (created_at);


--
-- Name: index_activities_on_episode_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_episode_id ON public.activities USING btree (episode_id);


--
-- Name: index_activities_on_episode_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_episode_record_id ON public.activities USING btree (episode_record_id);


--
-- Name: index_activities_on_multiple_episode_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_multiple_episode_record_id ON public.activities USING btree (multiple_episode_record_id);


--
-- Name: index_activities_on_status_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_status_id ON public.activities USING btree (status_id);


--
-- Name: index_activities_on_trackable_id_and_trackable_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_trackable_id_and_trackable_type ON public.activities USING btree (trackable_id, trackable_type);


--
-- Name: index_activities_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_work_id ON public.activities USING btree (work_id);


--
-- Name: index_activities_on_work_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activities_on_work_record_id ON public.activities USING btree (work_record_id);


--
-- Name: index_activity_groups_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activity_groups_on_created_at ON public.activity_groups USING btree (created_at);


--
-- Name: index_activity_groups_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_activity_groups_on_user_id ON public.activity_groups USING btree (user_id);


--
-- Name: index_casts_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_aasm_state ON public.casts USING btree (aasm_state);


--
-- Name: index_casts_on_character_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_character_id ON public.casts USING btree (character_id);


--
-- Name: index_casts_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_deleted_at ON public.casts USING btree (deleted_at);


--
-- Name: index_casts_on_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_person_id ON public.casts USING btree (person_id);


--
-- Name: index_casts_on_sort_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_sort_number ON public.casts USING btree (sort_number);


--
-- Name: index_casts_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_unpublished_at ON public.casts USING btree (unpublished_at);


--
-- Name: index_casts_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_casts_on_work_id ON public.casts USING btree (work_id);


--
-- Name: index_channel_groups_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_channel_groups_on_unpublished_at ON public.channel_groups USING btree (unpublished_at);


--
-- Name: index_channels_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_channels_on_deleted_at ON public.channels USING btree (deleted_at);


--
-- Name: index_channels_on_sort_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_channels_on_sort_number ON public.channels USING btree (sort_number);


--
-- Name: index_channels_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_channels_on_unpublished_at ON public.channels USING btree (unpublished_at);


--
-- Name: index_channels_on_vod; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_channels_on_vod ON public.channels USING btree (vod);


--
-- Name: index_character_favorites_on_character_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_character_favorites_on_character_id ON public.character_favorites USING btree (character_id);


--
-- Name: index_character_favorites_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_character_favorites_on_user_id ON public.character_favorites USING btree (user_id);


--
-- Name: index_character_favorites_on_user_id_and_character_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_character_favorites_on_user_id_and_character_id ON public.character_favorites USING btree (user_id, character_id);


--
-- Name: index_character_images_on_character_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_character_images_on_character_id ON public.character_images USING btree (character_id);


--
-- Name: index_character_images_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_character_images_on_user_id ON public.character_images USING btree (user_id);


--
-- Name: index_characters_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_characters_on_deleted_at ON public.characters USING btree (deleted_at);


--
-- Name: index_characters_on_favorite_users_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_characters_on_favorite_users_count ON public.characters USING btree (favorite_users_count);


--
-- Name: index_characters_on_name_and_series_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_characters_on_name_and_series_id ON public.characters USING btree (name, series_id);


--
-- Name: index_characters_on_series_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_characters_on_series_id ON public.characters USING btree (series_id);


--
-- Name: index_characters_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_characters_on_unpublished_at ON public.characters USING btree (unpublished_at);


--
-- Name: index_collection_items_on_collection_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collection_items_on_collection_id ON public.collection_items USING btree (collection_id);


--
-- Name: index_collection_items_on_collection_id_and_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_collection_items_on_collection_id_and_work_id ON public.collection_items USING btree (collection_id, work_id);


--
-- Name: index_collection_items_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collection_items_on_deleted_at ON public.collection_items USING btree (deleted_at);


--
-- Name: index_collection_items_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collection_items_on_user_id ON public.collection_items USING btree (user_id);


--
-- Name: index_collection_items_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collection_items_on_work_id ON public.collection_items USING btree (work_id);


--
-- Name: index_collections_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collections_on_deleted_at ON public.collections USING btree (deleted_at);


--
-- Name: index_collections_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_collections_on_user_id ON public.collections USING btree (user_id);


--
-- Name: index_collections_on_user_id_and_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_collections_on_user_id_and_name ON public.collections USING btree (user_id, name);


--
-- Name: index_comments_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_locale ON public.comments USING btree (locale);


--
-- Name: index_comments_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_work_id ON public.comments USING btree (work_id);


--
-- Name: index_db_activities_on_object_id_and_object_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_activities_on_object_id_and_object_type ON public.db_activities USING btree (object_id, object_type);


--
-- Name: index_db_activities_on_root_resource_id_and_root_resource_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_activities_on_root_resource_id_and_root_resource_type ON public.db_activities USING btree (root_resource_id, root_resource_type);


--
-- Name: index_db_activities_on_trackable_id_and_trackable_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_activities_on_trackable_id_and_trackable_type ON public.db_activities USING btree (trackable_id, trackable_type);


--
-- Name: index_db_comments_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_comments_on_locale ON public.db_comments USING btree (locale);


--
-- Name: index_db_comments_on_resource_id_and_resource_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_comments_on_resource_id_and_resource_type ON public.db_comments USING btree (resource_id, resource_type);


--
-- Name: index_db_comments_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_db_comments_on_user_id ON public.db_comments USING btree (user_id);


--
-- Name: index_email_confirmations_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_email_confirmations_on_token ON public.email_confirmations USING btree (token);


--
-- Name: index_email_confirmations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_email_confirmations_on_user_id ON public.email_confirmations USING btree (user_id);


--
-- Name: index_email_notifications_on_unsubscription_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_email_notifications_on_unsubscription_key ON public.email_notifications USING btree (unsubscription_key);


--
-- Name: index_email_notifications_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_email_notifications_on_user_id ON public.email_notifications USING btree (user_id);


--
-- Name: index_episode_records_on_episode_id_and_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_episode_id_and_deleted_at ON public.episode_records USING btree (episode_id, deleted_at);


--
-- Name: index_episode_records_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_locale ON public.episode_records USING btree (locale);


--
-- Name: index_episode_records_on_multiple_episode_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_multiple_episode_record_id ON public.episode_records USING btree (multiple_episode_record_id);


--
-- Name: index_episode_records_on_oauth_application_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_oauth_application_id ON public.episode_records USING btree (oauth_application_id);


--
-- Name: index_episode_records_on_rating_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_rating_state ON public.episode_records USING btree (rating_state);


--
-- Name: index_episode_records_on_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_episode_records_on_record_id ON public.episode_records USING btree (record_id);


--
-- Name: index_episode_records_on_review_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_review_id ON public.episode_records USING btree (review_id);


--
-- Name: index_episode_records_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episode_records_on_work_id ON public.episode_records USING btree (work_id);


--
-- Name: index_episodes_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_aasm_state ON public.episodes USING btree (aasm_state);


--
-- Name: index_episodes_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_deleted_at ON public.episodes USING btree (deleted_at);


--
-- Name: index_episodes_on_prev_episode_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_prev_episode_id ON public.episodes USING btree (prev_episode_id);


--
-- Name: index_episodes_on_ratings_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_ratings_count ON public.episodes USING btree (ratings_count);


--
-- Name: index_episodes_on_satisfaction_rate; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_satisfaction_rate ON public.episodes USING btree (satisfaction_rate);


--
-- Name: index_episodes_on_satisfaction_rate_and_ratings_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_satisfaction_rate_and_ratings_count ON public.episodes USING btree (satisfaction_rate, ratings_count);


--
-- Name: index_episodes_on_score; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_score ON public.episodes USING btree (score);


--
-- Name: index_episodes_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_episodes_on_unpublished_at ON public.episodes USING btree (unpublished_at);


--
-- Name: index_faq_categories_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_faq_categories_on_deleted_at ON public.faq_categories USING btree (deleted_at);


--
-- Name: index_faq_categories_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_faq_categories_on_locale ON public.faq_categories USING btree (locale);


--
-- Name: index_faq_contents_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_faq_contents_on_deleted_at ON public.faq_contents USING btree (deleted_at);


--
-- Name: index_faq_contents_on_faq_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_faq_contents_on_faq_category_id ON public.faq_contents USING btree (faq_category_id);


--
-- Name: index_faq_contents_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_faq_contents_on_locale ON public.faq_contents USING btree (locale);


--
-- Name: index_finished_tips_on_user_id_and_tip_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_finished_tips_on_user_id_and_tip_id ON public.finished_tips USING btree (user_id, tip_id);


--
-- Name: index_flashes_on_client_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_flashes_on_client_uuid ON public.flashes USING btree (client_uuid);


--
-- Name: index_forum_categories_on_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_forum_categories_on_slug ON public.forum_categories USING btree (slug);


--
-- Name: index_forum_comments_on_forum_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_comments_on_forum_post_id ON public.forum_comments USING btree (forum_post_id);


--
-- Name: index_forum_comments_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_comments_on_locale ON public.forum_comments USING btree (locale);


--
-- Name: index_forum_comments_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_comments_on_user_id ON public.forum_comments USING btree (user_id);


--
-- Name: index_forum_post_participants_on_forum_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_post_participants_on_forum_post_id ON public.forum_post_participants USING btree (forum_post_id);


--
-- Name: index_forum_post_participants_on_forum_post_id_and_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_forum_post_participants_on_forum_post_id_and_user_id ON public.forum_post_participants USING btree (forum_post_id, user_id);


--
-- Name: index_forum_post_participants_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_post_participants_on_user_id ON public.forum_post_participants USING btree (user_id);


--
-- Name: index_forum_posts_on_forum_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_posts_on_forum_category_id ON public.forum_posts USING btree (forum_category_id);


--
-- Name: index_forum_posts_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_posts_on_locale ON public.forum_posts USING btree (locale);


--
-- Name: index_forum_posts_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_forum_posts_on_user_id ON public.forum_posts USING btree (user_id);


--
-- Name: index_gumroad_subscribers_on_gumroad_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_gumroad_subscribers_on_gumroad_id ON public.gumroad_subscribers USING btree (gumroad_id);


--
-- Name: index_gumroad_subscribers_on_gumroad_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_gumroad_subscribers_on_gumroad_product_id ON public.gumroad_subscribers USING btree (gumroad_product_id);


--
-- Name: index_gumroad_subscribers_on_gumroad_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_gumroad_subscribers_on_gumroad_user_id ON public.gumroad_subscribers USING btree (gumroad_user_id);


--
-- Name: index_internal_statistics_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_internal_statistics_on_key ON public.internal_statistics USING btree (key);


--
-- Name: index_internal_statistics_on_key_and_date; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_internal_statistics_on_key_and_date ON public.internal_statistics USING btree (key, date);


--
-- Name: index_library_entries_on_next_episode_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_next_episode_id ON public.library_entries USING btree (next_episode_id);


--
-- Name: index_library_entries_on_next_slot_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_next_slot_id ON public.library_entries USING btree (next_slot_id);


--
-- Name: index_library_entries_on_program_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_program_id ON public.library_entries USING btree (program_id);


--
-- Name: index_library_entries_on_status_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_status_id ON public.library_entries USING btree (status_id);


--
-- Name: index_library_entries_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_user_id ON public.library_entries USING btree (user_id);


--
-- Name: index_library_entries_on_user_id_and_position; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_user_id_and_position ON public.library_entries USING btree (user_id, "position");


--
-- Name: index_library_entries_on_user_id_and_program_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_library_entries_on_user_id_and_program_id ON public.library_entries USING btree (user_id, program_id);


--
-- Name: index_library_entries_on_user_id_and_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_library_entries_on_user_id_and_work_id ON public.library_entries USING btree (user_id, work_id);


--
-- Name: index_library_entries_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_library_entries_on_work_id ON public.library_entries USING btree (work_id);


--
-- Name: index_multiple_episode_records_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_multiple_episode_records_on_user_id ON public.multiple_episode_records USING btree (user_id);


--
-- Name: index_multiple_episode_records_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_multiple_episode_records_on_work_id ON public.multiple_episode_records USING btree (work_id);


--
-- Name: index_mute_users_on_muted_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_mute_users_on_muted_user_id ON public.mute_users USING btree (muted_user_id);


--
-- Name: index_mute_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_mute_users_on_user_id ON public.mute_users USING btree (user_id);


--
-- Name: index_mute_users_on_user_id_and_muted_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_mute_users_on_user_id_and_muted_user_id ON public.mute_users USING btree (user_id, muted_user_id);


--
-- Name: index_number_formats_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_number_formats_on_name ON public.number_formats USING btree (name);


--
-- Name: index_oauth_access_grants_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_grants_on_token ON public.oauth_access_grants USING btree (token);


--
-- Name: index_oauth_access_tokens_on_refresh_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_refresh_token ON public.oauth_access_tokens USING btree (refresh_token);


--
-- Name: index_oauth_access_tokens_on_resource_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_access_tokens_on_resource_owner_id ON public.oauth_access_tokens USING btree (resource_owner_id);


--
-- Name: index_oauth_access_tokens_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_token ON public.oauth_access_tokens USING btree (token);


--
-- Name: index_oauth_applications_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_applications_on_deleted_at ON public.oauth_applications USING btree (deleted_at);


--
-- Name: index_oauth_applications_on_owner_id_and_owner_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_applications_on_owner_id_and_owner_type ON public.oauth_applications USING btree (owner_id, owner_type);


--
-- Name: index_oauth_applications_on_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_applications_on_uid ON public.oauth_applications USING btree (uid);


--
-- Name: index_organization_favorites_on_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organization_favorites_on_organization_id ON public.organization_favorites USING btree (organization_id);


--
-- Name: index_organization_favorites_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organization_favorites_on_user_id ON public.organization_favorites USING btree (user_id);


--
-- Name: index_organization_favorites_on_user_id_and_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_organization_favorites_on_user_id_and_organization_id ON public.organization_favorites USING btree (user_id, organization_id);


--
-- Name: index_organization_favorites_on_watched_works_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organization_favorites_on_watched_works_count ON public.organization_favorites USING btree (watched_works_count);


--
-- Name: index_organizations_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organizations_on_aasm_state ON public.organizations USING btree (aasm_state);


--
-- Name: index_organizations_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organizations_on_deleted_at ON public.organizations USING btree (deleted_at);


--
-- Name: index_organizations_on_favorite_users_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organizations_on_favorite_users_count ON public.organizations USING btree (favorite_users_count);


--
-- Name: index_organizations_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_organizations_on_name ON public.organizations USING btree (name);


--
-- Name: index_organizations_on_staffs_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organizations_on_staffs_count ON public.organizations USING btree (staffs_count);


--
-- Name: index_organizations_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organizations_on_unpublished_at ON public.organizations USING btree (unpublished_at);


--
-- Name: index_people_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_aasm_state ON public.people USING btree (aasm_state);


--
-- Name: index_people_on_casts_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_casts_count ON public.people USING btree (casts_count);


--
-- Name: index_people_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_deleted_at ON public.people USING btree (deleted_at);


--
-- Name: index_people_on_favorite_users_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_favorite_users_count ON public.people USING btree (favorite_users_count);


--
-- Name: index_people_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_people_on_name ON public.people USING btree (name);


--
-- Name: index_people_on_prefecture_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_prefecture_id ON public.people USING btree (prefecture_id);


--
-- Name: index_people_on_staffs_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_staffs_count ON public.people USING btree (staffs_count);


--
-- Name: index_people_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_people_on_unpublished_at ON public.people USING btree (unpublished_at);


--
-- Name: index_person_favorites_on_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_person_favorites_on_person_id ON public.person_favorites USING btree (person_id);


--
-- Name: index_person_favorites_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_person_favorites_on_user_id ON public.person_favorites USING btree (user_id);


--
-- Name: index_person_favorites_on_user_id_and_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_person_favorites_on_user_id_and_person_id ON public.person_favorites USING btree (user_id, person_id);


--
-- Name: index_person_favorites_on_watched_works_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_person_favorites_on_watched_works_count ON public.person_favorites USING btree (watched_works_count);


--
-- Name: index_prefectures_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_prefectures_on_name ON public.prefectures USING btree (name);


--
-- Name: index_programs_on_channel_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_programs_on_channel_id ON public.programs USING btree (channel_id);


--
-- Name: index_programs_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_programs_on_deleted_at ON public.programs USING btree (deleted_at);


--
-- Name: index_programs_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_programs_on_unpublished_at ON public.programs USING btree (unpublished_at);


--
-- Name: index_programs_on_vod_title_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_programs_on_vod_title_code ON public.programs USING btree (vod_title_code);


--
-- Name: index_programs_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_programs_on_work_id ON public.programs USING btree (work_id);


--
-- Name: index_reactions_on_collection_item_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_collection_item_id ON public.reactions USING btree (collection_item_id);


--
-- Name: index_reactions_on_target_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_target_user_id ON public.reactions USING btree (target_user_id);


--
-- Name: index_reactions_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_user_id ON public.reactions USING btree (user_id);


--
-- Name: index_records_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_records_on_deleted_at ON public.records USING btree (deleted_at);


--
-- Name: index_records_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_records_on_user_id ON public.records USING btree (user_id);


--
-- Name: index_records_on_watched_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_records_on_watched_at ON public.records USING btree (watched_at);


--
-- Name: index_records_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_records_on_work_id ON public.records USING btree (work_id);


--
-- Name: index_seasons_on_sort_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_seasons_on_sort_number ON public.seasons USING btree (sort_number);


--
-- Name: index_seasons_on_year; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_seasons_on_year ON public.seasons USING btree (year);


--
-- Name: index_seasons_on_year_and_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_seasons_on_year_and_name ON public.seasons USING btree (year, name);


--
-- Name: index_series_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_on_deleted_at ON public.series USING btree (deleted_at);


--
-- Name: index_series_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_series_on_name ON public.series USING btree (name);


--
-- Name: index_series_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_on_unpublished_at ON public.series USING btree (unpublished_at);


--
-- Name: index_series_works_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_works_on_deleted_at ON public.series_works USING btree (deleted_at);


--
-- Name: index_series_works_on_series_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_works_on_series_id ON public.series_works USING btree (series_id);


--
-- Name: index_series_works_on_series_id_and_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_series_works_on_series_id_and_work_id ON public.series_works USING btree (series_id, work_id);


--
-- Name: index_series_works_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_works_on_unpublished_at ON public.series_works USING btree (unpublished_at);


--
-- Name: index_series_works_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_series_works_on_work_id ON public.series_works USING btree (work_id);


--
-- Name: index_sessions_on_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_sessions_on_session_id ON public.sessions USING btree (session_id);


--
-- Name: index_sessions_on_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sessions_on_updated_at ON public.sessions USING btree (updated_at);


--
-- Name: index_settings_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_settings_on_user_id ON public.settings USING btree (user_id);


--
-- Name: index_slots_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_slots_on_aasm_state ON public.slots USING btree (aasm_state);


--
-- Name: index_slots_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_slots_on_deleted_at ON public.slots USING btree (deleted_at);


--
-- Name: index_slots_on_program_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_slots_on_program_id ON public.slots USING btree (program_id);


--
-- Name: index_slots_on_program_id_and_episode_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_slots_on_program_id_and_episode_id ON public.slots USING btree (program_id, episode_id);


--
-- Name: index_slots_on_program_id_and_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_slots_on_program_id_and_number ON public.slots USING btree (program_id, number);


--
-- Name: index_slots_on_sc_pid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_slots_on_sc_pid ON public.slots USING btree (sc_pid);


--
-- Name: index_slots_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_slots_on_unpublished_at ON public.slots USING btree (unpublished_at);


--
-- Name: index_staffs_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_aasm_state ON public.staffs USING btree (aasm_state);


--
-- Name: index_staffs_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_deleted_at ON public.staffs USING btree (deleted_at);


--
-- Name: index_staffs_on_resource_id_and_resource_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_resource_id_and_resource_type ON public.staffs USING btree (resource_id, resource_type);


--
-- Name: index_staffs_on_sort_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_sort_number ON public.staffs USING btree (sort_number);


--
-- Name: index_staffs_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_unpublished_at ON public.staffs USING btree (unpublished_at);


--
-- Name: index_staffs_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_staffs_on_work_id ON public.staffs USING btree (work_id);


--
-- Name: index_statuses_on_oauth_application_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_statuses_on_oauth_application_id ON public.statuses USING btree (oauth_application_id);


--
-- Name: index_syobocal_alerts_on_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_syobocal_alerts_on_kind ON public.syobocal_alerts USING btree (kind);


--
-- Name: index_syobocal_alerts_on_sc_prog_item_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_syobocal_alerts_on_sc_prog_item_id ON public.syobocal_alerts USING btree (sc_prog_item_id);


--
-- Name: index_tips_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tips_on_locale ON public.tips USING btree (locale);


--
-- Name: index_tips_on_slug_and_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_tips_on_slug_and_locale ON public.tips USING btree (slug, locale);


--
-- Name: index_trailers_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trailers_on_deleted_at ON public.trailers USING btree (deleted_at);


--
-- Name: index_trailers_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trailers_on_unpublished_at ON public.trailers USING btree (unpublished_at);


--
-- Name: index_trailers_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trailers_on_work_id ON public.trailers USING btree (work_id);


--
-- Name: index_userland_pm_on_uid_and_userland_pid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_userland_pm_on_uid_and_userland_pid ON public.userland_project_members USING btree (user_id, userland_project_id);


--
-- Name: index_userland_project_members_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_userland_project_members_on_user_id ON public.userland_project_members USING btree (user_id);


--
-- Name: index_userland_project_members_on_userland_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_userland_project_members_on_userland_project_id ON public.userland_project_members USING btree (userland_project_id);


--
-- Name: index_userland_projects_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_userland_projects_on_locale ON public.userland_projects USING btree (locale);


--
-- Name: index_userland_projects_on_userland_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_userland_projects_on_userland_category_id ON public.userland_projects USING btree (userland_category_id);


--
-- Name: index_users_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_aasm_state ON public.users USING btree (aasm_state);


--
-- Name: index_users_on_allowed_locales; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_allowed_locales ON public.users USING gin (allowed_locales);


--
-- Name: index_users_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: index_users_on_gumroad_subscriber_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_gumroad_subscriber_id ON public.users USING btree (gumroad_subscriber_id);


--
-- Name: index_vod_titles_on_channel_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_vod_titles_on_channel_id ON public.vod_titles USING btree (channel_id);


--
-- Name: index_vod_titles_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_vod_titles_on_deleted_at ON public.vod_titles USING btree (deleted_at);


--
-- Name: index_vod_titles_on_mail_sent_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_vod_titles_on_mail_sent_at ON public.vod_titles USING btree (mail_sent_at);


--
-- Name: index_vod_titles_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_vod_titles_on_unpublished_at ON public.vod_titles USING btree (unpublished_at);


--
-- Name: index_vod_titles_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_vod_titles_on_work_id ON public.vod_titles USING btree (work_id);


--
-- Name: index_work_comments_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_comments_on_locale ON public.work_comments USING btree (locale);


--
-- Name: index_work_comments_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_comments_on_user_id ON public.work_comments USING btree (user_id);


--
-- Name: index_work_comments_on_user_id_and_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_work_comments_on_user_id_and_work_id ON public.work_comments USING btree (user_id, work_id);


--
-- Name: index_work_comments_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_comments_on_work_id ON public.work_comments USING btree (work_id);


--
-- Name: index_work_images_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_images_on_user_id ON public.work_images USING btree (user_id);


--
-- Name: index_work_images_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_images_on_work_id ON public.work_images USING btree (work_id);


--
-- Name: index_work_records_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_records_on_deleted_at ON public.work_records USING btree (deleted_at);


--
-- Name: index_work_records_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_records_on_locale ON public.work_records USING btree (locale);


--
-- Name: index_work_records_on_oauth_application_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_records_on_oauth_application_id ON public.work_records USING btree (oauth_application_id);


--
-- Name: index_work_records_on_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_work_records_on_record_id ON public.work_records USING btree (record_id);


--
-- Name: index_work_records_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_records_on_user_id ON public.work_records USING btree (user_id);


--
-- Name: index_work_records_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_records_on_work_id ON public.work_records USING btree (work_id);


--
-- Name: index_work_taggables_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggables_on_locale ON public.work_taggables USING btree (locale);


--
-- Name: index_work_taggables_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggables_on_user_id ON public.work_taggables USING btree (user_id);


--
-- Name: index_work_taggables_on_user_id_and_work_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_work_taggables_on_user_id_and_work_tag_id ON public.work_taggables USING btree (user_id, work_tag_id);


--
-- Name: index_work_taggables_on_work_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggables_on_work_tag_id ON public.work_taggables USING btree (work_tag_id);


--
-- Name: index_work_taggings_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggings_on_user_id ON public.work_taggings USING btree (user_id);


--
-- Name: index_work_taggings_on_user_id_and_work_id_and_work_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_work_taggings_on_user_id_and_work_id_and_work_tag_id ON public.work_taggings USING btree (user_id, work_id, work_tag_id);


--
-- Name: index_work_taggings_on_work_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggings_on_work_id ON public.work_taggings USING btree (work_id);


--
-- Name: index_work_taggings_on_work_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_taggings_on_work_tag_id ON public.work_taggings USING btree (work_tag_id);


--
-- Name: index_work_tags_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_tags_on_deleted_at ON public.work_tags USING btree (deleted_at);


--
-- Name: index_work_tags_on_locale; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_tags_on_locale ON public.work_tags USING btree (locale);


--
-- Name: index_work_tags_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_work_tags_on_name ON public.work_tags USING btree (name);


--
-- Name: index_work_tags_on_work_taggings_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_work_tags_on_work_taggings_count ON public.work_tags USING btree (work_taggings_count);


--
-- Name: index_works_on_aasm_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_aasm_state ON public.works USING btree (aasm_state);


--
-- Name: index_works_on_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_deleted_at ON public.works USING btree (deleted_at);


--
-- Name: index_works_on_key_pv_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_key_pv_id ON public.works USING btree (key_pv_id);


--
-- Name: index_works_on_number_format_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_number_format_id ON public.works USING btree (number_format_id);


--
-- Name: index_works_on_ratings_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_ratings_count ON public.works USING btree (ratings_count);


--
-- Name: index_works_on_satisfaction_rate; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_satisfaction_rate ON public.works USING btree (satisfaction_rate);


--
-- Name: index_works_on_satisfaction_rate_and_ratings_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_satisfaction_rate_and_ratings_count ON public.works USING btree (satisfaction_rate, ratings_count);


--
-- Name: index_works_on_score; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_score ON public.works USING btree (score);


--
-- Name: index_works_on_season_year; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_season_year ON public.works USING btree (season_year);


--
-- Name: index_works_on_season_year_and_season_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_season_year_and_season_name ON public.works USING btree (season_year, season_name);


--
-- Name: index_works_on_unpublished_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_works_on_unpublished_at ON public.works USING btree (unpublished_at);


--
-- Name: likes_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX likes_user_id_idx ON public.likes USING btree (user_id);


--
-- Name: notifications_action_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_action_user_id_idx ON public.notifications USING btree (action_user_id);


--
-- Name: notifications_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_user_id_idx ON public.notifications USING btree (user_id);


--
-- Name: profiles_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX profiles_user_id_idx ON public.profiles USING btree (user_id);


--
-- Name: programs_channel_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX programs_channel_id_idx ON public.slots USING btree (channel_id);


--
-- Name: programs_episode_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX programs_episode_id_idx ON public.slots USING btree (episode_id);


--
-- Name: programs_work_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX programs_work_id_idx ON public.slots USING btree (work_id);


--
-- Name: providers_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX providers_user_id_idx ON public.providers USING btree (user_id);


--
-- Name: receptions_channel_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX receptions_channel_id_idx ON public.receptions USING btree (channel_id);


--
-- Name: receptions_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX receptions_user_id_idx ON public.receptions USING btree (user_id);


--
-- Name: statuses_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX statuses_user_id_idx ON public.statuses USING btree (user_id);


--
-- Name: statuses_work_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX statuses_work_id_idx ON public.statuses USING btree (work_id);


--
-- Name: works_season_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX works_season_id_idx ON public.works USING btree (season_id);


--
-- Name: activities activities_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: channel_works channel_works_channel_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_works
    ADD CONSTRAINT channel_works_channel_id_fk FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: channel_works channel_works_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_works
    ADD CONSTRAINT channel_works_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: channel_works channel_works_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_works
    ADD CONSTRAINT channel_works_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: channels channels_channel_group_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_channel_group_id_fk FOREIGN KEY (channel_group_id) REFERENCES public.channel_groups(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: episode_records checkins_episode_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT checkins_episode_id_fk FOREIGN KEY (episode_id) REFERENCES public.episodes(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: episode_records checkins_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT checkins_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: episode_records checkins_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT checkins_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: comments comments_checkin_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_checkin_id_fk FOREIGN KEY (episode_record_id) REFERENCES public.episode_records(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: comments comments_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: episodes episodes_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: finished_tips finished_tips_tip_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.finished_tips
    ADD CONSTRAINT finished_tips_tip_id_fk FOREIGN KEY (tip_id) REFERENCES public.tips(id) ON DELETE CASCADE;


--
-- Name: finished_tips finished_tips_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.finished_tips
    ADD CONSTRAINT finished_tips_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: work_comments fk_rails_00410fd248; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_comments
    ADD CONSTRAINT fk_rails_00410fd248 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: character_images fk_rails_057f9798e7; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_images
    ADD CONSTRAINT fk_rails_057f9798e7 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: comments fk_rails_09d346abb6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_rails_09d346abb6 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: series_works fk_rails_0b7ef06239; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_works
    ADD CONSTRAINT fk_rails_0b7ef06239 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: work_taggings fk_rails_0bc79546ba; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggings
    ADD CONSTRAINT fk_rails_0bc79546ba FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: db_comments fk_rails_179d1443d6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_comments
    ADD CONSTRAINT fk_rails_179d1443d6 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: activities fk_rails_24160df6bb; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_24160df6bb FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: records fk_rails_27f794c2d6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT fk_rails_27f794c2d6 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: vod_titles fk_rails_2ae1a1186c; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vod_titles
    ADD CONSTRAINT fk_rails_2ae1a1186c FOREIGN KEY (channel_id) REFERENCES public.channels(id);


--
-- Name: collection_items fk_rails_31b8b5e78c; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT fk_rails_31b8b5e78c FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: oauth_access_grants fk_rails_330c32d8d9; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants
    ADD CONSTRAINT fk_rails_330c32d8d9 FOREIGN KEY (resource_owner_id) REFERENCES public.users(id);


--
-- Name: work_taggings fk_rails_3431822583; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggings
    ADD CONSTRAINT fk_rails_3431822583 FOREIGN KEY (work_tag_id) REFERENCES public.work_tags(id);


--
-- Name: userland_project_members fk_rails_39320b176a; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_project_members
    ADD CONSTRAINT fk_rails_39320b176a FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: staffs fk_rails_39944239d2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.staffs
    ADD CONSTRAINT fk_rails_39944239d2 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: work_images fk_rails_3bad625b28; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_images
    ADD CONSTRAINT fk_rails_3bad625b28 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: person_favorites fk_rails_3bf5fa6ba3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person_favorites
    ADD CONSTRAINT fk_rails_3bf5fa6ba3 FOREIGN KEY (person_id) REFERENCES public.people(id);


--
-- Name: work_taggables fk_rails_3fbf41384e; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggables
    ADD CONSTRAINT fk_rails_3fbf41384e FOREIGN KEY (work_tag_id) REFERENCES public.work_tags(id);


--
-- Name: works fk_rails_41b1e89600; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.works
    ADD CONSTRAINT fk_rails_41b1e89600 FOREIGN KEY (number_format_id) REFERENCES public.number_formats(id);


--
-- Name: email_confirmations fk_rails_422b33d86c; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_confirmations
    ADD CONSTRAINT fk_rails_422b33d86c FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: multiple_episode_records fk_rails_43033bfda3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.multiple_episode_records
    ADD CONSTRAINT fk_rails_43033bfda3 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: library_entries fk_rails_431d8cd1d4; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_431d8cd1d4 FOREIGN KEY (next_episode_id) REFERENCES public.episodes(id);


--
-- Name: people fk_rails_456207bc59; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.people
    ADD CONSTRAINT fk_rails_456207bc59 FOREIGN KEY (prefecture_id) REFERENCES public.prefectures(id);


--
-- Name: settings fk_rails_459cc76823; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT fk_rails_459cc76823 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: character_favorites fk_rails_480f5306c2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_favorites
    ADD CONSTRAINT fk_rails_480f5306c2 FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: organization_favorites fk_rails_4b59cb0180; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_favorites
    ADD CONSTRAINT fk_rails_4b59cb0180 FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: activities fk_rails_4ef7271728; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_4ef7271728 FOREIGN KEY (activity_group_id) REFERENCES public.activity_groups(id);


--
-- Name: activities fk_rails_4f614ccd13; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_4f614ccd13 FOREIGN KEY (status_id) REFERENCES public.statuses(id);


--
-- Name: programs fk_rails_513ea7b8c6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs
    ADD CONSTRAINT fk_rails_513ea7b8c6 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: reactions fk_rails_53df387e77; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT fk_rails_53df387e77 FOREIGN KEY (collection_item_id) REFERENCES public.collection_items(id);


--
-- Name: library_entries fk_rails_561bebe304; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_561bebe304 FOREIGN KEY (next_slot_id) REFERENCES public.slots(id);


--
-- Name: trailers fk_rails_5751118f69; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trailers
    ADD CONSTRAINT fk_rails_5751118f69 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: db_activities fk_rails_5c5f39c67f; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.db_activities
    ADD CONSTRAINT fk_rails_5c5f39c67f FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: casts fk_rails_5cea49da53; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casts
    ADD CONSTRAINT fk_rails_5cea49da53 FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: email_notifications fk_rails_5ea7498254; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_notifications
    ADD CONSTRAINT fk_rails_5ea7498254 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: work_taggings fk_rails_61b358478b; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggings
    ADD CONSTRAINT fk_rails_61b358478b FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: person_favorites fk_rails_623cbf8cc5; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person_favorites
    ADD CONSTRAINT fk_rails_623cbf8cc5 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: casts fk_rails_691c85cc04; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casts
    ADD CONSTRAINT fk_rails_691c85cc04 FOREIGN KEY (person_id) REFERENCES public.people(id);


--
-- Name: activity_groups fk_rails_694252c49b; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activity_groups
    ADD CONSTRAINT fk_rails_694252c49b FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: mute_users fk_rails_6ceac60b15; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mute_users
    ADD CONSTRAINT fk_rails_6ceac60b15 FOREIGN KEY (muted_user_id) REFERENCES public.users(id);


--
-- Name: oauth_access_tokens fk_rails_732cb83ab7; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens
    ADD CONSTRAINT fk_rails_732cb83ab7 FOREIGN KEY (application_id) REFERENCES public.oauth_applications(id);


--
-- Name: episodes fk_rails_734a1c8423; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT fk_rails_734a1c8423 FOREIGN KEY (prev_episode_id) REFERENCES public.episodes(id);


--
-- Name: work_records fk_rails_74a66bd6c5; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records
    ADD CONSTRAINT fk_rails_74a66bd6c5 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: characters fk_rails_75b4872f7c; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters
    ADD CONSTRAINT fk_rails_75b4872f7c FOREIGN KEY (series_id) REFERENCES public.series(id);


--
-- Name: users fk_rails_878aeec0fd; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_rails_878aeec0fd FOREIGN KEY (gumroad_subscriber_id) REFERENCES public.gumroad_subscribers(id);


--
-- Name: forum_post_participants fk_rails_88b2df0cf7; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_post_participants
    ADD CONSTRAINT fk_rails_88b2df0cf7 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: character_favorites fk_rails_8a1652d591; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_favorites
    ADD CONSTRAINT fk_rails_8a1652d591 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: collection_items fk_rails_8f44cb7ace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT fk_rails_8f44cb7ace FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: records fk_rails_9502ee6cab; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT fk_rails_9502ee6cab FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: library_entries fk_rails_963acfc4d0; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_963acfc4d0 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: casts fk_rails_9762f07eaf; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casts
    ADD CONSTRAINT fk_rails_9762f07eaf FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: forum_posts fk_rails_98255e4c28; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_posts
    ADD CONSTRAINT fk_rails_98255e4c28 FOREIGN KEY (forum_category_id) REFERENCES public.forum_categories(id);


--
-- Name: episode_records fk_rails_991f4fd66d; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT fk_rails_991f4fd66d FOREIGN KEY (multiple_episode_record_id) REFERENCES public.multiple_episode_records(id);


--
-- Name: collections fk_rails_9b33697360; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT fk_rails_9b33697360 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: work_taggables fk_rails_9b7494f869; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_taggables
    ADD CONSTRAINT fk_rails_9b7494f869 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: reactions fk_rails_9f02fc96a0; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT fk_rails_9f02fc96a0 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: faq_contents fk_rails_a06684bdd4; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faq_contents
    ADD CONSTRAINT fk_rails_a06684bdd4 FOREIGN KEY (faq_category_id) REFERENCES public.faq_categories(id);


--
-- Name: activities fk_rails_a123739c18; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_a123739c18 FOREIGN KEY (multiple_episode_record_id) REFERENCES public.multiple_episode_records(id);


--
-- Name: activities fk_rails_a20b2f59b3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_a20b2f59b3 FOREIGN KEY (episode_id) REFERENCES public.episodes(id);


--
-- Name: library_entries fk_rails_ab6e2c9467; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_ab6e2c9467 FOREIGN KEY (status_id) REFERENCES public.statuses(id);


--
-- Name: library_entries fk_rails_ac7d3615bf; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_ac7d3615bf FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: userland_project_members fk_rails_afb35090c8; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_project_members
    ADD CONSTRAINT fk_rails_afb35090c8 FOREIGN KEY (userland_project_id) REFERENCES public.userland_projects(id);


--
-- Name: collection_items fk_rails_b1a778644b; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT fk_rails_b1a778644b FOREIGN KEY (collection_id) REFERENCES public.collections(id);


--
-- Name: activities fk_rails_b2055953e1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_b2055953e1 FOREIGN KEY (work_record_id) REFERENCES public.work_records(id);


--
-- Name: work_records fk_rails_b2885fe9d3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records
    ADD CONSTRAINT fk_rails_b2885fe9d3 FOREIGN KEY (record_id) REFERENCES public.records(id);


--
-- Name: oauth_access_grants fk_rails_b4b53e07b8; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants
    ADD CONSTRAINT fk_rails_b4b53e07b8 FOREIGN KEY (application_id) REFERENCES public.oauth_applications(id);


--
-- Name: episode_records fk_rails_b62b8cb1a9; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT fk_rails_b62b8cb1a9 FOREIGN KEY (review_id) REFERENCES public.work_records(id);


--
-- Name: forum_post_participants fk_rails_b725df9f16; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_post_participants
    ADD CONSTRAINT fk_rails_b725df9f16 FOREIGN KEY (forum_post_id) REFERENCES public.forum_posts(id);


--
-- Name: reactions fk_rails_b76961a906; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT fk_rails_b76961a906 FOREIGN KEY (target_user_id) REFERENCES public.users(id);


--
-- Name: work_comments fk_rails_b80e927375; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_comments
    ADD CONSTRAINT fk_rails_b80e927375 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: episode_records fk_rails_bbae928505; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT fk_rails_bbae928505 FOREIGN KEY (oauth_application_id) REFERENCES public.oauth_applications(id);


--
-- Name: work_images fk_rails_bd1806cf80; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_images
    ADD CONSTRAINT fk_rails_bd1806cf80 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: works fk_rails_bdb9fb31c3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.works
    ADD CONSTRAINT fk_rails_bdb9fb31c3 FOREIGN KEY (key_pv_id) REFERENCES public.trailers(id);


--
-- Name: forum_posts fk_rails_c76798dc77; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_posts
    ADD CONSTRAINT fk_rails_c76798dc77 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: userland_projects fk_rails_c8f0a2299f; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.userland_projects
    ADD CONSTRAINT fk_rails_c8f0a2299f FOREIGN KEY (userland_category_id) REFERENCES public.userland_categories(id);


--
-- Name: character_images fk_rails_cc4ab090bf; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_images
    ADD CONSTRAINT fk_rails_cc4ab090bf FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: forum_comments fk_rails_ce6ed0c47a; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_comments
    ADD CONSTRAINT fk_rails_ce6ed0c47a FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: work_records fk_rails_d475d93649; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records
    ADD CONSTRAINT fk_rails_d475d93649 FOREIGN KEY (oauth_application_id) REFERENCES public.oauth_applications(id);


--
-- Name: library_entries fk_rails_d60c2fc1be; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_entries
    ADD CONSTRAINT fk_rails_d60c2fc1be FOREIGN KEY (program_id) REFERENCES public.programs(id);


--
-- Name: organization_favorites fk_rails_d9d1d7e461; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_favorites
    ADD CONSTRAINT fk_rails_d9d1d7e461 FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: multiple_episode_records fk_rails_da1ee2634a; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.multiple_episode_records
    ADD CONSTRAINT fk_rails_da1ee2634a FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: work_records fk_rails_dadac170c9; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_records
    ADD CONSTRAINT fk_rails_dadac170c9 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: forum_comments fk_rails_e0e6d14a1e; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forum_comments
    ADD CONSTRAINT fk_rails_e0e6d14a1e FOREIGN KEY (forum_post_id) REFERENCES public.forum_posts(id);


--
-- Name: vod_titles fk_rails_e5ef5f40f9; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vod_titles
    ADD CONSTRAINT fk_rails_e5ef5f40f9 FOREIGN KEY (work_id) REFERENCES public.works(id);


--
-- Name: slots fk_rails_e99f16a883; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slots
    ADD CONSTRAINT fk_rails_e99f16a883 FOREIGN KEY (program_id) REFERENCES public.programs(id);


--
-- Name: oauth_access_tokens fk_rails_ee63f25419; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens
    ADD CONSTRAINT fk_rails_ee63f25419 FOREIGN KEY (resource_owner_id) REFERENCES public.users(id);


--
-- Name: series_works fk_rails_f4eb19863e; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_works
    ADD CONSTRAINT fk_rails_f4eb19863e FOREIGN KEY (series_id) REFERENCES public.series(id);


--
-- Name: programs fk_rails_f62ce4530d; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs
    ADD CONSTRAINT fk_rails_f62ce4530d FOREIGN KEY (channel_id) REFERENCES public.channels(id);


--
-- Name: activities fk_rails_f6761d2258; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT fk_rails_f6761d2258 FOREIGN KEY (episode_record_id) REFERENCES public.episode_records(id);


--
-- Name: mute_users fk_rails_f7dabf385e; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mute_users
    ADD CONSTRAINT fk_rails_f7dabf385e FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: statuses fk_rails_fb1024dbb6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statuses
    ADD CONSTRAINT fk_rails_fb1024dbb6 FOREIGN KEY (oauth_application_id) REFERENCES public.oauth_applications(id);


--
-- Name: episode_records fk_rails_ff2b5e1c03; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.episode_records
    ADD CONSTRAINT fk_rails_ff2b5e1c03 FOREIGN KEY (record_id) REFERENCES public.records(id);


--
-- Name: follows follows_following_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_following_id_fk FOREIGN KEY (following_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: follows follows_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: likes likes_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: notifications notifications_action_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_action_user_id_fk FOREIGN KEY (action_user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: notifications notifications_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: profiles profiles_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: slots programs_channel_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slots
    ADD CONSTRAINT programs_channel_id_fk FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: slots programs_episode_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slots
    ADD CONSTRAINT programs_episode_id_fk FOREIGN KEY (episode_id) REFERENCES public.episodes(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: slots programs_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slots
    ADD CONSTRAINT programs_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: providers providers_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: receptions receptions_channel_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.receptions
    ADD CONSTRAINT receptions_channel_id_fk FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: receptions receptions_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.receptions
    ADD CONSTRAINT receptions_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: statuses statuses_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statuses
    ADD CONSTRAINT statuses_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: statuses statuses_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statuses
    ADD CONSTRAINT statuses_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: syobocal_alerts syobocal_alerts_work_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.syobocal_alerts
    ADD CONSTRAINT syobocal_alerts_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE;


--
-- Name: works works_season_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.works
    ADD CONSTRAINT works_season_id_fk FOREIGN KEY (season_id) REFERENCES public.seasons(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- PostgreSQL database dump complete
--

SET search_path TO "$user", public;

INSERT INTO "schema_migrations" (version) VALUES
('20131229220727'),
('20131229230311'),
('20131230002041'),
('20131230090923'),
('20131230192401'),
('20131230192409'),
('20140103142823'),
('20140104063049'),
('20140104082751'),
('20140104123007'),
('20140104124924'),
('20140104130617'),
('20140104134553'),
('20140104211010'),
('20140104212322'),
('20140106144311'),
('20140106144902'),
('20140107145209'),
('20140114161232'),
('20140210075251'),
('20140210133144'),
('20140210135818'),
('20140210153830'),
('20140211183734'),
('20140214161838'),
('20140214172403'),
('20140214181913'),
('20140214183914'),
('20140215063411'),
('20140215211915'),
('20140221174325'),
('20140226162109'),
('20140226164038'),
('20140227103156'),
('20140228170557'),
('20140301073350'),
('20140301122812'),
('20140301183642'),
('20140305152616'),
('20140309135649'),
('20140315035008'),
('20140524213219'),
('20140602154510'),
('20140607063809'),
('20140607063810'),
('20140607063820'),
('20140607063828'),
('20140607070304'),
('20140607115153'),
('20140607142258'),
('20140611141958'),
('20140625141826'),
('20140703153137'),
('20140706075907'),
('20140706085430'),
('20140707140832'),
('20140906095802'),
('20140906175518'),
('20140906180904'),
('20140906212132'),
('20140907083437'),
('20140907083945'),
('20140907135553'),
('20140923115945'),
('20140927222132'),
('20141012034243'),
('20141018054223'),
('20141021143847'),
('20141021143930'),
('20141105161907'),
('20141119155031'),
('20150117060033'),
('20150117074637'),
('20150117125806'),
('20150117134132'),
('20150123161810'),
('20150213154048'),
('20150214071302'),
('20150214112613'),
('20150228061604'),
('20150328162305'),
('20150328162306'),
('20150404074101'),
('20150405033916'),
('20150418082724'),
('20150418164640'),
('20150601150738'),
('20150601150750'),
('20150601151515'),
('20150601152739'),
('20150601152819'),
('20150602150804'),
('20150602150805'),
('20150602150806'),
('20150602150807'),
('20150602150808'),
('20150616125506'),
('20150616140258'),
('20150626152128'),
('20150830031923'),
('20150913042201'),
('20150917124829'),
('20150919150654'),
('20150919150813'),
('20150920010436'),
('20151003131229'),
('20151011014019'),
('20151031031819'),
('20151031035357'),
('20151101141404'),
('20151102154624'),
('20151102191220'),
('20151209151528'),
('20151212001204'),
('20151212002901'),
('20151212003147'),
('20151212010835'),
('20151212101013'),
('20151212101014'),
('20151219113921'),
('20151226133209'),
('20151227035952'),
('20151227064731'),
('20151227101344'),
('20160124155103'),
('20160128143628'),
('20160207110855'),
('20160207115414'),
('20160207120355'),
('20160212162040'),
('20160214140638'),
('20160301144255'),
('20160302145703'),
('20160302151415'),
('20160304163745'),
('20160320055926'),
('20160320200438'),
('20160410033114'),
('20160426132549'),
('20160429044406'),
('20160507035528'),
('20160514062037'),
('20160515160618'),
('20160611081758'),
('20160710025846'),
('20160804053023'),
('20160911021552'),
('20160913135207'),
('20160918071042'),
('20161009121326'),
('20161010060659'),
('20161014154225'),
('20161015142423'),
('20161015143314'),
('20161015184145'),
('20161020170609'),
('20161024024431'),
('20161024131922'),
('20161102155615'),
('20161104210958'),
('20161111141135'),
('20161111141136'),
('20161130133758'),
('20161203102005'),
('20170106055713'),
('20170108160017'),
('20170108160156'),
('20170108160231'),
('20170113204733'),
('20170211110911'),
('20170211135244'),
('20170212044410'),
('20170222125706'),
('20170226112004'),
('20170307135652'),
('20170307135653'),
('20170307135654'),
('20170318171853'),
('20170319141322'),
('20170319153219'),
('20170320070746'),
('20170331183853'),
('20170401041109'),
('20170401042901'),
('20170402135324'),
('20170404121800'),
('20170407131410'),
('20170407131534'),
('20170408033245'),
('20170528054105'),
('20170531154955'),
('20170601142332'),
('20170601143635'),
('20170604123700'),
('20170617163408'),
('20170617192644'),
('20170618013307'),
('20170624160323'),
('20170624180724'),
('20170629141843'),
('20170629144318'),
('20170629144332'),
('20170702164228'),
('20170712144231'),
('20170712145119'),
('20170716104237'),
('20170724170348'),
('20170724170350'),
('20170803123211'),
('20170803123223'),
('20170803132636'),
('20170812162147'),
('20170813054530'),
('20170814131925'),
('20170822123657'),
('20170822141842'),
('20170902100856'),
('20170926175406'),
('20171014094051'),
('20171103133758'),
('20171103133759'),
('20171120125737'),
('20171121111423'),
('20171123083841'),
('20171126032600'),
('20171209032152'),
('20171216120948'),
('20171216120949'),
('20171216120950'),
('20171223071225'),
('20171225125419'),
('20180104123629'),
('20180104131620'),
('20180104134415'),
('20180120150327'),
('20180120155138'),
('20180210073441'),
('20180304030717'),
('20180401123249'),
('20180403091308'),
('20180526135117'),
('20180527171424'),
('20180617144946'),
('20180710152207'),
('20180722042758'),
('20180925112149'),
('20190114090325'),
('20190216165345'),
('20190218014736'),
('20190317000000'),
('20190330112541'),
('20190413221017'),
('20190501175021'),
('20190608074205'),
('20191013172849'),
('20191019230403'),
('20191123150532'),
('20191123191135'),
('20191130150830'),
('20191207094223'),
('20191207113735'),
('20191208154530'),
('20200310195638'),
('20200322025837'),
('20200503204607'),
('20200504053317'),
('20200513125708'),
('20200513125709'),
('20200515062450'),
('20200525110837'),
('20200525201620'),
('20200808233237'),
('20200809161251'),
('20201015141021'),
('20201017210101'),
('20210809083311'),
('20210919175411'),
('20211016135715'),
('20211017074902');


