package dbpack

import (
	"errors"
	"log"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type khlNetease struct {
	gorm.Model
	KaiheilaID      string `gorm:"primaryKey;autoIncrement:false"`
	NetEaseID       string `gorm:"primaryKey;autoIncrement:false"`
	NetEasePhone    string
	NetEasePassword string
}

type khlMusicDownload struct {
	gorm.Model
	SongID   string `gorm:"primaryKey;autoIncrement:false"`
	Filepath string `gorm:"primaryKey;autoIncrement:false"`
}

func (music *khlMusicDownload) DownloadMusicDB() {
	db := GetDbConnection()
	err := db.AutoMigrate(&khlMusicDownload{})
	if err != nil {
		log.Println(err.Error())
	}
}

// CheckIsAdmin 检查是否是管理员
//  @param userID
//  @return isAdmin
func CheckIsAdmin(userID string) (isAdmin bool) {
	db := GetDbConnection()
	userIDInt, _ := strconv.Atoi(userID)
	res := db.Find(&Administrator{}, []int{userIDInt})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false
	}
	if errors.Is(res.Error, nil) {
		return true
	}
	return
}

// RegistAndBind 注册并绑定
//  @param data
//  @return err
func RegistAndBind(data *khlNetease) (err error) {
	dsn := "host=localhost user=postgres password=heyuheng1.22.3 dbname=betago port=55433 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema
	err = db.AutoMigrate(&khlNetease{})
	if err != nil {
		log.Println(err.Error())
	}
	res := db.Save(data)
	if res.Error != nil {
		err = res.Error
		return
	}
	res = db.Create(data)
	if res.Error != nil {
		err = res.Error
		return
	}
	// // Create
	// db.Create(&Product{Code: "D42", Price: 100})

	// // Read
	// var product Product
	// db.First(&product, 1)                 // 根据整形主键查找
	// db.First(&product, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

	// // Update - 将 product 的 price 更新为 200
	// db.Model(&product).Update("Price", 200)
	// // Update - 更新多个字段
	// db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
	// db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// // Delete - 删除 product
	// db.Delete(&product, 1)
	return
}
