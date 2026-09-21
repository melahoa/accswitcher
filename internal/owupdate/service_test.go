package owupdate

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"0.5.0", "0.4.0", true},
		{"0.4.0", "0.5.0", false},
		{"0.4.0", "0.4.0", false},
		{"1.0.0", "0.9.9", true},
		{"0.4.10", "0.4.9", true},
		{"0.4.9", "0.4.10", false},
	}
	for _, c := range cases {
		if got := isNewer(c.latest, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestParseVersionToleratesGarbage(t *testing.T) {
	got := parseVersion("not-a-version")
	want := [3]int{0, 0, 0}
	if got != want {
		t.Errorf("parseVersion(garbage) = %v, want %v", got, want)
	}
}

func TestParseVersionHandlesTwoSegments(t *testing.T) {
	got := parseVersion("0.5")
	want := [3]int{0, 5, 0}
	if got != want {
		t.Errorf("parseVersion(\"0.5\") = %v, want %v", got, want)
	}
}
