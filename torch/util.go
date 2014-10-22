package torch

import (
	"io"
	"net/http"
)

type redirectResponse string

func getRedirectResponse(to string) (*redirectResponse, error) {
	return GetRedirectResponse(to)
}

func GetRedirectResponse(to string) (*redirectResponse, error) {
	rr := redirectResponse(to)
	return &rr, nil
}

func (r *redirectResponse) ContentType() string {
	return "text/plain"
}

func (r *redirectResponse) HTTPExtra(w http.ResponseWriter) {
	w.Header().Add("location", string(*r))
	w.WriteHeader(http.StatusSeeOther)
}

func (r *redirectResponse) WriteTo(w io.Writer) error {
	return nil
}
