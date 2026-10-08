package parser

import (
	"strings"
	"testing"

	"github.com/flum1025/sql-enum-generator/internal/entity"
)

func TestPostgresParser_ParseDefinitions_ToID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		def     entity.SchemaTable
		want    ID
		wantErr string
	}{
		{
			name: "uuid",
			source: `CREATE TABLE "public"."clinics" (
  "id" UUID NOT NULL DEFAULT gen_random_uuid(),
  "name" CHARACTER VARYING (32) NOT NULL,
  PRIMARY KEY ("id")
);`,
			def:  entity.SchemaTable{Name: "clinics", IDType: "ClinicID"},
			want: ID{Name: "ClinicID", Table: "clinics", Type: IDTypeUUID},
		},
		{
			name: "serial",
			source: `CREATE TABLE "public"."richmenu_types" (
  "id" SERIAL NOT NULL,
  "name" CHARACTER VARYING (32) NOT NULL,
  PRIMARY KEY ("id")
);`,
			def:  entity.SchemaTable{Name: "richmenu_types", IDType: "RichMenuTypeID"},
			want: ID{Name: "RichMenuTypeID", Table: "richmenu_types", Type: IDTypeInt32},
		},
		{
			name:   "integer",
			source: `CREATE TABLE products (id integer NOT NULL, PRIMARY KEY (id));`,
			def:    entity.SchemaTable{Name: "products", IDType: "ProductID"},
			want:   ID{Name: "ProductID", Table: "products", Type: IDTypeInt32},
		},
		{
			name:   "bigint",
			source: `CREATE TABLE events (id BIGINT NOT NULL, PRIMARY KEY (id));`,
			def:    entity.SchemaTable{Name: "events", IDType: "EventID"},
			want:   ID{Name: "EventID", Table: "events", Type: IDTypeInt64},
		},
		{
			name:   "bigserial",
			source: `CREATE TABLE events (id BIGSERIAL NOT NULL, PRIMARY KEY (id));`,
			def:    entity.SchemaTable{Name: "events", IDType: "EventID"},
			want:   ID{Name: "EventID", Table: "events", Type: IDTypeInt64},
		},
		{
			name:   "unquoted and not schema qualified",
			source: `CREATE TABLE clinics (id uuid NOT NULL, PRIMARY KEY (id));`,
			def:    entity.SchemaTable{Name: "clinics", IDType: "ClinicID"},
			want:   ID{Name: "ClinicID", Table: "clinics", Type: IDTypeUUID},
		},
		{
			name:   "schema qualified without quotes",
			source: `CREATE TABLE public.clinics (id uuid NOT NULL, PRIMARY KEY (id));`,
			def:    entity.SchemaTable{Name: "clinics", IDType: "ClinicID"},
			want:   ID{Name: "ClinicID", Table: "clinics", Type: IDTypeUUID},
		},
		{
			name:   "column level primary key",
			source: `CREATE TABLE "public"."users" ("id" UUID PRIMARY KEY DEFAULT gen_random_uuid(), "name" TEXT);`,
			def:    entity.SchemaTable{Name: "users", IDType: "UserID"},
			want:   ID{Name: "UserID", Table: "users", Type: IDTypeUUID},
		},
		{
			name: "select table from multiple statements",
			source: `CREATE EXTENSION ltree;
CREATE TABLE "public"."users" ("id" UUID NOT NULL, PRIMARY KEY ("id"));
CREATE TABLE "public"."menus" ("id" SERIAL NOT NULL, PRIMARY KEY ("id"));
ALTER TABLE "public"."menus" ADD CONSTRAINT "menus_id_key" UNIQUE (id);`,
			def:  entity.SchemaTable{Name: "menus", IDType: "MenuID"},
			want: ID{Name: "MenuID", Table: "menus", Type: IDTypeInt32},
		},
		{
			name: "primary key added by alter table only",
			source: `CREATE TABLE public.clinics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    en_name character varying(50) NOT NULL
);

ALTER TABLE ONLY public.clinics
    ADD CONSTRAINT clinics_pkey PRIMARY KEY (id);`,
			def:  entity.SchemaTable{Name: "clinics", IDType: "ClinicID"},
			want: ID{Name: "ClinicID", Table: "clinics", Type: IDTypeUUID},
		},
		{
			name: "primary key added by alter table before create table",
			source: `ALTER TABLE "public"."clinics" ADD PRIMARY KEY ("id");
CREATE TABLE "public"."clinics" ("id" UUID NOT NULL);`,
			def:  entity.SchemaTable{Name: "clinics", IDType: "ClinicID"},
			want: ID{Name: "ClinicID", Table: "clinics", Type: IDTypeUUID},
		},
		{
			name: "serial as integer with sequence default",
			source: `CREATE TABLE public.richmenu_types (
    id integer NOT NULL,
    ja_name character varying(32) NOT NULL
);

CREATE SEQUENCE public.richmenu_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.richmenu_types_id_seq OWNED BY public.richmenu_types.id;

ALTER TABLE ONLY public.richmenu_types ALTER COLUMN id SET DEFAULT nextval('public.richmenu_types_id_seq'::regclass);

ALTER TABLE ONLY public.richmenu_types
    ADD CONSTRAINT richmenu_types_pkey PRIMARY KEY (id);`,
			def:  entity.SchemaTable{Name: "richmenu_types", IDType: "RichMenuTypeID"},
			want: ID{Name: "RichMenuTypeID", Table: "richmenu_types", Type: IDTypeInt32},
		},
		{
			name: "psql meta commands and other statements",
			source: `\restrict abcdef
SET statement_timeout = 0;
SELECT pg_catalog.set_config('search_path', '', false);
CREATE FUNCTION public.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;
CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    updated_at timestamp without time zone NOT NULL
);
COMMENT ON TABLE public.users IS 'users';
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_updated_at_key UNIQUE (updated_at);
CREATE INDEX users_updated_at_idx ON public.users USING btree (updated_at);
CREATE TRIGGER users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();
ALTER TABLE ONLY public.unknown_table
    ADD CONSTRAINT unknown_table_pkey PRIMARY KEY (id);
\unrestrict abcdef`,
			def:  entity.SchemaTable{Name: "users", IDType: "UserID"},
			want: ID{Name: "UserID", Table: "users", Type: IDTypeUUID},
		},
		{
			name:    "alter table targets unknown configured table",
			source:  `ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);`,
			def:     entity.SchemaTable{Name: "users", IDType: "UserID"},
			wantErr: "table not found: users",
		},
		{
			name: "composite primary key added by alter table",
			source: `CREATE TABLE public.user_clinics (user_id uuid NOT NULL, clinic_id uuid NOT NULL);
ALTER TABLE ONLY public.user_clinics
    ADD CONSTRAINT user_clinics_pkey PRIMARY KEY (user_id, clinic_id);`,
			def:     entity.SchemaTable{Name: "user_clinics", IDType: "UserClinicID"},
			wantErr: "table user_clinics must have exactly one primary key column, got 2",
		},
		{
			name: "primary key column not found",
			source: `CREATE TABLE public.users (uuid uuid NOT NULL);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);`,
			def:     entity.SchemaTable{Name: "users", IDType: "UserID"},
			wantErr: "table users: primary key column not found: id",
		},
		{
			name:    "composite primary key",
			source:  `CREATE TABLE "public"."user_clinics" ("user_id" UUID NOT NULL, "clinic_id" UUID NOT NULL, PRIMARY KEY ("user_id", "clinic_id"));`,
			def:     entity.SchemaTable{Name: "user_clinics", IDType: "UserClinicID"},
			wantErr: "table user_clinics must have exactly one primary key column, got 2",
		},
		{
			name:    "no primary key",
			source:  `CREATE TABLE logs (message TEXT);`,
			def:     entity.SchemaTable{Name: "logs", IDType: "LogID"},
			wantErr: "table logs must have exactly one primary key column, got 0",
		},
		{
			name:    "missing table",
			source:  `CREATE TABLE clinics (id uuid NOT NULL, PRIMARY KEY (id));`,
			def:     entity.SchemaTable{Name: "users", IDType: "UserID"},
			wantErr: "table not found: users",
		},
		{
			name:    "unsupported type",
			source:  `CREATE TABLE codes (code CHARACTER VARYING (8) NOT NULL, PRIMARY KEY (code));`,
			def:     entity.SchemaTable{Name: "codes", IDType: "CodeID"},
			wantErr: "table codes column code: unsupported primary key type: varchar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			definitions, err := (&PostgresParser{}).ParseDefinitions(tt.source)
			if err != nil {
				t.Fatalf("ParseDefinitions() error = %v", err)
			}

			got, err := definitions.ToID(tt.def)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ToID() error = %v, want %q", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("ToID() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("ToID() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPostgresParser_ParseDefinitions_InvalidSQL(t *testing.T) {
	t.Parallel()

	if _, err := (&PostgresParser{}).ParseDefinitions(`CREATE TABLE (`); err == nil {
		t.Fatal("ParseDefinitions() error = nil, want error")
	}
}
