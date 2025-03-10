package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/geteduroam/linux-app/internal/utils"
)

func mockDir(t *testing.T, dir string) {
	// mock XDG_DATA_HOME
	err := os.Setenv("XDG_DATA_HOME", dir)
	if err != nil {
		t.Fatalf("failed setting environment for XDG_DATA_HOME: %v", err)
	}
}

func TestWrite(t *testing.T) {
	// create a test dir
	// this will be cleaned up when the test finishes, neat!
	dir := t.TempDir()
	mockDir(t, dir)
	c := &Config{
		UUIDs: []string{"test"},
	}
	err := c.Write()
	if err != nil {
		t.Fatalf("error occurred when writing config: %v", err)
	}

	cdir, err := Directory()
	if err != nil {
		t.Fatalf("error occurred when getting config directory: %v", dir)
	}
	r, err := os.ReadFile(filepath.Join(cdir, configName))
	if err != nil {
		t.Fatalf("failed when reading config file: %v", err)
	}
	got := string(r)
	want := `{"v2":{"uuids":["test"]}}`
	if got != want {
		t.Fatalf("config not as expected, got: %v, want: %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	mockDir(t, "test_data")
	cases := []struct {
		filename string
		wantc    *Config
		wanterr  string
	}{
		{
			filename: "invalid.json",
			wantc:    nil,
			wanterr:  "json: cannot unmarshal string into Go value of type config.Versioned",
		},
		{
			filename: "old.json",
			wantc: &Config{
				UUIDs: []string{"test"},
			},
			wanterr: "",
		},
		{
			filename: "valid.json",
			wantc: &Config{
				UUIDs: []string{"test"},
			},
			wanterr: "",
		},
	}

	for _, curr := range cases {
		// mock the config name
		configName = curr.filename
		gotc, goterr := Load()
		if utils.ErrorString(goterr) != curr.wanterr {
			t.Fatalf("expected config error not equal to want, got: %v, want: %v", goterr, curr.wanterr)
		}
		// to compare structs we can use deepequal
		if !reflect.DeepEqual(gotc, curr.wantc) {
			t.Fatalf("expected config not equal to want, got: %v, want: %v", gotc, curr.wantc)
		}
	}
}
