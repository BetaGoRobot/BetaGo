package main

import (
	"fmt"

	"github.com/dlclark/regexp2"
)

var re = regexp2.MustCompile(`((@[^ ]+\s+)|^)\/(?P<commands>\w+( )*)+( )*(--(?P<arg_name>\w+)=(?P<arg_value>("[^"]*"|\S+)))*`, regexp2.RE2)

func main() {
	s := "The Chainsmokers新歌的男声是谁呀？没错，就是The Chainsmokers成员中的Drew！长得超帅而且会唱歌也是没谁了！"
	fmt.Println(len(s))
}
