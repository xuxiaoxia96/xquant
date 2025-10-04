package env

import (
	"strings"
	"xquant/pkg/utils"
)

var ConfEnv = utils.GetEnvWithDefault("XQUANT_ENV", "local")

type StringCommaSepList string

func (sc StringCommaSepList) Contains(s string) bool {
	return sc.contains(s) || sc.contains("[all]")
}

func (sc StringCommaSepList) contains(s string) bool {
	if sc == "" {
		return false
	}

	for _, ss := range strings.Split(string(sc), ",") {
		if s == strings.TrimSpace(ss) {
			return true
		}
	}

	return false
}

type StringCommaSepListIgnoreCase string

func (sc StringCommaSepListIgnoreCase) Contains(s string) bool {
	return sc.contains(s) || sc.contains("[all]")
}

func (sc StringCommaSepListIgnoreCase) contains(s string) bool {
	if sc == "" {
		return false
	}

	for _, ss := range strings.Split(string(sc), ",") {
		if strings.EqualFold(s, strings.TrimSpace(ss)) {
			return true
		}
	}

	return false
}
