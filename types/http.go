package types

type HTTPParam struct {
	Key   string
	Value string
}

type Request struct {
	Path        string
	Method      string
	Query       []HTTPParam
	Params      []HTTPParam
	Headers     []HTTPParam
	QueryString string
	IP          string
	URL         string
}
