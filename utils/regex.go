package utils

import "regexp"

func MapRegexMatches(r *regexp.Regexp, v string) (m map[string]string) {
	if r == nil {
		return nil
	}

	match := r.FindStringSubmatch(v)

	if match == nil {
		return nil
	}

	m = make(map[string]string)
	m[""] = match[0]

	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			m[name] = match[i]
		}
	}

	return
}
