package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSocketLink(t *testing.T) {
	cases := []struct {
		link string
		want uint32
		ok   bool
	}{
		{"socket:[12345]", 12345, true},
		{"socket:[0]", 0, true},
		{"pipe:[999]", 0, false},
		{"/dev/null", 0, false},
		{"socket:[abc]", 0, false},
		{"socket:[123", 0, false},
	}
	for _, c := range cases {
		got, ok := parseSocketLink(c.link)
		if got != c.want || ok != c.ok {
			t.Errorf("parseSocketLink(%q) = %d, %v; want %d, %v", c.link, got, ok, c.want, c.ok)
		}
	}
}

func TestScanSocketOwners(t *testing.T) {
	root := t.TempDir()
	link := func(pid, fd, target string) {
		dir := filepath.Join(root, pid, "fd")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(dir, fd)); err != nil {
			t.Fatal(err)
		}
	}
	link("100", "3", "socket:[555]")
	link("100", "4", "/dev/null")
	link("200", "3", "socket:[555]") // shared with 100: first PID wins
	link("200", "4", "socket:[777]")
	if err := os.MkdirAll(filepath.Join(root, "self"), 0o755); err != nil { // not a PID
		t.Fatal(err)
	}

	got := scanSocketOwners(root)
	if len(got) != 2 || got[555] != 100 || got[777] != 200 {
		t.Fatalf("owners = %v, want map[555:100 777:200]", got)
	}
}
