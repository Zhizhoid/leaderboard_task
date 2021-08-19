package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"time"
	"strings"
	"strconv"
	
	"github.com/labstack/echo/v4"
	"github.com/jinzhu/now"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	DBUser     		= "root"
	DBPassword 		= ""
	DBHost     		= "127.0.0.1"
	DBPort     		= "3306"
	DBName     		= "leaderboard"
	DBtype     		= "mysql"
	
	PageLength 		= 10
	MaxLimit		= 10000000000 // You can't use offset without limit
	PlayersAfter 	= 3
	ModeDefault		= 1 // Default leaderboard mode (0 - all time leaderboard, 1 - monthly leaderboard)
	
	AuthToken		= "token"
)

type SqlHandler struct {
	db *gorm.DB
}

type Player struct {
	Name string
	Score int `gorm:primary_key`
	SubmittedAt time.Time
}

type Score struct {
    Name string	`json:"name"`
    Score int   `json:"score"`
}

type RankedScore struct {
	Name string	`json:"name"`
	Score int	`json:"score"`
	Rank int	`json:"rank"`
}

type JSON_Response struct {
	Results []RankedScore	`json:"results"`
	NextPage int			`json:"next_page"`
}

type JSON_Response_With_Name struct {
	Results []RankedScore	`json:"results"`
	AroundMe []RankedScore	`json:"around_me"`
	NextPage int			`json:"next_page"`
}

func OpenDB() SqlHandler {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error)
	}
	return SqlHandler{db}
}

// storeScore stores new score into the database
func (handler *SqlHandler) storeScore(score Score) error {
	player := Player{Name: score.Name, Score: score.Score, SubmittedAt: time.Now()}
	
	res := handler.db.Select("Name", "Score", "SubmittedAt").Create(&player)
	
	return res.Error
}

// currentScore returns current score of a player & a boolean that shows if the player exists
func (handler *SqlHandler) currentScore(name string) (int, bool) { //true - Player already exists, false - Player doesn't exist yet
	var player Player
	
	res := handler.db.
			Where("name = ?", name).
			Limit(1).
			Find(&player)
	
	var found bool
	found = res.RowsAffected > 0
	
	return player.Score, found
}

// updateScore updates score of an existing player
func (handler *SqlHandler) updateScore(score Score) error {
	player := Player{Name: score.Name, Score: score.Score, SubmittedAt: time.Now()}
	res := handler.db.
			Model(&player).
			Where("name = ?", score.Name).
			Limit(1).
			Updates(player)
	
	return res.Error
}

// getPage returns all scores from given page
func (handler *SqlHandler) getPage(page int, mode int) ([]RankedScore, error) {
	var (
		players []Player
		res *gorm.DB
		)
		
	switch mode{
		case 0:
			res = handler.db.
				Order("score DESC, submitted_at ASC").
				Offset((page-1)*PageLength).
				Limit(PageLength).
				Find(&players)
		case 1:
			res = handler.db.
				Where("submitted_at >= ?", now.BeginningOfMonth()).
				Order("score DESC, submitted_at ASC").
				Offset((page-1)*PageLength).
				Limit(PageLength).
				Find(&players)
	}
	
	rScores := make([]RankedScore, 0, PageLength)
	for i, v := range players {
		rScores = append(rScores, RankedScore{Name: v.Name, Score: v.Score, Rank: (page-1)*PageLength + 1 + i})
	}
	
	return rScores, res.Error
}

// aroundMe returns scores of other players around the given player if he is on any page after the given one
func (handler *SqlHandler) aroundMe(currentPage int, name string, mode int) ([]RankedScore, bool) { //true - Player exists on some page after the given one, false - Player doesn't exist on any page after the given one
	var players []Player
		
		switch mode {
			case 0:
				handler.db.
				Order("score DESC, submitted_at ASC").
				Offset(currentPage*PageLength).
				Limit(MaxLimit).
				Find(&players)
			case 1:
				handler.db.
				Where("submitted_at >= ?", now.BeginningOfMonth()).
				Order("score DESC, submitted_at ASC").
				Offset(currentPage*PageLength).
				Limit(MaxLimit).
				Find(&players)
		}
		
		var (
			PlayerExists bool
			add int
			)
		
		for i, v := range players {
			if v.Name == name {
				switch {
					case i > 0 :
						switch {
							case (len(players) - i) > PlayersAfter :
								players = players[i-1:i+4]
							case (len(players) - i) <= PlayersAfter :
								players = players[i-1:len(players)]
						}
					case i == 0 :
						switch {
							case (len(players) - i) > PlayersAfter :
								players = players[i:i+4]
							case (len(players) - i) <= PlayersAfter :
								players = players[i:len(players)]
						}
						players = append(players, Player{})
						copy(players[1:], players)
						players[0] = handler.findPlayer(currentPage*PageLength, mode)
				}
				PlayerExists = true
				add = i
				break
			}
		}
		
		rScores := make([]RankedScore, 0)
	for i, v := range players {
		rScores = append(rScores, RankedScore{Name: v.Name, Score: v.Score, Rank: currentPage*PageLength + add + i})
	}
	
	return rScores, PlayerExists
}

