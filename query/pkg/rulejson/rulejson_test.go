package rulejson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestEvaluateTarget(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		target    any
		input     string
		kind      string
		wantBool  bool
		wantError bool
	}{
		// simple case.
		{
			name:     "valid equal",
			operator: "equal",
			target: &TargetValue{
				Value: "123",
			},
			input:     "123",
			kind:      "number",
			wantBool:  true,
			wantError: false,
		},
		// simple case.
		{
			name:     "invalid equal",
			operator: "equal",
			target: &TargetValue{
				Value: "124",
			},
			input:     "123",
			kind:      "number",
			wantBool:  false,
			wantError: true,
		},
		// is_substring_of_1
		{
			name:     "is_substring_of_1",
			operator: "isSubstringOf",
			target: &TargetValue{
				Value: "123456",
			},
			input:     "56",
			kind:      "string",
			wantBool:  true,
			wantError: false,
		},
		// is_substring_of_1
		{
			name:     "is_substring_of_1",
			operator: "isSubstringOf",
			target: &TargetValue{
				Value: "123456",
			},
			input:     "12",
			kind:      "string",
			wantBool:  true,
			wantError: false,
		},
		// is_substring_of_1
		{
			name:     "is_substring_of_1",
			operator: "isSubstringOf",
			target: &TargetValue{
				Value: "123456",
			},
			input:     "56",
			kind:      "string",
			wantBool:  true,
			wantError: false,
		},
		// is_substring_of_2
		{
			name:     "is_substring_of_2",
			operator: "isSubstringOf",
			target: &TargetValue{
				Value: "123456",
			},
			input:     "46",
			kind:      "string",
			wantBool:  false,
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := evaluateTarget(tt.operator, tt.target, tt.input, tt.kind)
			if !reflect.DeepEqual(got, tt.wantBool) {
				t.Errorf("evaluateTarget() got = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

func TestEvaluateRuleWithJson(t *testing.T) {
	policy := `
     {
        "type": "group",
        "operator": "OR",
        "items": [
          {
            "type": "attribute",
            "attribute": {"name": "userLevel", "kind": "number"},
            "operator": "equal",
            "assert": {
              "value": "4"
            }
          },
          {
            "type": "group",
            "operator": "AND",
            "items": [
              {
                "type": "attribute",
                "attribute": {"name": "userLevel", "kind": "number"},
                "operator": "equal",
                "assert": {
                  "value": "3"
                }
              },
              {
                "type": "attribute",
                "attribute": {"name": "secretLevel", "kind": "number"},
                "operator": "range",
                "assert": {
                  "from": "1",
                  "to": "3"
                }
              }
            ]
          },
          {
            "type": "group",
            "operator": "AND",
            "items": [
              {
                "type": "attribute",
                "attribute": {"name": "userLevel", "kind": "number"},
                "operator": "equal",
                "assert": {
                  "value": "2"
                }
              },
              {
                "type": "attribute",
                "attribute": {"name": "secretLevel", "kind": "number"},
                "operator": "range",
                "assert": {
                  "from": "1",
                  "to": "2"
                }
              }
            ]
          },
          {
            "type": "group",
            "operator": "AND",
            "items": [
              {
                "type": "attribute",
                "attribute": {"name": "userLevel", "kind": "number"},
                "operator": "equal",
                "assert": {
                  "value": "1"
                }
              },
              {
                "type": "attribute",
                "attribute": {"name": "secretLevel", "kind": "number"},
                "operator": "equal",
                "assert": {
                  "value": "1"
                }
              }
            ]
          }
        ]
      }`

	rule := &Rule{}
	err := json.Unmarshal([]byte(policy), rule)
	if err != nil {
		t.Errorf("failed to unmarshal json: %v", err)
	}
	rErr := Validate(rule)
	if len(rErr) != 0 {
		t.Errorf("failed to validate rule: %v", rErr)
	}
	rule, err = rule.Evaluate(map[string]string{
		"userLevel":   "1",
		"userCountry": "de",
	})
	if err != nil {
		t.Errorf("failed to evaluate rule: %v", err)
	}

	wantString := "( false OR ( false ) OR ( false ) OR ( true AND secretLevel = 1 ) )"
	if rule.Stringer() != wantString {
		t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), wantString)
	}

	rule, err = rule.Evaluate(map[string]string{
		"secretLevel": "1",
	})
	if err != nil {
		t.Errorf("failed to evaluate rule: %v", err)
	}
	wantString = "( false OR ( false ) OR ( false ) OR ( true AND secretLevel = 1 ) )"
	if rule.Stringer() != wantString {
		t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), wantString)
	}

}

