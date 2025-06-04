package cts

import "golang.org/x/exp/constraints"

type ValidType interface {
	constraints.Integer | constraints.Float | string
}

type Numeric interface {
	constraints.Integer | constraints.Float
}
