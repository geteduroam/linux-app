package method

import "testing"

func Test_IsValid(t *testing.T) {
	cases := []struct {
		input int
		want  bool
	}{
		{
			input: 0,
			want:  false,
		},
		{
			input: 67,
			want:  false,
		},
		{
			input: 13,
			want:  true,
		},
		{
			input: 21,
			want:  true,
		},
		{
			input: 25,
			want:  true,
		},
		{
			input: -13,
			want:  false,
		},
	}

	for _, c := range cases {
		got := IsValid(c.input)
		if got != c.want {
			t.Fatalf("Got: %v, Want: %v", got, c.want)
		}
	}
}
