package updater

import (
	"fmt"
	"strconv"
	"strings"
)

// Compares numeric versions such as 1.2.3 and 1.10.0.
// This intentionally does not implement prerelease semver ordering.
func isNewerVersion(latest, current string) (bool, error) {
	parse := func(s string) ([]uint64, error) {
		s = strings.TrimPrefix(s, "v")
		parts := strings.Split(s, ".")
		if len(parts) == 0 {
			return nil, fmt.Errorf("invalid version %q", s)
		}

		values := make([]uint64, len(parts))
		for i, part := range parts {
			if part == "" {
				return nil, fmt.Errorf("invalid version %q", s)
			}
			v, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid version %q", s)
			}
			values[i] = v
		}
		return values, nil
	}

	a, err := parse(latest)
	if err != nil {
		return false, err
	}
	b, err := parse(current)
	if err != nil {
		return false, err
	}

	n := len(a)
	if len(b) > n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		var av, bv uint64
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av != bv {
			return av > bv, nil
		}
	}

	return false, nil
}
