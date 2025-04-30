package history

import (
	"context"
	"testing"

	"github.com/defensestation/osquery"
)

func TestHelper_GetMsg(t *testing.T) {
	New(context.Background()).
		Query(
			osquery.Bool().Must(
				osquery.
					Term("chat_id", "oc_dd3a0d0e70a484c8ea7fbed022f1538c"),
				osquery.
					Term("message_type", "post"),
			),
		).
		Size(uint64(100)).
		Sort("create_time", "desc").GetMsg()
}
