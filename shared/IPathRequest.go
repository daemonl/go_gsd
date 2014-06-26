package shared

type IPathRequest interface {
	IRequest
	ScanPath(dests ...interface{}) error
}
