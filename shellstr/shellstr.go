package shellstr

import (
	"strings"
)

// Op describes a chainable operation that can be applied to a string.
type Op interface {
	Apply(string) string
}

// Head takes the first N lines of the output.
type Head struct {
	N int
}

// Apply implements Op.
func (o *Head) Apply(in string) string {
	var (
		out      strings.Builder
		consumed int
	)
	for i := 0; i < o.N; i++ {
		// fmt.Printf("consumed=%d, out=%q, pending=%q, i=%d\n", consumed, out.String(), in[consumed:], i)
		idx := strings.Index(in[consumed:], "\n")
		if idx == -1 {
			return in
		}
		out.WriteString(in[consumed : idx+consumed+1])
		consumed += idx + 1
	}
	return out.String()
}

// Tail takes the last N lines of the output.
type Tail struct {
	N int
}

// Apply implements Op.
func (o *Tail) Apply(in string) string {
	index := make([]int, 0, 16)
	for i := 0; i < len(in); i++ {
		// fmt.Printf("consumed=%d, out=%q, pending=%q, i=%d\n", consumed, out.String(), in[consumed:], i)
		idx := strings.Index(in[i:], "\n")
		if idx == -1 {
			break
		}
		index = append(index, i+idx)
		i += idx
	}

	if o.N >= len(index) {
		return in
	}

	return in[index[len(index)-o.N-1]+1:]
}

// Cut cuts out a field from each line of the input.
type Cut struct {
	Delim string
	To    int
	From  int
}

// Apply implements Op.
func (o *Cut) Apply(in string) string {
	var (
		out   strings.Builder
		lines = strings.Split(in, "\n")
	)
	for i, l := range lines {
		spl := strings.Split(l, o.Delim)
		from, to := 0, len(spl)
		if o.To < len(spl) && o.To > 0 {
			to = o.To
		}
		if o.From-1 < len(spl) && o.From > 0 {
			from = o.From - 1
		}
		out.WriteString(strings.Join(spl[from:to], o.Delim))
		if i < len(lines)-1 {
			out.WriteRune('\n')
		}
	}
	return out.String()
}

// Trim trims the given string by sequentially applying the given operations.
func Trim(s string, ops ...Op) string {
	for _, op := range ops {
		s = op.Apply(s)
	}
	return s
}
