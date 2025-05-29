package controller

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/request"
	"live_replay_project/backend/internal/service"
)

type UserController struct {
	service service.UserService
}

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

func (uc *UserController) Register(ctx *gin.Context) {
	userRegisterDTO := request.UserRegisterDTO{}
	err := ctx.Bind(&userRegisterDTO)
	if err != nil {
		global.Logger.Debug("UserController: Register Binding Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	err = uc.service.Register(ctx, userRegisterDTO)
	if err != nil {
		global.Logger.Debug("UserController: Register Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, "")
}

func (uc *UserController) LoginByPassword(ctx *gin.Context) {
	userLoginDTO := request.UserLoginDTO{}
	err := ctx.Bind(&userLoginDTO)
	if err != nil {
		global.Logger.Debug("UserController: Login Binding Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	login, err := uc.service.Login(ctx, userLoginDTO)
	if err != nil {
		global.Logger.Debug("UserController: Login Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, login)
}

func (uc *UserController) GetCode(ctx *gin.Context) {
	userGetCodeDTO := request.UserGetCodeDTO{}
	err := ctx.Bind(&userGetCodeDTO)
	if err != nil {
		global.Logger.Debug("UserController: GetCode Binding Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	code, err := uc.service.GetCode(ctx, userGetCodeDTO.Phone)
	if err != nil {
		global.Logger.Debug("UserController: GetCode Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, code)
}

func (uc *UserController) LoginByCode(ctx *gin.Context) {
	dto := request.UserLoginCodeDTO{}
	err := ctx.Bind(&dto)
	if err != nil {
		global.Logger.Debug("UserController: LoginCode Binding Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	login, err := uc.service.VerifyCode(ctx, dto.Phone, dto.Code)
	if err != nil {
		global.Logger.Debug("UserController: LoginCode Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, login)
}

func (uc *UserController) GetCaptcha(ctx *gin.Context) {
	getCaptchaVO, err := uc.service.GetCaptcha(ctx)
	if err != nil {
		global.Logger.Debug("UserController: GetCaptcha Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, getCaptchaVO)
}

func (uc *UserController) VerifyCaptcha(ctx *gin.Context) {
	verifyCaptchaDTO := request.UserVerifyCaptchaDTO{}
	err := ctx.Bind(&verifyCaptchaDTO)
	if err != nil {
		global.Logger.Debug("UserController: VerifyCaptcha Binding Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	err = uc.service.VerifyCaptcha(ctx, verifyCaptchaDTO)
	if err != nil {
		global.Logger.Debug("UserController: VerifyCaptcha Failed")
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, "")
}
