package shared_structs

type ActionSummary struct {
	UserId     uint64
	Action     string
	Collection string
	Pk         uint64
	Fields     map[string]interface{}
}
