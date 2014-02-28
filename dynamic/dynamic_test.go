package dynamic

import (
	"fmt"
	"github.com/daemonl/go_lib/databath"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestGeneral(t *testing.T) {

	bath := databath.RunABath("mysql", "root:scfaC6000@/pov", 1)
	dr := DynamicRunner{
		DataBath:      bath,
		BaseDirectory: "/home/daemonl/schkit/impl/pov/script/",
	}

	res, err := dr.Run("test.js")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
