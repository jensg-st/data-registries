package rulejson

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type RuleAttribute struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type Rule struct {
	Name string `json:"name"`
	// possible values "attribute" or "group".
	Type string `json:"type"`
	// with type=group, possible values "AND" or "OR" only,
	// with type=attribute this field is the target type.
	Operator string `json:"operator"`
	// only relevant with type=group
	Items []Rule `json:"items"`
	// only relevant with type=attribute
	Attribute RuleAttribute `json:"attribute"`
	// only relevant with type=attribute
	Attributes []RuleAttribute `json:"attributes"`

	Assert       json.RawMessage `json:"assert"`
	ParsedTarget any             `json:"-"`
	BoolValue    string          `json:"-"`
}

type RuleError struct {
	Name string `json:"name"`
	Err  string `json:"error"`
}

func Validate(rule *Rule) []RuleError {
	var errs []RuleError
	validate(rule, &errs)
	return errs
}

func validate(rule *Rule, errs *[]RuleError) {
	if rule.Name == "" {
		rule.Name = "MissingName"
	}
	if rule.Type == "" {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "empty or missing type",
		})
	}
	if !slices.Contains([]string{"group", "attribute", "bool", "comparison", ""}, rule.Type) {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "invalid rule type, must be `group`, `bool` or `attribute`",
		})
	}
	if rule.Type == "group" && len(rule.Items) == 0 {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `group` rule must have at least one item",
		})
	}
	if rule.Type == "attribute" && len(rule.Items) > 0 {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `attribute` must have no child item",
		})
	}
	if rule.Type == "group" && !slices.Contains([]string{"AND", "OR"}, rule.Operator) {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `group` must have `AND` or `OR` operator",
		})
	}
	if rule.Attribute.Name == "" && rule.Type == "attribute" {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `attribute` must have field `attribute` set",
		})
	}
	if rule.Attribute.Name != "" && rule.Type == "group" {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `group` shouldn't have field `attribute`",
		})
	}
	if rule.Assert == nil && rule.Type == "attribute" {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `attribute` must have field `target` set",
		})
	}
	if rule.Assert != nil && rule.Type == "group" {
		*errs = append(*errs, RuleError{
			Name: rule.Name,
			Err:  "rule with type `group` shouldn't have field `target` set",
		})
	}
	if rule.Type == "attribute" {
		var err error
		switch rule.Operator {
		case "range":
			val := &TargetRange{}
			err = json.Unmarshal(rule.Assert, val)
			rule.ParsedTarget = val
		case "equal", "isSubstringOf", "matchesWildcard":
			val := &TargetValue{}
			err = json.Unmarshal(rule.Assert, val)
			rule.ParsedTarget = val
		}
		if err != nil {
			*errs = append(*errs, RuleError{
				Name: rule.Name,
				Err:  "could not decode rule target",
			})
		}
	}
	if len(rule.Items) > 0 {
		for i := range rule.Items {
			validate(&rule.Items[i], errs)
		}
	}
}

func (rule *Rule) Stringer() string {
	if rule.Type == "bool" && !slices.Contains([]string{"true", "false"}, rule.Operator) {
		return "N/A"
	}
	if rule.Type == "bool" {
		return rule.Operator
	}
	if rule.Type == "attribute" && rule.BoolValue == "" {
		return "N/A"
	}
	if rule.Type == "attribute" {
		return rule.BoolValue
	}
	if rule.Type == "comparison" {
		return rule.BoolValue
	}
	if rule.Type == "group" && rule.BoolValue != "" {
		return "( " + rule.BoolValue + " )"
	}
	if rule.Type == "group" {
		values := []string{}
		for _, child := range rule.Items {
			values = append(values, child.Stringer())
		}

		return "( " + strings.Join(values, " "+rule.Operator+" ") + " )"
	}

	return "INVALID RULE"
}

func (rule *Rule) Evaluate(input map[string]string) (*Rule, error) {
	cop := &Rule{}
	cloneRule(rule, cop)

	err := evaluateRule(cop, input)
	if err != nil {
		return nil, err
	}

	return cop, nil
}

//nolint:gocognit
func evaluateRule(rule *Rule, input map[string]string) error {
	if rule.Type == "bool" {
		rule.BoolValue = rule.Operator

		return nil
	}
	if rule.Type == "attribute" {
		if rule.BoolValue != "" {
			return nil
		}
		inputField, ok := input[rule.Attribute.Name]
		if !ok {
			rule.BoolValue = sqlCompileTarget(rule.Operator, rule.ParsedTarget, rule.Attribute)
		} else {
			boolValue := evaluateTarget(rule.Operator, rule.ParsedTarget, inputField, rule.Attribute.Kind)
			rule.BoolValue = strconv.FormatBool(boolValue)
		}

		return nil
	}

	if rule.Type == "comparison" {
		if rule.BoolValue != "" {
			return nil
		}

		var attr1, attr2 RuleAttribute
		attr1 = rule.Attributes[0]
		attr2 = rule.Attributes[1]

		t1 := attr1.Name
		t2 := attr2.Name

		_, ok := input[attr1.Name]
		if ok {
			t1 = input[attr1.Name]
			if attr1.Kind == "string" {
				t1 = "'" + input[attr1.Name] + "'"
			}
		}

		_, ok = input[attr2.Name]
		if ok {
			t2 = input[attr2.Name]
			if attr2.Kind == "string" {
				t2 = "'" + input[attr2.Name] + "'"
			}
		}

		rule.BoolValue = fmt.Sprintf("%s = %s", t1, t2)

		return nil
	}

	//nolint:nestif
	if rule.Type == "group" {
		if rule.BoolValue != "" {
			return nil
		}
		cancelItems := false
		allTrue := true
		allFalse := true
		for i := range rule.Items {
			err := evaluateRule(&rule.Items[i], input)
			if err != nil {
				return err
			}
			if rule.Items[i].BoolValue == "false" && rule.Operator == "AND" {
				rule.BoolValue = "false"
				cancelItems = true
				break
			}
			if rule.Items[i].BoolValue == "true" && rule.Operator == "OR" {
				rule.BoolValue = "true"
				cancelItems = true
				break
			}
			if rule.Items[i].BoolValue == "false" {
				allTrue = false
			}
			if rule.Items[i].BoolValue == "true" {
				allFalse = false
			}
			if rule.Items[i].BoolValue != "true" && rule.Items[i].BoolValue != "false" {
				allTrue = false
				allFalse = false
			}
		}
		if cancelItems {
			rule.Items = nil
		} else {
			if allTrue && rule.Operator == "AND" {
				rule.BoolValue = "true"
				rule.Items = nil
			}
			if allFalse && rule.Operator == "OR" {
				rule.BoolValue = "false"
				rule.Items = nil
			}
		}

		return nil
	}

	return fmt.Errorf("allowed rule types are `attribute`, `bool` or `group` got: `%s`", rule.Type)
}

func cloneRule(src *Rule, dist *Rule) {
	dist.Type = src.Type
	dist.Name = src.Name
	dist.Attribute = src.Attribute
	dist.Attributes = src.Attributes
	dist.Assert = src.Assert
	dist.ParsedTarget = src.ParsedTarget
	dist.BoolValue = src.BoolValue
	dist.Operator = src.Operator
	dist.Items = nil

	if src.Type != "group" {
		return
	}

	dist.Items = make([]Rule, len(src.Items))

	for i := range src.Items {
		cloneRule(&src.Items[i], &dist.Items[i])
	}
}
