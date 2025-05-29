package middle

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/common"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/common/utils"
	"live_replay_project/backend/global"
	"net/http"
)

func VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := http.StatusOK
		token := c.Request.Header.Get(global.Config.Jwt.Name)
		if token == "" {
			token = c.Query("auth")
		}
		payload, err := utils.ParseToken(token, global.Config.Jwt.Secret)
		if err != nil {
			code = http.StatusUnauthorized
			c.JSON(http.StatusUnauthorized, common.Result{Code: code, Msg: "请先登录！"})
			c.Abort()
			return
		}
		c.Set(enum.CurrentId, payload.UserId)
		c.Set(enum.CurrentName, payload.GrantScope)
		c.Next()
	}
}
