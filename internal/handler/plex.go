package handler

import (
	"MediaWarp/internal/config"
	"MediaWarp/internal/logging"
	"MediaWarp/internal/utils"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PlexServerHandler Plex服务器处理器
type PlexServerHandler struct {
	proxy      *httputil.ReverseProxy
	plexHost   string
	plexToken  string
	routeRules []RegexpRouteRule
}

// NewPlexServerHandler 创建Plex服务器处理器
func NewPlexServerHandler(plexHost, plexToken string) (MediaServerHandler, error) {
	if plexHost == "" {
		return nil, fmt.Errorf("Plex服务器地址不能为空")
	}

	// 解析Plex服务器URL
	target, err := url.Parse(plexHost)
	if err != nil {
		return nil, fmt.Errorf("无效的Plex服务器地址: %v", err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 添加Plex Token认证
		if plexToken != "" {
			q := req.URL.Query()
			if q.Get("X-Plex-Token") == "" {
				q.Set("X-Plex-Token", plexToken)
				req.URL.RawQuery = q.Encode()
			}
			// 同时设置Header
			if req.Header.Get("X-Plex-Token") == "" {
				req.Header.Set("X-Plex-Token", plexToken)
			}
		}

		// 设置Host头
		req.Host = target.Host
		req.Header.Set("Host", target.Host)

		// 记录请求日志
		logging.Debugf("代理Plex请求: %s %s", req.Method, req.URL.String())
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logging.Errorf("Plex代理错误: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf("代理到Plex服务器失败: %v", err)))
	}

	// 自定义响应修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 处理CORS
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		resp.Header.Set("Access-Control-Allow-Headers", "*")

		// 记录响应日志
		logging.Debugf("Plex响应: %d %s", resp.StatusCode, resp.Request.URL.String())
		return nil
	}

	handler := &PlexServerHandler{
		proxy:     proxy,
		plexHost:  plexHost,
		plexToken: plexToken,
	}

	// 定义路由规则
	routeRules := []RegexpRouteRule{
		// 媒体文件直接播放
		{
			Regexp:  regexp.MustCompile(`^/library/parts/\d+/\d+/file`),
			Handler: handler.HandleMediaRedirect,
		},
		// 转码流
		{
			Regexp:  regexp.MustCompile(`^/video/:/transcode/universal/`),
			Handler: handler.HandleTranscodeRedirect,
		},
		// 图片转码
		{
			Regexp:  regexp.MustCompile(`^/photo/:/transcode`),
			Handler: handler.HandlePhotoRedirect,
		},
		// 字幕流
		{
			Regexp:  regexp.MustCompile(`^/library/streams/\d+`),
			Handler: handler.HandleSubtitleRedirect,
		},
	}

	handler.routeRules = routeRules

	return handler, nil
}

// ReverseProxy 反向代理到Plex服务器
func (h *PlexServerHandler) ReverseProxy(w http.ResponseWriter, r *http.Request) {
	// 处理OPTIONS请求（CORS预检）
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 记录访问日志
	start := time.Now()
	logging.AccessLogf("%s %s %s", utils.GetClientIP(r), r.Method, r.URL.String())

	// 执行代理
	h.proxy.ServeHTTP(w, r)

	// 记录处理时间
	duration := time.Since(start)
	logging.Debugf("Plex请求处理完成: %s %s (%v)", r.Method, r.URL.String(), duration)
}

// GetRegexpRouteRules 获取正则路由规则
func (h *PlexServerHandler) GetRegexpRouteRules() []RegexpRouteRule {
	return h.routeRules
}

// HandleMediaRedirect 处理媒体文件重定向
func (h *PlexServerHandler) HandleMediaRedirect(c *gin.Context) {
	// 从URL路径中提取partID和fileID
	path := c.Request.URL.Path
	re := regexp.MustCompile(`^/library/parts/(\d+)/(\d+)/file`)
	matches := re.FindStringSubmatch(path)
	
	if len(matches) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体文件路径"})
		return
	}

	partID := matches[1]
	fileID := matches[2]

	// 构建原始Plex URL
	originalURL := fmt.Sprintf("%s/library/parts/%s/%s/file", h.plexHost, partID, fileID)
	query := c.Request.URL.RawQuery
	
	if query != "" {
		originalURL += "?" + query
		if h.plexToken != "" && !strings.Contains(query, "X-Plex-Token") {
			originalURL += "&X-Plex-Token=" + h.plexToken
		}
	} else if h.plexToken != "" {
		originalURL += "?X-Plex-Token=" + h.plexToken
	}

	// 检查是否需要路径映射
	if config.HTTPStrm.Enable {
		// 这里可以添加路径映射逻辑
		// 暂时直接重定向到原始URL
	}

	// 重定向到原始URL
	c.Redirect(http.StatusFound, originalURL)
	logging.Infof("媒体文件重定向: %s -> %s", c.Request.URL.String(), originalURL)
}

// HandleTranscodeRedirect 处理转码重定向
func (h *PlexServerHandler) HandleTranscodeRedirect(c *gin.Context) {
	// 构建转码URL
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	transcodeURL := fmt.Sprintf("%s%s", h.plexHost, path)
	
	if query != "" {
		transcodeURL += "?" + query
		if h.plexToken != "" && !strings.Contains(query, "X-Plex-Token") {
			transcodeURL += "&X-Plex-Token=" + h.plexToken
		}
	} else if h.plexToken != "" {
		transcodeURL += "?X-Plex-Token=" + h.plexToken
	}

	// 重定向到转码URL
	c.Redirect(http.StatusFound, transcodeURL)
	logging.Infof("转码重定向: %s -> %s", c.Request.URL.String(), transcodeURL)
}

// HandlePhotoRedirect 处理图片重定向
func (h *PlexServerHandler) HandlePhotoRedirect(c *gin.Context) {
	// 构建图片URL
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	photoURL := fmt.Sprintf("%s%s", h.plexHost, path)
	
	if query != "" {
		photoURL += "?" + query
		if h.plexToken != "" && !strings.Contains(query, "X-Plex-Token") {
			photoURL += "&X-Plex-Token=" + h.plexToken
		}
	} else if h.plexToken != "" {
		photoURL += "?X-Plex-Token=" + h.plexToken
	}

	// 重定向到图片URL
	c.Redirect(http.StatusFound, photoURL)
	logging.Infof("图片重定向: %s -> %s", c.Request.URL.String(), photoURL)
}

// HandleSubtitleRedirect 处理字幕重定向
func (h *PlexServerHandler) HandleSubtitleRedirect(c *gin.Context) {
	// 从URL路径中提取streamID
	path := c.Request.URL.Path
	re := regexp.MustCompile(`^/library/streams/(\d+)`)
	matches := re.FindStringSubmatch(path)
	
	if len(matches) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的字幕流路径"})
		return
	}

	streamID := matches[1]
	query := c.Request.URL.RawQuery
	subtitleURL := fmt.Sprintf("%s/library/streams/%s", h.plexHost, streamID)
	
	if query != "" {
		subtitleURL += "?" + query
		if h.plexToken != "" && !strings.Contains(query, "X-Plex-Token") {
			subtitleURL += "&X-Plex-Token=" + h.plexToken
		}
	} else if h.plexToken != "" {
		subtitleURL += "?X-Plex-Token=" + h.plexToken
	}

	// 重定向到字幕URL
	c.Redirect(http.StatusFound, subtitleURL)
	logging.Infof("字幕重定向: %s -> %s", c.Request.URL.String(), subtitleURL)
}