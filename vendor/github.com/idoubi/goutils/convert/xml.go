package convert

import (
	"bytes"
	"fmt"

	xml2json "github.com/basgys/goxml2json"
	"github.com/yoda-of-soda/map2xml"
)

// Xml2Json xml转换成json
func Xml2Json(x []byte) ([]byte, error) {
	xr := bytes.NewReader(x)
	j, err := xml2json.Convert(xr)
	if err != nil {
		return nil, err
	}

	return j.Bytes(), nil
}

// Map2Xml: transfer map[string]interface{} to xml byte
func Map2Xml(m map[string]interface{}) (b []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	x := map2xml.New(m)
	x.WithIndent("", "  ")
	x.WithRoot("xml", nil)
	b, err = x.Marshal()

	return
}
