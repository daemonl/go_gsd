package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/types"
)

type csvResponse struct {
	rows     []map[string]interface{}
	filename string
	//fields   []CSVField
	colNames     []string
	mappedFields map[string]*databath.Field
}

var DisplayTimezone *time.Location
var DateTimeFormat = "2006-01-02 15:04"

func init() {
	var err error
	DisplayTimezone, err = time.LoadLocation("Australia/Melbourne")
	if err != nil {
		panic("Pre loading melbourne timezone failed")
	}
}

func (r *csvResponse) ContentType() string {
	return "application/pdf"

}
func (r *csvResponse) WriteTo(out io.Writer) error {

	csvWriter := csv.NewWriter(out)

	csvWriter.Write(r.colNames)

	// TODO: Type assertions etc are done on each row... this seems rather inefficient.

	for _, row := range r.rows {

		record := make([]string, len(r.colNames), len(r.colNames))
		for i, colName := range r.colNames {
			v, ok := row[colName]

			if ok && v != nil {
				field, ok := r.mappedFields[colName]
				if !ok {
					log.Printf("No field %s in %#v\n", colName, r.mappedFields)
				}

				switch fieldImpl := field.Impl.(type) {
				case *types.FieldEnum:
					record[i] = fieldImpl.Choices[v.(string)]
				case *types.FieldDateTime, *types.FieldTimestamp:
					d := time.Unix(v.(int64), 0)
					record[i] = d.In(DisplayTimezone).Format(DateTimeFormat)
				default:
					record[i] = fmt.Sprintf("%v", v)
				}
			} else {
				record[i] = ""
			}
		}

		csvWriter.Write(record)
	}
	csvWriter.Flush()
	return nil
}

func (r *csvResponse) HTTPExtra(w http.ResponseWriter) {

	w.Header().Add("content-type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename="+r.filename)

}
