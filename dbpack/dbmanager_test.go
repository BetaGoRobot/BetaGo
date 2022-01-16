package dbpack

import (
	"testing"

	"gorm.io/gorm"
)

func TestA(t *testing.T) {
	khl := khlMusicDownload{}
	khl.DownloadMusicDB()
}
func TestRegistAndBind(t *testing.T) {
	RegistAndBind(&khlNetease{KaiheilaID: "123", NetEaseID: "kevinmatt", NetEasePhone: "1111111", NetEasePassword: "adadas"})
}

func Test_khlMusicDownload_DownloadMusicDB(t *testing.T) {
	type fields struct {
		Model    gorm.Model
		SongID   string
		Filepath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			music := &khlMusicDownload{
				Model:    tt.fields.Model,
				SongID:   tt.fields.SongID,
				Filepath: tt.fields.Filepath,
			}
			music.DownloadMusicDB()
		})
	}
}
