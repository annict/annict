-- migrate:up

-- syobocal_alertsはRails版からもGo版からも読み書きされていないため削除する。
-- 行は戻せない。
DROP TABLE IF EXISTS public.syobocal_alerts;

-- migrate:down

-- 同じ列・インデックス・外部キーを持つ空のテーブルを作り直す。
CREATE TABLE public.syobocal_alerts (
    id bigserial PRIMARY KEY,
    work_id bigint,
    kind integer NOT NULL,
    sc_prog_item_id integer,
    sc_sub_title character varying(255),
    sc_prog_comment character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    CONSTRAINT syobocal_alerts_work_id_fk FOREIGN KEY (work_id) REFERENCES public.works(id) ON DELETE CASCADE
);
CREATE INDEX index_syobocal_alerts_on_kind ON public.syobocal_alerts USING btree (kind);
CREATE INDEX index_syobocal_alerts_on_sc_prog_item_id ON public.syobocal_alerts USING btree (sc_prog_item_id);
