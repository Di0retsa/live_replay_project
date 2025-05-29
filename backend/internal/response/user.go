package response

type UserLoginVO struct {
	UserID   uint32 `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type UserGetCaptchaVO struct {
	CaptchaId       string `json:"captcha_id"`
	BackgroundImage string `json:"background_image"`
	PuzzleImage     string `json:"puzzle_image"`
	PuzzleY         int    `json:"puzzle_y"`
}
