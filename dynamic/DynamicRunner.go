package dynamic

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/robertkrimen/otto"

	"github.com/daemonl/go_gsd/components"
)

type DynamicRunner struct {
	BaseDirectory string
	Mailer        components.Mailer
	Xero          components.Xero
}

func (dr *DynamicRunner) RunScript(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error) {

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
	rc.otto.Set("runScript", rc.RunScript)
	rc.otto.Set("sendMail", rc.SendMail)
	rc.otto.Set("fail", rc.Fail)
	rc.otto.Set("setResponseVal", rc.SetResponseVal)
	rc.otto.Set("end", rc.End)

	//if dr.Xero != nil {
	rc.otto.Set("XERO_Post", rc.XERO_Post)
	//}

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
