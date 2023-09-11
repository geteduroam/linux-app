package utils

import "testing"

func TestRemoveDiacritics(t *testing.T) {
	cases := []struct {
		input string
		want  string
		e     error
	}{
		{
			input: "foobar",
			want:  "foobar",
			e:     nil,
		},
		{
			input: "fòóbår",
			want:  "foobar",
			e:     nil,
		},
	}

	for _, c := range cases {
		result, e := RemoveDiacritics(c.input)
		if result != c.want || e != c.e {
			t.Fatalf("Result: %s, %v Want: %s, %v", result, e, c.want, c.e)
		}
	}
}
