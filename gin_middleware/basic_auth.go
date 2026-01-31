package gin_middleware

import "github.com/gin-gonic/gin"

// GetDynamicPasswordFunc 定义函数类型：输入用户名，返回该用户的动态密码和是否找到该用户
type GetDynamicPasswordFunc func(username string) (password string, exists bool)

func CustomMultiUserBasicAuth(getPassWord GetDynamicPasswordFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		//  获取请求头中的账号密码
		user, password, hasAuth := c.Request.BasicAuth()

		if hasAuth {
			// 动态获取该用户当前应有的密码
			expectedPassword, exists := getPassWord(user)

			// 校验用户是否存在且密码匹配
			if exists && password == expectedPassword {
				// 将用户名存入上下文，方便后续 Handler 使用
				c.Set(gin.AuthUserKey, user)
				c.Next()
				return
			}
		}

		// 认证失败
		c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
		c.AbortWithStatus(401)
	}
}
