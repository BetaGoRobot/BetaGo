package utility

import (
	"fmt"
	"regexp"
	"strings"
)

func RemoveArgFromStr(s string, args ...string) string {
	re := fmt.Sprintf(`\s+--(%s)=(?:\"[^\"]*\"|[^\s]+)`, strings.Join(args, "|"))
	re = regexp.MustCompile(re).ReplaceAllString(s, "")
	return strings.Join(strings.Fields(re), " ")
}
