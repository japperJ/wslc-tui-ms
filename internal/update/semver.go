package update

import (
	"fmt"
	"strconv"
	"strings"
)

type version struct {
	major, minor, patch int
	pre                 []string
}

func parseVersion(raw string) (version, error) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "v")
	main := s
	var pre string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		main, pre = s[:i], s[i+1:]
	}
	parts := strings.Split(main, ".")
	if len(parts) != 3 {
		return version{}, fmt.Errorf("invalid SemVer %q", raw)
	}
	values := [3]int{}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return version{}, fmt.Errorf("invalid SemVer %q", raw)
		}
		values[i] = n
	}
	v := version{major: values[0], minor: values[1], patch: values[2]}
	if pre != "" {
		v.pre = strings.Split(pre, ".")
	}
	return v, nil
}

func compare(a, b version) int {
	for _, pair := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] != pair[1] {
			if pair[0] < pair[1] {
				return -1
			}
			return 1
		}
	}
	if len(a.pre) == 0 && len(b.pre) > 0 {
		return 1
	}
	if len(a.pre) > 0 && len(b.pre) == 0 {
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		if a.pre[i] != b.pre[i] {
			return strings.Compare(a.pre[i], b.pre[i])
		}
	}
	return len(a.pre) - len(b.pre)
}

func newer(candidate, current string) bool {
	a, e1 := parseVersion(candidate)
	b, e2 := parseVersion(current)
	return e1 == nil && e2 == nil && compare(a, b) > 0
}
