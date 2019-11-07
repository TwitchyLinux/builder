package shellstr

import "testing"

func TestHead(t *testing.T) {
	tcs := []struct {
		name    string
		head    Head
		in, out string
	}{
		{
			name: "trim 2/3",
			head: Head{N: 2},
			in:   "line1\nline2\nline3\n",
			out:  "line1\nline2\n",
		},
		{
			name: "trim 1/3",
			head: Head{N: 1},
			in:   "line1\nline2\nline3\n",
			out:  "line1\n",
		},
		{
			name: "match 3/3",
			head: Head{N: 3},
			in:   "line1\nline2\nline3\n",
			out:  "line1\nline2\nline3\n",
		},
		{
			name: "more 5/3",
			head: Head{N: 5},
			in:   "line1\nline2\nline3\n",
			out:  "line1\nline2\nline3\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if out := tc.head.Apply(tc.in); tc.out != out {
				t.Errorf("output = %q, want %q", out, tc.out)
			}
		})
	}
}

func TestTail(t *testing.T) {
	tcs := []struct {
		name    string
		head    Tail
		in, out string
	}{
		{
			name: "last 2/3",
			head: Tail{N: 2},
			in:   "line1\nline2\nline3\n",
			out:  "line2\nline3\n",
		},
		{
			name: "last 1/3",
			head: Tail{N: 1},
			in:   "line1\nline2\nline3\n",
			out:  "line3\n",
		},
		{
			name: "match 3/3",
			head: Tail{N: 3},
			in:   "line1\nline2\nline3\n",
			out:  "line1\nline2\nline3\n",
		},
		{
			name: "more 5/3",
			head: Tail{N: 5},
			in:   "line1\nline2\nline3\n",
			out:  "line1\nline2\nline3\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if out := tc.head.Apply(tc.in); tc.out != out {
				t.Errorf("output = %q, want %q", out, tc.out)
			}
		})
	}
}

func TestCut(t *testing.T) {
	tcs := []struct {
		name    string
		head    Cut
		in, out string
	}{
		{
			name: "space 1-2",
			head: Cut{Delim: " ", From: 1, To: 2},
			in:   "blue green red yellow\n",
			out:  "blue green\n",
		},
		{
			name: "space 1-3",
			head: Cut{Delim: " ", From: 1, To: 3},
			in:   "blue green red yellow\n",
			out:  "blue green red\n",
		},
		{
			name: "space 2-2",
			head: Cut{Delim: " ", From: 2, To: 2},
			in:   "blue green red yellow\n",
			out:  "green\n",
		},
		{
			name: "space 2-4",
			head: Cut{Delim: " ", From: 2, To: 4},
			in:   "blue green red yellow\n",
			out:  "green red yellow\n",
		},
		{
			name: "space 2-3",
			head: Cut{Delim: " ", From: 2, To: 3},
			in:   "blue green red yellow\n",
			out:  "green red\n",
		},
		{
			name: "space 1-30",
			head: Cut{Delim: " ", From: 1, To: 30},
			in:   "blue green red yellow\n",
			out:  "blue green red yellow\n",
		},
		{
			name: "space 2-",
			head: Cut{Delim: " ", From: 2},
			in:   "blue green red yellow\n",
			out:  "green red yellow\n",
		},
		{
			name: "space -3",
			head: Cut{Delim: " ", To: 3},
			in:   "blue green red yellow\n",
			out:  "blue green red\n",
		},
		{
			name: "comma 2-3",
			head: Cut{Delim: ",", From: 2, To: 3},
			in:   "blue,green, red,yellow\n",
			out:  "green, red\n",
		},
		{
			name: "multiline 2-3",
			head: Cut{Delim: " ", From: 2, To: 3},
			in:   "blue green red yellow\n\ngreen blue red",
			out:  "green red\n\nblue red",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if out := tc.head.Apply(tc.in); tc.out != out {
				t.Errorf("output = %q, want %q", out, tc.out)
			}
		})
	}
}
