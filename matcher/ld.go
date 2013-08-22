package matcher

import (
	"github.com/srinathgs/wurflgo/levenshtein"
	)

type LDMatcher struct{

}

func (ld *LDMatcher) Match(collection []string, needle string, tolerance int) string{
	best := tolerance
	match := ""
	needleLength := len(needle)
	for _, ua := range collection {
		uaLen := len(ua)
		var diff int
		if uaLen > needleLength {
			diff = uaLen - needleLength
		} else {
			diff = needleLength - uaLen
		}
		var current int
		if diff <= tolerance {
			current = levenshtein.LD(needle, ua)
			if current <= best {
				best = current - 1
				match = ua
			}
		}
	}
	return match
}