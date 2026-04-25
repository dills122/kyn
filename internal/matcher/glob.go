package matcher

import "github.com/bmatcuk/doublestar/v4"

func MatchAny(patterns []string, p string) (bool, error) {
	for _, pattern := range patterns {
		ok, err := doublestar.PathMatch(pattern, p)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
