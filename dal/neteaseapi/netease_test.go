package neteaseapi

import (
	"context"
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
	ctx.LoginNetEase(context.Background())
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
			if err := ctx.LoginNetEase(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("NetEaseContext.LoginNetEase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNetEaseContext_GetMusicURL(t *testing.T) {
	NetEaseGCtx.GetMusicURL(context.Background(), "423228325")
}

func TestNetEaseContext_LoginNetEaseQR(t *testing.T) {
	ctx := context.Background()
	c := &NetEaseContext{
		retryCnt: 0,
		qrStruct: struct {
			isOutDated bool
			uniKey     string
			qrBase64   string
		}{},
	}
	c.LoginNetEaseQR(ctx)
}

func TestNetEaseContext_GetLyrics(t *testing.T) {
	NetEaseGCtx.GetLyrics(context.Background(), "423228325")
}
