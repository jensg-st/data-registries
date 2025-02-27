package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"query/pkg/rulejson"
	"strings"

	// "github.com/Azure/azure-sdk-for-go/sdk/azcore/internal/exported"

	da "github.com/direktiv/direktiv-apps/pkg/direktivapps"
)

type UserAttribute struct {
	Description  string `json:"description"`
	Name         string `json:"name"`
	Value        string `json:"value"`
	ID           any    `json:"id"`
	SrcAttr      any    `json:"srcAttr"`
	SrcDirectory any    `json:"srcDirectory"`
	SrcName      any    `json:"srcName"`
	Typ          any    `json:"type"`
	RegExp       any    `json:"regExp"`
	Project      any    `json:"project"`
}

type input struct {
	Query struct {
		Policies []rulejson.Rule `json:"policies"`
		User     []UserAttribute `json:"user"`
	} `json:"query"`
}

const (
	errCode = "com.query.%s"
)

func main() {
	da.StartServer(coreLogic)
}

func reportError(w http.ResponseWriter, code string, err error) {
	da.RespondWithError(w, fmt.Sprintf(errCode, code), err.Error())
}

func coreLogic(w http.ResponseWriter, r *http.Request) {
	obj := new(input)
	aid, err := da.Unmarshal(obj, r)
	if err != nil {
		fmt.Println(err)
		reportError(w, "inputUnmarshal", err)
		return
	}

	da.LogDouble(aid, "Hello")

	userAttrs := map[string]string{}
	for i := range obj.Query.User {
		userAttrs[obj.Query.User[i].Name] = obj.Query.User[i].Value
	}

	whereClauses := []string{}
	for i := range obj.Query.Policies {
		rule := &obj.Query.Policies[i]
		rErr := rulejson.Validate(rule)
		if len(rErr) != 0 {
			writeError(w, fmt.Sprintf("Policy validation error: %v", rErr))
			return
		}

		rule, err = rule.Evaluate(userAttrs)
		if err != nil {
			writeError(w, fmt.Sprintf("Policy evaluation error: %v", rErr))
			return
		}

		rule.Stringer()
		whereClauses = append(whereClauses, rule.Stringer())
	}

	result := "WHERE (" + strings.Join(whereClauses, ") OR (") + ")"

	writeJSON(w, result)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}

func writeError(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	payLoad := struct {
		Error any `json:"error"`
	}{
		Error: err,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}
