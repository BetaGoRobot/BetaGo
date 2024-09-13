package utility

import (
	"github.com/bytedance/sonic"
)

func MustMashal(input any) string {
	res, err := sonic.MarshalString(input)
	if err != nil {
		panic(err)
	}
	return res
}

func AddressORNil[P any, T *P](input T) P {
	if input == nil {
		return *new(P)
	}
	return *input
}
