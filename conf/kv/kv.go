package kv

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseBasic(r io.Reader) (map[string]string, error) {
	var (
		s   = bufio.NewScanner(r)
		out = map[string]string{}
		err error
		i   int
	)

	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("line %d: no key in key=value entry", i+1)
		}
		if idx+1 >= len(line) {
			return nil, fmt.Errorf("line %d: no value for key %q", i+1, line[:idx])
		}
		key, val := strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			if val, err = strconv.Unquote(val); err != nil {
				return nil, fmt.Errorf("line %d: bad quotation: %v", i+1, err)
			}
		}

		i++
		out[key] = val
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning: %v", err)
	}
	return out, nil
}
