package dbpack

import "testing"

func TestRegistAndBind(t *testing.T) {
	RegistAndBind(&khlNetease{KaiheilaID: "123", NetEaseID: "kevinmatt", NetEasePhone: "1111111", NetEasePassword: "adadas"})
}
