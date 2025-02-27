package rulejson

import (
	"fmt"
	"strconv"
	"strings"
)

type Target map[string]string

type TargetValue struct {
	Value string `json:"value"`
}

type TargetRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func sqlCompileTarget(operator string, a any, attr RuleAttribute) string {
	switch operator {
	case "equal":
		return sqlCompileTargetEqual(a, attr)
	case "range":
		return sqlCompileTargetRange(a, attr)
	case "isSubstringOf":
		return sqlCompileTargetIsSubstringOf(a, attr)
	case "matchesWildcard":
		return sqlCompileTargetMatchesWildcard(a, attr)
	default:
		return "N/A"
	}
}

func evaluateTarget(operator string, a any, input string, kind string) bool {
	switch operator {
	case "equal":
		return targetEqual(a, input, kind)
	case "range":
		return targetRange(a, input, kind)
	case "isSubstringOf":
		return targetIsSubstringOf(a, input, kind)
	case "matchesWildcard":
		return targetMatchesWildcard(a, input, kind)
	default:
		return false
	}
}

func targetEqual(a any, value string, kind string) bool {
	target := a.(*TargetValue)
	return target.Value == value
}

func targetRange(a any, value string, kind string) bool {
	//nolint:forcetypeassert
	target := a.(*TargetRange)

	v1, _ := strconv.ParseFloat(target.From, 64)
	v2, _ := strconv.ParseFloat(target.To, 64)
	v, _ := strconv.ParseFloat(value, 64)

	return v >= v1 && v <= v2
}

func targetIsSubstringOf(a any, value string, kind string) bool {
	//nolint:forcetypeassert
	target := a.(*TargetValue)

	return strings.Contains(target.Value, value)
}

func targetMatchesWildcard(a any, value string, kind string) bool {
	//nolint:forcetypeassert
	target := a.(*TargetValue)

	return strings.Contains(target.Value, value)
}

func sqlCompileTargetEqual(a any, atrr RuleAttribute) string {
	//nolint:forcetypeassert
	target := a.(*TargetValue)
	if atrr.Kind == "string" {
		return fmt.Sprintf(`%s = "%v"`, atrr.Name, target.Value)
	} else {
		return fmt.Sprintf(`%s = %v`, atrr.Name, target.Value)
	}
}

func sqlCompileTargetRange(a any, atrr RuleAttribute) string {
	//nolint:forcetypeassert
	target := a.(*TargetRange)

	return fmt.Sprintf(`%s BETWEEN %v and %v`, atrr.Name, target.From, target.To)
}

func sqlCompileTargetIsSubstringOf(a any, atrr RuleAttribute) string {
	//nolint:forcetypeassert
	target := a.(*TargetValue)

	return fmt.Sprintf(`%s LIKE %%%v%%`, atrr.Name, target.Value)
}

func sqlCompileTargetMatchesWildcard(a any, atrr RuleAttribute) string {
	//nolint:forcetypeassert
	target := a.(*TargetValue)

	return fmt.Sprintf(`%s LIKE '%v'`, atrr.Name, target.Value)
}
