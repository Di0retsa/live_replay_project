package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/internal/model"
	"net/http"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (ud *UserDao) GetUserByPhone(ctx *gin.Context, phone string) (*model.User, error) {
	var user model.User
	err := ud.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "User Not Found, phone: "+phone)
	}
	return &user, nil
}

func (ud *UserDao) Insert(ctx *gin.Context, phone string, password string) (uint32, error) {
	user := model.User{Username: "用户" + phone, Phone: phone, Password: password}
	err := ud.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return 0, retcode.NewError(http.StatusInternalServerError, "Create User Failed, phone: "+phone)
	}
	return user.UserId, nil
}
