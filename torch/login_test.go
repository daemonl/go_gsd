package torch

import (
	"fmt"
	"github.com/daemonl/go_lib/databath"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {

	parser := Parser{
		Store: InMemorySessionStore(),
		Bath:  databath.RunABath("mysql", "root:@tcp(athena.local:3306)/gsd_test", 1),
	}

	req, err := http.NewRequest("GET", "http://localhost/check", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	handler := parser.WrapReturn(HandleLogin)
	parsedRequest := handler(w, req)
	fmt.Println(w.Header().Get("Set-Cookie"))
	fmt.Println(*parsedRequest.Session.Key)

}
