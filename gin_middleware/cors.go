package gin_middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// corsWriter 自定义 ResponseWriter，用于拦截并覆盖下游的跨域 Header
type corsWriter struct {
	gin.ResponseWriter
	cfg    *CorsConfig
	origin string
}

// overrideHeaders 核心逻辑：清空代理方产生的冲突头，并设置中间件的配置头
func (w *corsWriter) overrideHeaders() {
	h := w.Header()

	// 1. 删除下游（被代理方）可能设置的跨域 Header
	h.Del("Access-Control-Allow-Origin")
	h.Del("Access-Control-Allow-Headers")
	h.Del("Access-Control-Allow-Methods")
	h.Del("Access-Control-Expose-Headers")
	h.Del("Access-Control-Allow-Credentials")
	h.Del("Access-Control-Max-Age")

	// 2. 重新按照我们的配置设置跨域 Header
	if w.cfg.AllowOrigins != "" {
		h.Set("Access-Control-Allow-Origin", w.cfg.AllowOrigins)
	} else {
		// 如果未指定允许的域名，则使用请求中的Origin
		if w.origin != "" {
			h.Set("Access-Control-Allow-Origin", w.origin)
		} else {
			// 如果没有Origin头，可以设置为*或保持为空
			h.Set("Access-Control-Allow-Origin", "*")
		}
	}

	if w.cfg.AllowHeaders != "" {
		h.Set("Access-Control-Allow-Headers", w.cfg.AllowHeaders)
	}

	if w.cfg.AllowMethods != "" {
		h.Set("Access-Control-Allow-Methods", w.cfg.AllowMethods)
	}

	if w.cfg.ExposeHeaders != "" {
		h.Set("Access-Control-Expose-Headers", w.cfg.ExposeHeaders)
	}

	if w.cfg.AllowCredentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	}

	if w.cfg.MaxAge > 0 {
		// 【已修复】原来使用 string(rune(cfg.MaxAge)) 是错误的，会转换成乱码字符
		h.Set("Access-Control-Max-Age", strconv.Itoa(w.cfg.MaxAge))
	}
}

// WriteHeader 拦截 Header 写入
func (w *corsWriter) WriteHeader(code int) {
	if !w.Written() {
		w.overrideHeaders()
	}
	w.ResponseWriter.WriteHeader(code)
}

// WriteHeaderNow 拦截立即 Header 写入
func (w *corsWriter) WriteHeaderNow() {
	if !w.Written() {
		w.overrideHeaders()
	}
	w.ResponseWriter.WriteHeaderNow()
}

// Write 拦截 Body 写入（隐式触发 Header 写入）
func (w *corsWriter) Write(b []byte) (int, error) {
	if !w.Written() {
		w.overrideHeaders()
	}
	return w.ResponseWriter.Write(b)
}

// WriteString 拦截 String 写入
func (w *corsWriter) WriteString(s string) (int, error) {
	if !w.Written() {
		w.overrideHeaders()
	}
	return w.ResponseWriter.WriteString(s)
}

// Flush 拦截 Flush 操作（对于流式返回和反向代理非常关键）
func (w *corsWriter) Flush() {
	if !w.Written() {
		w.overrideHeaders()
	}
	w.ResponseWriter.Flush()
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

		// 1. 如果是 OPTIONS 请求，直接拦截并返回
		if method == "OPTIONS" {
			// 直接用 corsWriter 设置并立刻结束
			cw := &corsWriter{ResponseWriter: c.Writer, cfg: cfg, origin: origin}
			cw.overrideHeaders()
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 2. 对于正常请求，将 c.Writer 包装为自定义的 corsWriter
		// 以便在反向代理或后续 Handler 写入数据时拦截 Header
		c.Writer = &corsWriter{
			ResponseWriter: c.Writer,
			cfg:            cfg,
			origin:         origin,
		}

		// 处理下游请求
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
