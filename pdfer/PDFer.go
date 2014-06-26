package pdfer

import (
	"github.com/daemonl/go_gsd/shared"
)

type PDFer struct {
	Binary string
}

func (p *PDFer) ResponseAsPDF(htmlResponse shared.IResponse) shared.IResponse {
	resp := &PDFResponse{
		binary: p.Binary,
		htmlIn: htmlResponse,
	}
	return resp
}
