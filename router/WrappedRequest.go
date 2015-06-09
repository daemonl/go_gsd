package router

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/daemonl/go_gsd/shared"
)

type WrappedRequest struct {
	shared.IRequest
	route *route
}

func wrapRequest(tr shared.IRequest, route *route) shared.IPathRequest {
	return &WrappedRequest{
		IRequest: tr,
		route:    route,
	}
}

func (wr *WrappedRequest) ScanPath(dests ...interface{}) error {
	_, r := wr.GetRaw()

	urlParts := wr.route.re.FindStringSubmatch(r.URL.RequestURI())

	if len(urlParts) != len(dests)+1 {
		fmt.Println(urlParts)
		return fmt.Errorf("scanning '%s', %d parts into %d parts (%s), length mismatch", r.URL.Path, len(urlParts), len(dests), wr.route.re.String())
	}

	for idx, dest := range dests {
		raw, err := url.QueryUnescape(urlParts[idx+1])
		if err != nil {
			return fmt.Errorf("scanning '%s' into '%s': %s", r.URL.Path, wr.route.re.String(), err.Error())
		}
		switch d := dest.(type) {
		case *string:
			*d = raw
		case *uint64:
			num, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = num
		case *int64:
			num, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = num
		case *uint32:
			num, err := strconv.ParseUint(raw, 10, 32)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = uint32(num)
		case *int32:
			num, err := strconv.ParseInt(raw, 10, 32)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = int32(num)
		default:
			return fmt.Errorf("Type %T not implemented for URL matching", d)
		}
	}

	return nil
}
