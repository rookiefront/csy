package gin_middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CorsConfig 跨域配置结构体
type CorsConfig struct {
	// AllowOrigins 允许的域名列表，使用逗号分隔，为空则使用请求中的Origin
	AllowOrigins string
	// AllowHeaders 允许的请求头
	AllowHeaders string
	// AllowMethods 允许的HTTP方法
	AllowMethods string
	// ExposeHeaders 暴露给客户端的响应头
	ExposeHeaders string
	// AllowCredentials 是否允许发送Cookie等凭据
	AllowCredentials bool
	// MaxAge 预检请求的缓存时间（秒）
	MaxAge int
}

// DefaultCorsConfig 返回默认的跨域配置
func DefaultCorsConfig() *CorsConfig {
	return &CorsConfig{
		AllowOrigins:     "",
		AllowHeaders:     "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, X-Token, X-User-Id",
		AllowMethods:     "POST, GET, OPTIONS, DELETE, PUT, PATCH, HEAD",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, New-Token, New-Expires-At",
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}
}

// Cors 跨域请求中间件，支持两种调用方式：
// 1. 不带参数: router.Use(Cors())
// 2. 带配置参数: router.Use(Cors(config))
func Cors(config ...*CorsConfig) gin.HandlerFunc {
	var cfg *CorsConfig

	// 处理参数
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultCorsConfig()
	}

	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 设置Access-Control-Allow-Origin
		if cfg.AllowOrigins != "" {
			c.Header("Access-Control-Allow-Origin", cfg.AllowOrigins)
		} else {
			// 如果未指定允许的域名，则使用请求中的Origin
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				// 如果没有Origin头，可以设置为*或保持为空
				c.Header("Access-Control-Allow-Origin", "*")
			}
		}

		// 设置其他CORS头部
		if cfg.AllowHeaders != "" {
			c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)
		}

		if cfg.AllowMethods != "" {
			c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		}

		if cfg.ExposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", cfg.ExposeHeaders)
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(cfg.MaxAge)))
		}

		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 处理请求
		c.Next()
	}
}

// CorsWithOrigins 快速创建指定允许域名的跨域中间件
func CorsWithOrigins(origins string) gin.HandlerFunc {
	config := DefaultCorsConfig()
	config.AllowOrigins = origins
	return Cors(config)
}

// CorsWithCredentials 快速创建允许凭据的跨域中间件
func CorsWithCredentials(allowCredentials bool) gin.HandlerFunc {
	config := DefaultCorsConfig()
	config.AllowCredentials = allowCredentials
	return Cors(config)
}

// CorsAllowAll 快速创建允许所有域名的跨域中间件（默认行为）
func CorsAllowAll() gin.HandlerFunc {
	return Cors(&CorsConfig{
		AllowHeaders: "*",
		AllowOrigins: "*",
	})
}
