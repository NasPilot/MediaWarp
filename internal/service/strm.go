package service

import (
	"MediaWarp/internal/config"
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// StrmService strm文件服务
type StrmService struct {
	config *config.Strm302Setting
}

// NewStrmService 创建strm服务
func NewStrmService(cfg *config.Strm302Setting) *StrmService {
	return &StrmService{
		config: cfg,
	}
}

// IsStrmFile 判断是否为strm文件
func (s *StrmService) IsStrmFile(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".strm")
}

// ReadStrmContent 读取strm文件内容
func (s *StrmService) ReadStrmContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开strm文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			return line, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取strm文件失败: %v", err)
	}

	return "", fmt.Errorf("strm文件为空或无有效内容")
}

// GetDirectLinkFromStrm 从strm文件获取直链
func (s *StrmService) GetDirectLinkFromStrm(strmPath string) (string, error) {
	content, err := s.ReadStrmContent(strmPath)
	if err != nil {
		return "", err
	}

	// 如果内容已经是HTTP链接，直接返回
	if strings.HasPrefix(content, "http://") || strings.HasPrefix(content, "https://") {
		return content, nil
	}

	// 如果是本地路径，尝试通过Alist获取直链
	if filepath.IsAbs(content) {
		return s.getDirectLinkFromAlist(content)
	}

	return "", fmt.Errorf("无效的strm内容: %s", content)
}

// getDirectLinkFromAlist 通过Alist获取文件直链
func (s *StrmService) getDirectLinkFromAlist(localPath string) (string, error) {
	// 应用路径映射规则
	mappedPath := s.applyPathMapping(localPath)
	if mappedPath == "" {
		return "", fmt.Errorf("路径映射失败: %s", localPath)
	}

	// 获取Alist服务器
	alistServer, err := GetAlistServer(config.Alist.Addr)
	if err != nil {
		return "", fmt.Errorf("获取Alist服务器失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := alistServer.FsGet(mappedPath)
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %v", err)
	}

	if fileInfo.IsDir {
		return "", fmt.Errorf("路径指向目录而非文件: %s", mappedPath)
	}

	if fileInfo.RawURL == "" {
		return "", fmt.Errorf("无法获取文件直链")
	}

	// 如果配置了公网地址，替换内网地址
	if config.Alist.PublicAddr != "" {
		parsedURL, err := url.Parse(fileInfo.RawURL)
		if err == nil {
			publicURL, err := url.Parse(config.Alist.PublicAddr)
			if err == nil {
				parsedURL.Scheme = publicURL.Scheme
				parsedURL.Host = publicURL.Host
				return parsedURL.String(), nil
			}
		}
	}

	return fileInfo.RawURL, nil
}

// applyPathMapping 应用路径映射规则
func (s *StrmService) applyPathMapping(localPath string) string {
	for _, rule := range config.Redirect.MediaPathMapping {
		if strings.HasPrefix(localPath, rule.From) {
			return strings.Replace(localPath, rule.From, rule.To, 1)
		}
	}
	return ""
}

// IsInMediaMountPath 判断路径是否在媒体挂载路径中
func (s *StrmService) IsInMediaMountPath(filePath string) bool {
	for _, mountPath := range s.config.MediaMountPath {
		if strings.HasPrefix(filePath, mountPath) {
			return true
		}
	}
	return false
}

// ShouldRedirect 判断是否应该进行302重定向
func (s *StrmService) ShouldRedirect(filePath string, userAgent string) bool {
	if !s.config.Enable {
		return false
	}

	// 检查是否为strm文件
	if !s.IsStrmFile(filePath) {
		return false
	}

	// 检查是否在媒体挂载路径中
	if !s.IsInMediaMountPath(filePath) {
		return false
	}

	// 如果启用了转码，检查是否应该回退到转码
	if s.config.TranscodeEnable && s.shouldFallbackToTranscode(userAgent) {
		return false
	}

	return true
}

// shouldFallbackToTranscode 判断是否应该回退到转码
func (s *StrmService) shouldFallbackToTranscode(userAgent string) bool {
	if !s.config.FallbackOriginal {
		return false
	}

	// 检查用户代理是否匹配回退规则
	// 这里可以根据实际需求添加更复杂的判断逻辑
	return false
}

// HandleRedirect 处理302重定向
func (s *StrmService) HandleRedirect(w http.ResponseWriter, r *http.Request, filePath string) error {
	directLink, err := s.GetDirectLinkFromStrm(filePath)
	if err != nil {
		return fmt.Errorf("获取直链失败: %v", err)
	}

	// 执行302重定向
	http.Redirect(w, r, directLink, http.StatusFound)
	log.Printf("302重定向: %s -> %s", filePath, directLink)
	return nil
}

// CheckHealth 检查strm相关服务的健康状态
func (s *StrmService) CheckHealth() error {
	if !s.config.Enable {
		return nil
	}

	// 检查Alist连接
	if config.Alist.Addr != "" {
		alistServer, err := GetAlistServer(config.Alist.Addr)
		if err != nil {
			return fmt.Errorf("Alist服务器不可用: %v", err)
		}

		// 尝试获取根目录信息来测试连接
		_, err = alistServer.FsGet("/")
		if err != nil {
			return fmt.Errorf("Alist连接测试失败: %v", err)
		}
	}

	return nil
}

// ExtractFilePathFromRequest 从请求中提取文件路径
func ExtractFilePathFromRequest(r *http.Request) string {
	// 从URL路径中提取文件路径
	path := r.URL.Path
	
	// 移除可能的前缀
	prefixes := []string{"/library/parts/", "/video/:/transcode/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break
		}
	}

	// 从查询参数中获取路径信息
	if filePath := r.URL.Query().Get("path"); filePath != "" {
		return filePath
	}

	// 尝试从其他参数中解析路径
	if sessionID := r.URL.Query().Get("session"); sessionID != "" {
		// 这里可以根据session ID查找对应的文件路径
		// 暂时返回空，实际实现时需要根据具体的会话管理逻辑
	}

	return path
}

// IsMediaRequest 判断是否为媒体请求
func IsMediaRequest(r *http.Request) bool {
	// 定义媒体相关的URL模式
	mediaPatterns := []*regexp.Regexp{
		regexp.MustCompile(`/library/parts/\d+`),
		regexp.MustCompile(`/video/:/transcode/`),
		regexp.MustCompile(`/photo/:/transcode/`),
		regexp.MustCompile(`/music/:/transcode/`),
	}

	path := r.URL.Path
	for _, pattern := range mediaPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	return false
}

// IsTranscodeRequest 判断是否为转码请求
func IsTranscodeRequest(r *http.Request) bool {
	return strings.Contains(r.URL.Path, "/transcode/")
}