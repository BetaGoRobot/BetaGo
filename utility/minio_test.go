package utility

import (
	"context"
	"fmt"
	"testing"
)

func TestUploadFile(t *testing.T) {
	// MinioUploadFile()
	ctx := context.Background()
	fmt.Println(MinioUploadFileFromURL(ctx, "testbucket", "http://m701.music.126.net/20240506181042/07be99afc29737a9c93766ee1dbd879b/jdymusic/obj/wo3DlMOGwrbDjj7DisKw/22259855829/13af/ffc0/d76b/375da125883ef3487cb160dde8258b9b.flac", "test.flac", "audio/mpeg;charset=UTF-8"))
}
