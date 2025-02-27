package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	da "github.com/direktiv/direktiv-apps/pkg/direktivapps"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type input struct {
	Data struct {
		DB    map[string]interface{} `json:"db"`
		Where string                 `json:"where"`
		Table string                 `json:"table"`
	} `json:"data"`
}

const (
	errCode = "com.azure.%s"
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
		reportError(w, "inputUnmarshal", err)
		return
	}

	da.LogDouble(aid, "executing sql")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=require",
		obj.Data.DB["host"], obj.Data.DB["port"], obj.Data.DB["username"], obj.Data.DB["password"], obj.Data.DB["database"])

	fmt.Println(psqlInfo)

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		da.WriteError(da.ActionError{
			"io.direktiv.conn.error",
			err.Error(),
		})
		return
	}

	selectStmt := fmt.Sprintf(`select * from %s where %s`, obj.Data.Table, obj.Data.Where)

	rows, err := db.Queryx(selectStmt)
	if err != nil {
		fmt.Println(err)
		da.WriteError(da.ActionError{
			"io.direktiv.select.error",
			err.Error(),
		})
		return
	}
	defer rows.Close()

	type Row map[string]interface{}

	res := make([]Row, 0)
	for rows.Next() {
		results := Row{}
		err := rows.MapScan(results)
		if err != nil {
			continue
		}
		res = append(res, results)
	}

	bb, err := json.Marshal(res)
	if err != nil {
		da.WriteError(da.ActionError{
			"io.direktiv.json.error",
			err.Error(),
		})
		return
	}

	w.Write(bb)
}
