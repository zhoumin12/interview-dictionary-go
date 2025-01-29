package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

const (
	userName = "root"
	password = "root"
	ip       = "8.141.80.133"
	port     = "3306"
	dbName   = "test_zm"
)

var db *gorm.DB

type Question struct {
	QuestionID   int    `json:"question_id"`
	QuestionTest string `json:"question_test"`
}

type Answer struct {
	AnswerID   int    `json:"answer_id"`
	AnswerTest string `json:"answer_test"`
}

type Login struct {
	PhoneNumber int    `json:"phone_number"`
	PassWord    string `json:"pass_word"`
}

type Favorite struct {
	PhoneNumber int `json:"phone_number"`
	QuestionID  int `json:"question_id"`
}

type PassWord struct {
	PhoneNumber int    `json:"phone_number"`
	PassWord    string `json:"pass_word"`
}

// 初始化数据库连接
func InitDB() {
	//user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userName, password, ip, port, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
		return
	}
	// 自动迁移模式
	db.AutoMigrate(&Question{})
	db.AutoMigrate(&Login{})
}

func containkeyword(s, keyword string) bool {
	lowers := strings.ToLower(s)
	lowerkeyword := strings.ToLower(keyword)
	return strings.Contains(lowers, lowerkeyword)
}

func SearchQuestion(keyword string) ([]Question, error) {
	var results []Question
	// 使用 GORM 查询数据库
	err := db.Where("question_test LIKE ? ", "%"+keyword+"%").Find(&results).Error
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for keyword:%s", keyword)
	}
	return results, nil
}

func favoriteExits(phonenumber, questionid int) (bool, error) {
	var record Favorite
	var a bool
	result := db.First(&record, "phone_number = ? AND question_id = ?", phonenumber, questionid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			a = false
		} else {
			a = true
		}
	}
	return a, result.Error
}

func addFavorite(phonenumber, questionid int) error {
	return db.Create(&Favorite{PhoneNumber: phonenumber, QuestionID: questionid}).Error
}

func PhonenumberAndPassward(phonenumber int, password string) (*Login, error) {
	var log Login
	// 执行查询，并将结果扫描到log变量中
	err := db.Raw("SELECT phone_number,pass_ward FROM users WHERE phone_number=?", phonenumber).Scan(&log).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, err
	}
	return &log, nil
}

func Init() {
	r := gin.Default()
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "zhouhuiwen"})
	})

	//获取题目接口
	r.GET("/get_question/:question_id", func(c *gin.Context) {
		//获取题目的逻辑
		// 从URL参数中获取问题ID
		idStr := c.Param("question_id")
		questionId, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "Invalid question ID"})
			return
		}
		// 使用GORM查询问题
		var question Question
		result := db.First(&question, questionId)
		if result.Error != nil {
			// 如果找不到问题，返回404错误
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// 处理记录未找到的情况
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "Question not found"})
			} else {
				// 其他数据库错误
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "Database error"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "11"})
	})
	//获取答案接口
	r.GET("/get_answer/:answer_id", func(c *gin.Context) {
		idStr := c.Param("answer_id")
		answerId, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "Invalid Answer ID"})
			return
		}

		var answer Answer
		result := db.First(&answer, answerId)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "Answer not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "Database error"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "11"})
	})
	//登录接口
	r.GET("/login", func(c *gin.Context) {
		var l Login
		if err := c.ShouldBindJSON(&l); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login credentials"})
			return
		}
		u, err := PhonenumberAndPassward(l.PhoneNumber, l.PassWord)
		u = u
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "Login Successful"})
	})
	//注册接口
	r.GET("/register", func(c *gin.Context) {
		var r Login
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login credentials"})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.PassWord), bcrypt.DefaultCost)
		// 检查手机号是否已存在
		var existingUser Login
		db.First(&existingUser, "phone_number = ?", r.PhoneNumber)
		if existingUser.PhoneNumber != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number already exists"})
			return
		}

		// 插入新用户到数据库
		newUser := Login{
			PhoneNumber: r.PhoneNumber,
			PassWord:    string(hashedPassword),
		}
		db.Create(&newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "Registration Successful"})
	})
	//保存用户密码接口
	r.GET("/save_user_password", func(c *gin.Context) {
		var p PassWord
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var pp PassWord
		result := db.First(&pp, "phone_number=?", p.PhoneNumber)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.PassWord), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		// 更新用户密码
		pp.PassWord = string(hashedPassword)
		result = db.Save(&pp)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "Password Saved"})
	})
	//收藏接口
	r.GET("/favorite", func(c *gin.Context) {
		var f Favorite
		if err := c.ShouldBindJSON(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if exitst, err := favoriteExits(f.PhoneNumber, f.QuestionID); exitst {
			err = err
			c.JSON(http.StatusConflict, gin.H{"error": "Question already favorited"})
			return
		}
		if err := addFavorite(f.PhoneNumber, f.QuestionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "data": "Successful"})
	})

	//搜索接口
	r.GET("/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		if keyword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
			return
		}
		// 搜索问题
		question, err := SearchQuestion(keyword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "Database error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": question})
	})
	r.Run(":8080")
}
