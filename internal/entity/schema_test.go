package entity

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  string
		want    Config
		wantErr string
	}{
		{
			name:   "enum only",
			config: "version: \"1\"\ntables:\n  - name: sexes\n    key: en_name\n    value: id\n",
			want: Config{
				Version: "1",
				Tables:  []SchemaTable{{Name: "sexes", Key: "en_name", Value: "id"}},
			},
		},
		{
			name:   "enum and id_type",
			config: "version: \"1\"\ntables:\n  - name: clinics\n    key: en_name\n    value: id\n    id_type: ClinicID\n",
			want: Config{
				Version: "1",
				Tables:  []SchemaTable{{Name: "clinics", Key: "en_name", Value: "id", IDType: "ClinicID"}},
			},
		},
		{
			name:   "id_type only",
			config: "version: \"1\"\ntables:\n  - name: users\n    id_type: UserID\n",
			want: Config{
				Version: "1",
				Tables:  []SchemaTable{{Name: "users", IDType: "UserID"}},
			},
		},
		{
			name:    "key without value",
			config:  "version: \"1\"\ntables:\n  - name: clinics\n    key: en_name\n    id_type: ClinicID\n",
			wantErr: "table clinics: key and value must be specified together",
		},
		{
			name:    "value without key",
			config:  "version: \"1\"\ntables:\n  - name: clinics\n    value: id\n",
			wantErr: "table clinics: key and value must be specified together",
		},
		{
			name:    "neither key/value nor id_type",
			config:  "version: \"1\"\ntables:\n  - name: clinics\n",
			wantErr: "table clinics: key and value or id_type is required",
		},
		{
			name:    "unknown key",
			config:  "version: \"1\"\ntables:\n  - name: clinics\n    key: en_name\n    value: id\n    unknown: x\n",
			wantErr: "field unknown not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "sqlenumgen.yml")
			if err := os.WriteFile(path, []byte(tt.config), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := NewConfigFromFile(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NewConfigFromFile() error = %v, want %q", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("NewConfigFromFile() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigFromFile() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfig_EnumTables(t *testing.T) {
	t.Parallel()

	config := Config{
		Tables: []SchemaTable{
			{Name: "clinics", Key: "en_name", Value: "id", IDType: "ClinicID"},
			{Name: "users", IDType: "UserID"},
			{Name: "sexes", Key: "en_name", Value: "id"},
		},
	}

	want := []SchemaTable{
		{Name: "clinics", Key: "en_name", Value: "id", IDType: "ClinicID"},
		{Name: "sexes", Key: "en_name", Value: "id"},
	}

	if got := config.EnumTables(); !reflect.DeepEqual(got, want) {
		t.Errorf("EnumTables() = %+v, want %+v", got, want)
	}

	if !config.HasIDType() {
		t.Error("HasIDType() = false, want true")
	}

	if (Config{Tables: want[1:]}).HasIDType() {
		t.Error("HasIDType() = true, want false")
	}
}
