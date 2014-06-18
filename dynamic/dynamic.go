package dynamic

import (
	"database/sql"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DynamicRunner struct {
	BaseDirectory string
	SendMail      func(to string, subject string, body string)
}

type RunContext struct {
	otto         *otto.Otto
	errorMessage string
	runner       *DynamicRunner
	db           *sql.DB
	Response     map[string]interface{}
	EndChan      chan bool
}

func (dr *DynamicRunner) Run(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error) {

	log.Println("OTTO FUNC START")
	file, err := os.Open(dr.BaseDirectory + filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	rc := RunContext{
		runner:   dr,
		otto:     otto.New(),
		db:       db,
		Response: make(map[string]interface{}),
		EndChan:  make(chan bool),
	}

	log.Println("OTTO SETUP COMPLETE")

	rc.otto.Interrupt = make(chan func())

	rc.otto.Set("args", parameters)
	rc.otto.Set("sqlExec", rc.SqlExec)
	rc.otto.Set("sqlQuery", rc.SqlQuery)
	rc.otto.Set("sendMail", rc.SendMail)
	rc.otto.Set("fail", rc.Fail)
	rc.otto.Set("setResponseVal", rc.SetResponseVal)
	rc.otto.Set("end", rc.End)

	log.Println("START")

	_, err = rc.otto.Run(string(script))

	log.Println("Natural End")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	timeout := time.After(3 * time.Second)
	select {
	case _ = <-rc.EndChan:

	case _ = <-timeout:
		rc.Err("Timeout")
		go func() {
			_ = <-rc.EndChan
		}()
	}

	log.Println(rc.Response)
	return rc.Response, nil
}

func (rc *RunContext) Stop() {
	log.Println("STOP OTTO")
	go func() {
		rc.EndChan <- true
	}()
	go func() {
		rc.otto.Interrupt <- func() {
			log.Println("Otto context halted")
		}
	}()
}

func (rc *RunContext) Err(message string) otto.Value {
	log.Printf("Error in otto run: %s\n", message)
	rc.errorMessage = message
	rc.Stop()
	return otto.NullValue()
}

func (rc *RunContext) Fail(call otto.FunctionCall) otto.Value {
	errorString, err := call.ArgumentList[0].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}
	return rc.Err(errorString)
}

func (rc *RunContext) End(call otto.FunctionCall) otto.Value {
	log.Println("END OTTO")
	go func() {
		rc.EndChan <- true
	}()
	return otto.NullValue()
}

func (rc *RunContext) SetResponseVal(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		return rc.Err("SetResponseVal must be (string, interface{})")
	}
	key, err := call.ArgumentList[0].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}
	val, err := call.ArgumentList[1].Export()
	if err != nil {
		return rc.Err(err.Error())
	}

	rc.Response[key] = val
	return otto.NullValue()
}

func (rc *RunContext) SendMail(call otto.FunctionCall) otto.Value {

	if len(call.ArgumentList) < 3 {
		return rc.Err("sendMail called with too few parameters")
	}

	to, err := call.ArgumentList[0].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}

	subject, err := call.ArgumentList[1].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}

	body, err := call.ArgumentList[2].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}

	rc.runner.SendMail(to, subject, body)

	return otto.NullValue()

}
func (rc *RunContext) SqlExec(call otto.FunctionCall) otto.Value {

	if len(call.ArgumentList) < 2 {
		return rc.Err("Sql query called with too few parameters")
	}

	sqlString, err := call.ArgumentList[0].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}

	sqlArgumentsRaw, err := call.Argument(1).Export()
	if err != nil {
		return rc.Err(err.Error())
	}

	sqlArguments, ok := sqlArgumentsRaw.([]interface{})
	if !ok {
		return rc.Err("Sql query parameter 2 must be an array")
	}

	log.Printf("EXEC: %s %#v\n", sqlString, sqlArguments)

	res, err := rc.db.Exec(sqlString, sqlArguments...)
	if err != nil {
		return rc.Err(err.Error())
	}

	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		log.Printf("Insert ID: %d\n", id)
		val, _ := otto.ToValue(id)
		return val
	}

	affected, err := res.RowsAffected()
	if err == nil {
		log.Printf("Rows: %d\n", affected)
		val, _ := otto.ToValue(affected)
		return val
	}

	val, _ := otto.ToValue(nil)
	return val
}

func (rc *RunContext) SqlQuery(call otto.FunctionCall) otto.Value {

	log.Println("QUERY")
	if len(call.ArgumentList) < 3 {
		return rc.Err("Sql query called with too few parameters")
	}

	sqlString, err := call.ArgumentList[0].ToString()
	if err != nil {
		return rc.Err(err.Error())
	}

	sqlArgumentsRaw, err := call.Argument(1).Export()
	if err != nil {
		return rc.Err(err.Error())
	}

	sqlArguments, ok := sqlArgumentsRaw.([]interface{})
	if !ok {
		return rc.Err("Sql query parameter 2 must be an array")
	}

	if !call.ArgumentList[2].IsFunction() {
		return rc.Err("Argument 3 should be the callback function")
	}

	onRowCallback := call.Argument(2)

	log.Printf("Otto SQL: \"%s\"\n", sqlString)

	res, err := rc.db.Query(sqlString, sqlArguments...)
	if err != nil {
		return rc.Err(err.Error())
	}
	defer res.Close()

	cols, err := res.Columns()
	if err != nil {
		return rc.Err(err.Error())
	}

	for res.Next() {
		rowInterfaces := make([]*string, len(cols), len(cols))
		rowScanInterfaces := make([]interface{}, len(cols), len(cols))
		for i := range rowInterfaces {
			rowScanInterfaces[i] = &(rowInterfaces[i])
		}

		err := res.Scan(rowScanInterfaces...)
		if err != nil {
			return rc.Err(err.Error())
		}

		ottoVals := make([]interface{}, len(cols), len(cols))

		for i := range rowInterfaces {
			val := rowInterfaces[i]
			if val == nil {
				ottoVals[i] = otto.NullValue()
			} else {
				oval, err := otto.ToValue(*(rowInterfaces[i]))
				if err != nil {
					return rc.Err(err.Error())
				}
				ottoVals[i] = oval
			}

		}
		onRowCallback.Call(otto.NullValue(), ottoVals...)
	}

	log.Println("END QUERY")

	return otto.NullValue()
}
