package main

import (
	"encoding/base64"
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
	fmt.Println("1")
	obj := new(input)
	aid, err := da.Unmarshal(obj, r)
	if err != nil {
		fmt.Println(err)
		reportError(w, "inputUnmarshal", err)
		return
	}
	fmt.Println("2")
	da.LogDouble(aid, "Hello")

	userAttrs := map[string]string{}
	for i := range obj.Query.User {
		userAttrs["user."+obj.Query.User[i].Name] = obj.Query.User[i].Value
	}
	fmt.Println("3")
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

		str := strings.ReplaceAll(rule.Stringer(), "data.", "")
		whereClauses = append(whereClauses, str)
	}

	result := strings.Join(whereClauses, " OR ")
	encoded := base64.StdEncoding.EncodeToString([]byte(result))
	fmt.Println("4")
	writeJSON(w, result, encoded)
}

func writeJSON(w http.ResponseWriter, data string, base64 string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data   any    `json:"data"`
		Base64 string `json:"base64"`
	}{
		Data:   data,
		Base64: base64,
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
