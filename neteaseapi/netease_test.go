package neteaseapi

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("NETEASE_PASSWORD", "heyuheng1.22.3")
	os.Setenv("NETEASE_PHONE", "18681655914")
}

func TestNetEaseContext_getNewCommendMusic(t *testing.T) {
	os.Setenv("NETEASE_PASSWORD", "heyuheng1.22.3")
	os.Setenv("NETEASE_PHONE", "18681655914")
	ctx := &NetEaseContext{}
	ctx.LoginNetEase()
	ctx.GetNewRecommendMusic()
}

func TestTime(t *testing.T) {
	for {
		time.Sleep(10 * time.Second)
		if string(time.Now().Local().Format("15:04:05")) == "04:54:00" {
			fmt.Println("test")
			break
		}
	}

	fmt.Println(time.Now().Local().Format("15:04:05")) // 13:57:52
}

func TestNetEaseContext_LoginNetEase(t *testing.T) {
	type fields struct {
		cookies []*http.Cookie
		err     error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &NetEaseContext{
				cookies: tt.fields.cookies,
				err:     tt.fields.err,
			}
			if err := ctx.LoginNetEase(); (err != nil) != tt.wantErr {
				t.Errorf("NetEaseContext.LoginNetEase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