func TestEvaluateRule(t *testing.T) {

	tests := []struct {
		name       string
		rule       *Rule
		input      map[string]string
		wantString string
		wantErr    bool
	}{
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule comparison",
			rule: &Rule{
				Type:     "comparison",
				Operator: "equal",
				Attributes: []RuleAttribute{
					{
						Name: "doc.City",
						Kind: "string",
					},
					{
						Name: "user.City",
						Kind: "string",
					},
				},
			},
			input: map[string]string{
				"user.City": "New York",
			},
			wantString: `doc.City = 'New York'`,
			wantErr:    false,
		},

		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule comparison",
			rule: &Rule{
				Type:     "comparison",
				Operator: "equal",
				Attributes: []RuleAttribute{
					{
						Name: "doc.City",
						Kind: "string",
					},
					{
						Name: "user.City",
						Kind: "string",
					},
				},
			},
			input:      map[string]string{},
			wantString: `doc.City = user.City`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule comparison",
			rule: &Rule{
				Type:     "comparison",
				Operator: "equal",
				Attributes: []RuleAttribute{
					{
						Name: "doc.Level",
						Kind: "number",
					},
					{
						Name: "user.Level",
						Kind: "number",
					},
				},
			},
			input: map[string]string{
				"user.Level": "25",
			},
			wantString: `doc.Level = 25`,
			wantErr:    false,
		},

		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "bool",
				Operator: "true",
			},
			input: map[string]string{
				"user.age": "25",
			},
			wantString: `true`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "bool",
				Operator: "false",
			},
			input: map[string]string{
				"user.age": "25",
			},
			wantString: `false`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:      "attribute",
				Operator:  "equal",
				Attribute: RuleAttribute{Name: "user.age", Kind: "number"},
				Assert:    json.RawMessage(`{"value": "25"}`),
			},
			input: map[string]string{
				"user.age": "25",
			},
			wantString: `true`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:      "attribute",
				Operator:  "isSubstringOf",
				Attribute: RuleAttribute{Name: "user.city", Kind: "string"},
				Assert:    json.RawMessage(`{"value": "York2"}`),
			},
			input: map[string]string{
				"user.city": "New York",
			},
			wantString: `false`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:      "attribute",
				Operator:  "isSubstringOf",
				Attribute: RuleAttribute{Name: "user.city", Kind: "string"},
				Assert:    json.RawMessage(`{"value": "New York"}`),
			},
			input: map[string]string{
				"user.city": "York",
			},
			wantString: `true`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "group",
				Operator: "AND",
				Items: []Rule{
					{
						Type:      "attribute",
						Operator:  "equal",
						Attribute: RuleAttribute{Name: "user.age", Kind: "number"},
						Assert:    json.RawMessage(`{"value": "25"}`),
					},
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.gender",
							Kind: "string",
						},
						Assert: json.RawMessage(`{"value": "male"}`),
					},
				},
			},
			input: map[string]string{
				"user.age":    "25",
				"user.gender": "male",
			},
			wantString: `( true )`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////

		{
			name: "valid rule",
			rule: &Rule{
				Type:     "group",
				Operator: "AND",
				Items: []Rule{
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.age",
							Kind: "number",
						},
						Assert: json.RawMessage(`{"value": "25"}`),
					},
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.gender",
							Kind: "string",
						},
						Assert: json.RawMessage(`{"value": "male"}`),
					},
				},
			},
			input: map[string]string{
				"user.age":    "26",
				"user.gender": "male",
			},
			wantString: `( false )`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "group",
				Operator: "AND",
				Items: []Rule{
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.age",
							Kind: "number",
						},
						Assert: json.RawMessage(`{"value": "25"}`),
					},
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.gender",
							Kind: "string",
						},
						Assert: json.RawMessage(`{"value": "male"}`),
					},
				},
			},
			input: map[string]string{
				"user.gender": "male",
			},
			wantString: `( user.age = 25 AND true )`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "group",
				Operator: "AND",
				Items: []Rule{
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.age",
							Kind: "number",
						},
						Assert: json.RawMessage(`{"value": "25"}`),
					},
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.gender",
							Kind: "string",
						},
						Assert: json.RawMessage(`{"value": "male"}`),
					},
				},
			},
			input: map[string]string{
				"user.age": "25",
			},
			wantString: `( true AND user.gender = "male" )`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
		{
			name: "valid rule",
			rule: &Rule{
				Type:     "group",
				Operator: "OR",
				Items: []Rule{
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.age",
							Kind: "number",
						},
						Assert: json.RawMessage(`{"value": "25"}`),
					},
					{
						Type:     "attribute",
						Operator: "equal",
						Attribute: RuleAttribute{
							Name: "user.gender",
							Kind: "string",
						},
						Assert: json.RawMessage(`{"value": "male"}`),
					},
				},
			},
			input: map[string]string{
				"user.age": "26",
			},
			wantString: `( false OR user.gender = "male" )`,
			wantErr:    false,
		},
		////////////////////////////////////////////////////////////////////////////////////////
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.rule); err != nil {
				t.Fatalf("failed to validate rule: %v", err)
			}
			rule, err := tt.rule.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateRule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rule.Stringer() != tt.wantString {
				t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), tt.wantString)
			}
		})
	}
}

