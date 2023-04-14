package config

import (
	"os"
	"path"
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
		UUID: "test",
	}
	err := c.Write()
	if err != nil {
		t.Fatalf("error occurred when writing config: %v", err)
	}

	r, err := os.ReadFile(path.Join(Directory(), configName))
	if err != nil {
		t.Fatalf("failed when reading config file: %v", err)
	}
	got := string(r)
	want := `{"v1":{"uuid":"test"}}`
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
			filename: "valid.json",
			wantc: &Config{
				UUID: "test",
			},
			wanterr: "",
		},
	}

	for _, curr := range cases {
		// mock the config name
		configName = curr.filename
		gotc, goterr := Load()
		// to compare structs we can use deepequal
		if !reflect.DeepEqual(gotc, curr.wantc) {
			t.Fatalf("expected config not equal to want, got: %v, want: %v", gotc, curr.wantc)
		}
		if utils.ErrorString(goterr) != curr.wanterr {
			t.Fatalf("expected config error not equal to want, got: %v, want: %v", goterr, curr.wanterr)
		}
	}
}
