package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/common/utils"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/request"
	"live_replay_project/backend/internal/response"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UserService interface {
	Register(ctx *gin.Context, dto request.UserRegisterDTO) error
	Login(ctx *gin.Context, dto request.UserLoginDTO) (*response.UserLoginVO, error)
	GetCode(ctx *gin.Context, phone string) (string, error)
	VerifyCode(ctx *gin.Context, phone string, code string) (*response.UserLoginVO, error)
	GetCaptcha(ctx *gin.Context) (*response.UserGetCaptchaVO, error)
	VerifyCaptcha(ctx *gin.Context, dto request.UserVerifyCaptchaDTO) error
}

type UserServiceImpl struct {
	repo *dao.UserDao
}

func NewUserService(repo *dao.UserDao) UserService {
	return &UserServiceImpl{repo: repo}
}

func (u *UserServiceImpl) Register(ctx *gin.Context, dto request.UserRegisterDTO) error {
	user, _ := u.repo.GetUserByPhone(ctx, dto.Phone)
	if user != nil {
		return retcode.NewError(http.StatusForbidden, "该手机号已被注册")
	}
	password := utils.MD5V(dto.Password, "", enum.Md5vIteration)
	_, err := u.repo.Insert(ctx, dto.Phone, password)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserServiceImpl) Login(ctx *gin.Context, dto request.UserLoginDTO) (*response.UserLoginVO, error) {
	user, err := u.repo.GetUserByPhone(ctx, dto.Phone)
	if err != nil || user == nil {
		return nil, retcode.NewError(http.StatusForbidden, "该手机号不存在，请先注册")
	}
	password := utils.MD5V(dto.Password, "", enum.Md5vIteration)
	if user.Password != password {
		return nil, retcode.NewError(http.StatusForbidden, "手机号或密码错误")
	}

	jwtConfig := global.Config.Jwt
	token, err := utils.GenerateToken(uint64(user.UserId), jwtConfig.Name, jwtConfig.Secret)
	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "生成JWT失败")
	}
	resp := response.UserLoginVO{
		UserID:   user.UserId,
		Username: user.Username,
		Token:    token,
	}
	return &resp, nil
}

func (u *UserServiceImpl) GetCode(_ *gin.Context, phone string) (string, error) {
	// 生成六位验证码
	code := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000)
	codeStr := strconv.Itoa(code)
	if len(codeStr) < 6 {
		codeStr = strings.Repeat("0", 6-len(codeStr)) + codeStr
	}

	redisConn := global.RedisClient.Get()
	defer redisConn.Close()

	_, err := redisConn.Do("SET", phone, codeStr, "EX", 60)
	if err != nil {
		global.Logger.Error(err.Error())
		return "", retcode.NewError(http.StatusInternalServerError, "Redis设置验证码失败")
	}

	return codeStr, nil
}

func (u *UserServiceImpl) VerifyCode(ctx *gin.Context, phone string, code string) (*response.UserLoginVO, error) {
	// 简单验证
	if len(phone) != 11 || len(code) != 6 {
		return nil, retcode.NewError(http.StatusBadRequest, "验证失败！")
	}

	redisConn := global.RedisClient.Get()
	defer redisConn.Close()

	redisCode, err := redis.String(redisConn.Do("GET", phone))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, retcode.NewError(http.StatusForbidden, "验证码错误！")
		}
		global.Logger.Error(err.Error())
		return nil, retcode.NewError(http.StatusInternalServerError, "Redis获取验证码失败")
	}
	if redisCode != code {
		return nil, retcode.NewError(http.StatusForbidden, "验证码错误！")
	}

	return u.loginOrRegister(ctx, phone)
}

func (u *UserServiceImpl) loginOrRegister(ctx *gin.Context, phone string) (*response.UserLoginVO, error) {
	user, _ := u.repo.GetUserByPhone(ctx, phone)
	var userId uint32
	var username string
	if user == nil {
		password := utils.MD5V("123456", "", enum.Md5vIteration)
		id, err := u.repo.Insert(ctx, phone, password)
		if err != nil {
			global.Logger.Error(err.Error())
			return nil, retcode.NewError(http.StatusInternalServerError, "注册新用户失败")
		}
		userId = id
		username = "用户" + phone
	} else {
		userId = user.UserId
		username = user.Username
	}

	jwtConfig := global.Config.Jwt
	token, err := utils.GenerateToken(uint64(userId), username, jwtConfig.Secret)
	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "生成JWT失败")
	}
	resp := response.UserLoginVO{
		UserID:   userId,
		Username: username,
		Token:    token,
	}
	return &resp, nil
}

func (u *UserServiceImpl) GetCaptcha(_ *gin.Context) (*response.UserGetCaptchaVO, error) {
	captchaId, bgImage, puzzleImage, puzzleY, err := utils.GenerateCaptcha()
	if err != nil {
		global.Logger.Error(err.Error())
		return nil, retcode.NewError(http.StatusInternalServerError, "生成滑块验证码失败")
	}
	return &response.UserGetCaptchaVO{
		CaptchaId:       captchaId,
		BackgroundImage: bgImage,
		PuzzleImage:     puzzleImage,
		PuzzleY:         puzzleY,
	}, nil
}

func (u *UserServiceImpl) VerifyCaptcha(ctx *gin.Context, dto request.UserVerifyCaptchaDTO) error {
	conn := global.RedisClient.Get()
	defer conn.Close()
	correctX, err := redis.Int(conn.Do("GET", "captcha:"+dto.CaptchaId))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return retcode.NewError(http.StatusForbidden, "验证码已过期")
		} else {
			global.Logger.Error(err.Error())
			return retcode.NewError(http.StatusInternalServerError, "出现未知错误，请稍后重试")
		}
	}

	if dto.X >= correctX+enum.Tolerance || dto.X <= correctX-enum.Tolerance {
		return retcode.NewError(http.StatusForbidden, "验证不通过！")
	}
	return nil
}
