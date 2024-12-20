package filter

import (
	"testing"
)

func TestMatchesAny(t *testing.T) {
	cases := []struct {
		path          string
		patterns      []string
		caseSensitive bool
		want          bool
	}{
		{
			path:          "my-folder/subdir/deeper/file.go",
			patterns:      []string{"my-folder/**/*.go"},
			caseSensitive: false,
			want:          true,
		},
		{
			path:          "my-folder/subdir/file.txt",
			patterns:      []string{"my-folder/**/*.go"},
			caseSensitive: false,
			want:          false,
		},
		{
			path:          "my-folder/file.go",
			patterns:      []string{"my-folder/**/*.go"},
			caseSensitive: false,
			want:          true,
		},
		{
			path:          "another-folder/file.go",
			patterns:      []string{"my-folder/**/*.go"},
			caseSensitive: false,
			want:          false,
		},
	}

	for _, tc := range cases {
		got := MatchesAny(tc.path, tc.patterns, tc.caseSensitive)
		if got != tc.want {
			t.Errorf("MatchesAny(%q, %v) = %v; want %v", tc.path, tc.patterns, got, tc.want)
		}
	}
}
