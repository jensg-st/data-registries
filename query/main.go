package main

import (
	"fmt"
	"net/http"

	// "github.com/Azure/azure-sdk-for-go/sdk/azcore/internal/exported"

	da "github.com/direktiv/direktiv-apps/pkg/direktivapps"
)

type input struct {
	Query struct {
		Policies interface{} `json:"policies"`
		User     interface{} `json:"user"`
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

	w.Write([]byte("{ \"where\": \"true\"}"))
}