// findPlayer finds player with the given rank
func (handler *SqlHandler) findPlayer(rank int, mode int) Player {
	var player Player
	
	switch mode{
		case 0:
			handler.db.
			Order("score DESC, submitted_at ASC").
			Offset(rank-1).
			Limit(1).
			Find(&player)
		case 1:
			handler.db.
			Where("submitted_at >= ?", now.BeginningOfMonth()).
			Order("score DESC, submitted_at ASC").
			Offset(rank-1).
			Limit(1).
			Find(&player)
	}
	
	return player
}

// STORE PLAYER'S SCORE //
func StoreScore(g *echo.Group, handler SqlHandler) {
	g.POST("/store", func(c echo.Context) error {
		score := Score{}
		
		// READING FROM JSON
		err := json.NewDecoder(c.Request().Body).Decode(&score)
		c.Request().Body.Close()
		
		if err != nil {
			log.Printf("Failed reading in StoreScore: %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		
		//STORING SCORE IN DB		
		CurrentScore, exists := handler.currentScore(score.Name)
		
		switch {
			case !exists: //store new value
				err = handler.storeScore(score)
				if err != nil {
					log.Printf("Failed storing new score in StoreScore: %s\n", err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
			case exists && score.Score > CurrentScore: //update existing value
				err = handler.updateScore(score)
				if err != nil {
					log.Printf("Failed updating score in StoreScore: %s\n", err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
			case exists && score.Score <= CurrentScore: //don't do anything
				log.Printf("Score was less or the same as the one stored in database: %#v\n", score)
				return c.String(http.StatusOK, "Score was less or the same as the one stored in database")
		}
		
		log.Printf("Score has been stored: %#v\n", score)
		return c.String(http.StatusCreated, "Score has been stored")
	})
}

// GET SCORE //
func GetScore(g *echo.Group, handler SqlHandler){
	g.GET("/get", func(c echo.Context) error {
		var (
			page, mode int
			err error
			)
		
		//READING PARAMS
		pageStr := c.QueryParam("page")
		if pageStr == "" {
			page = 1
		} else {
			page, _ = strconv.Atoi( pageStr )
			if page < 1 {
				log.Println("Invalid page parameter in request (GetScore)")
				return c.String(http.StatusInternalServerError, "Invalid page parameter in request")
			}
		}
		
		modeStr := c.QueryParam("mode")
		if modeStr == "" {
			mode = ModeDefault
		} else {
			mode, err = strconv.Atoi( modeStr )
			if err != nil || mode < 0 || mode > 1 {
				log.Println("Invalid mode parameter in request (GetScore)")
				return c.String(http.StatusInternalServerError, "Invalid mode parameter in request")
			}
		}
		name := c.QueryParam("name")
		
		//RETURNING RESULTS
			// Getting results from the given page
		rScores, err := handler.getPage(page, mode)
		if err != nil {
			log.Printf("Failed reading from DB in GetScore: %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
			// Calculating if there are any results on the next page
		var valueCount, nextPage int64
		
		switch mode {
			case 0:
				handler.db.Table("players").Count(&valueCount)
			case 1:
				handler.db.Table("players").Where("submitted_at >= ?", now.BeginningOfMonth()).Count(&valueCount)
		}
		
		nextPage = int64(page+1)
		if nextPage * PageLength - valueCount >= PageLength {
			nextPage = 0
		}
		
		if name == "" { // IF THE NAME ISN'T PASSED THAT'S ALL
			return c.JSON(http.StatusOK, JSON_Response{Results: rScores, NextPage: int(nextPage)})
		}
		
		//IF THE NAME IS PASSED
		PlayersAroundMe, PlayerExists := handler.aroundMe(page, name, mode)
		
		if PlayerExists {
			return c.JSON(http.StatusOK, JSON_Response_With_Name{Results: rScores, AroundMe: PlayersAroundMe, NextPage: int(nextPage)})
		} else {
			return c.JSON(http.StatusOK, JSON_Response{Results: rScores, NextPage: int(nextPage)})
		}
	})
}

//////////////////////////// MIDDLEWARES ////////////////////////////

func CheckToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := strings.Split(c.Request().Header.Get("Authorization"), " ")
		
		if len(authHeader) == 2 {
			if authHeader[0] == "Bearer" && authHeader[1] == AuthToken {
				return next(c)
			}
		}
		log.Println("Invalid token in authorisation header in CheckToken")
		return c.String(http.StatusUnauthorized, "Invalid token")
	}
}

func main() {
	handler := OpenDB()
	
	e := echo.New()
	
	lbGroup := e.Group("/leaderboard")
	
	lbGroup.Use(CheckToken)
	
	StoreScore(lbGroup, handler)
	GetScore(lbGroup, handler)
	
	e.Logger.Fatal(e.Start(":1234"))
}
