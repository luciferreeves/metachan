package meta

import "metachan/types"

type facade struct {
	req types.Request
}

type required struct {
	req types.Request
}

type withDefault struct {
	req types.Request
	def string
}