func TestEvaluateRuleMultipleTimes(t *testing.T) {
	rule := &Rule{
		Type:     "group",
		Operator: "AND",
		Items: []Rule{
			{
				Type:     "bool",
				Operator: "true",
			},
			{
				Type:     "attribute",
				Operator: "equal",
				Attribute: RuleAttribute{
					Name: "user.age",
					Kind: "number",
				},
				Assert: json.RawMessage(`{"value": "25"}`),
			},
			{
				Type:     "attribute",
				Operator: "equal",
				Attribute: RuleAttribute{
					Name: "user.gender",
					Kind: "string",
				},
				Assert: json.RawMessage(`{"value": "male"}`),
			},
		},
	}
	input := map[string]string{
		"user.gender": "male",
	}
	wantString := `( true AND user.age = 25 AND true )`
	wantErr := false

	if err := Validate(rule); err != nil {
		t.Fatalf("failed to validate rule: %v", err)
	}
	rule, err := rule.Evaluate(input)
	if (err != nil) != wantErr {
		t.Errorf("EvaluateRule() error = %v, wantErr %v", err, wantErr)
	}
	if rule == nil {
		t.Fatalf("EvaluateRule() got = <nil>, want <nil>")
	}
	if rule.Stringer() != wantString {
		t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), wantString)
	}

	input = map[string]string{
		"user.age": "25",
	}
	wantString = `( true AND user.age = 25 AND true )`
	wantErr = false

	rule, err = rule.Evaluate(input)
	if (err != nil) != wantErr {
		t.Errorf("EvaluateRule() error = %v, wantErr %v", err, wantErr)
	}
	if rule == nil {
		t.Fatalf("EvaluateRule() got = <nil>, want <nil>")
	}
	if rule.Stringer() != wantString {
		t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), wantString)
	}
}

func TestEvaluateRuleWithJsonComplex(t *testing.T) {
	policy := `{
      "type": "group",
      "operator": "AND",
      "items": [
        {
          "type": "bool",
          "operator": "true"
        },
        {
          "type": "group",
          "operator": "OR",
          "items": [
            {
              "type": "group",
              "operator": "AND",
              "items": [
                {
                  "type": "attribute",
                  "operator": "equal",
                  "assert": {
                    "value": "street1"
                  },
                  "attribute": {
                    "name": "user.street",
                    "kind": "string"
                  }
                }
              ]
            },
            {
              "type": "group",
              "operator": "AND",
              "items": [
                {
                  "type": "attribute",
                  "operator": "equal",
                  "assert": {
                    "value": "Berlin"
                  },
                  "attribute": {
                    "name": "user.city",
                    "kind": "string"
                  }
                }
              ]
            }
          ]
        },
        {
          "type": "attribute",
          "operator": "equal",
          "assert": {
            "value": "hello"
          },
          "attribute": {
            "name": "data.work_order",
            "kind": "string"
          }
        },
        {
          "type": "group",
          "operator": "OR",
          "items": [
            {
              "type": "group",
              "operator": "AND",
              "items": [
                {
                  "type": "attribute",
                  "operator": "equal",
                  "assert": {
                    "value": "world"
                  },
                  "attribute": {
                    "name": "data.work_order",
                    "kind": "string"
                  }
                }
              ]
            },
            {
              "type": "group",
              "operator": "AND",
              "items": [
                {
                  "type": "attribute",
                  "operator": "equal",
                  "assert": {
                    "value": "world3"
                  },
                  "attribute": {
                    "name": "data.work_order",
                    "kind": "string"
                  }
                }
              ]
            }
          ]
        }
      ]
    }`

	rule := &Rule{}
	err := json.Unmarshal([]byte(policy), rule)
	if err != nil {
		t.Errorf("failed to unmarshal json: %v", err)
	}

	rErr := Validate(rule)
	if len(rErr) != 0 {
		t.Errorf("failed to validate rule: %v", rErr)
	}

	rule, err = rule.Evaluate(map[string]string{
		"userLevel":   "1",
		"userCountry": "de",
		"user.street": "street1",
	})
	if err != nil {
		t.Errorf("failed to evaluate rule: %v", err)
	}

	wantString := `( true AND ( true ) AND data.work_order = "hello" AND ( ( data.work_order = "world" ) OR ( data.work_order = "world3" ) ) )`
	if rule.Stringer() != wantString {
		t.Errorf("EvaluateRule() got = >%v<, want >%v<", rule.Stringer(), wantString)
	}
}
