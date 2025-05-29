package request

type UserRegisterDTO struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginDTO struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginCodeDTO struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type UserGetCodeDTO struct {
	Phone string `json:"phone" binding:"required"`
}

type UserVerifyCaptchaDTO struct {
	CaptchaId string `json:"captcha_id" binding:"required"`
	X         int    `json:"x"`
}
