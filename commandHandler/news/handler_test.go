package news

import "testing"

func TestHandler(t *testing.T) {
	type args struct {
		targetID string
		quoteID  string
		authorID string
		args     []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Handler(tt.args.targetID, tt.args.quoteID, tt.args.authorID, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Handler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
