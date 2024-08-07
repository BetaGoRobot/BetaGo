package main

import (
	"fmt"

	"github.com/dlclark/regexp2"
)

var re = regexp2.MustCompile(`((@[^ ]+\s+)|^)\/(?P<commands>\w+( )*)+( )*(--(?P<arg_name>\w+)=(?P<arg_value>("[^"]*"|\S+)))*`, regexp2.RE2)

func main() {
	s := "<p>@_user_1 这里</p>"
	m, _ := re.MatchString(s)
	fmt.Println(m)
}
