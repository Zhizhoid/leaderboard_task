package main

import (
	"fmt"
	"time"
	"strconv"
	"math/rand"
	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	DBUser     		= "root"
	DBPassword 		= "mypassword1029"
	DBHost     		= "127.0.0.1"
	DBPort     		= "3306"
	DBName     		= "leaderboard"
	DBtype     		= "mysql"
	
	ValueAmount		= 100
)

type SqlHandler struct {
	db *gorm.DB
}

type Player struct {
	Name string
	Score int `gorm:primary_key`
	SubmittedAt time.Time
}

func OpenDB() SqlHandler {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error)
	}
	return SqlHandler{db}
}

func randate() time.Time {
    min := time.Date(2021, 6, 0, 0, 0, 0, 0, time.UTC).Unix()
    max := time.Date(2021, 8, 19, 24, 60, 60, 0, time.UTC).Unix()
    delta := max - min

    sec := rand.Int63n(delta) + min
    return time.Unix(sec, 0)
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error)
	}
	
	var player Player
	
	for i := 1; i <= ValueAmount; i++ {
		player.Name = "Player" + strconv.Itoa(i)
		player.Score = rand.Intn(1000)
		player.SubmittedAt = randate()
		db.Select("Name", "Score", "SubmittedAt").Create(&player)
	}
}
