package config

import "MediaWarp/constants"

// 程序版本信息
type VersionInfo struct {
	AppVersion string // 程序版本号
	CommitHash string // GIt Commit Hash
	BuildData  string // 编译时间
	GoVersion  string // 编译 Golang 版本
	OS         string // 操作系统
	Arch       string //  架构
}

// 上游媒体服务器相关设置
type MediaServerSetting struct {
	Type constants.MediaServerType // 媒体服务器类型
	ADDR string                    // 地址
	AUTH string                    // 认证授权KEY
}

// 日志设置
type LoggerSetting struct {
	AccessLogger  BaseLoggerSetting // 访问日志相关配置
	ServiceLogger BaseLoggerSetting // 服务日志相关配置
}

// 基础日志配置字段
type BaseLoggerSetting struct {
	Console bool // 是否将日志输出到终端中
	File    bool // 是否将日志输出到文件中
}

// Web前端自定义设置
type WebSetting struct {
	Enable            bool   // 启用自定义前端设置
	Custom            bool   // 启用用户自定义静态资源
	Index             bool   // 是否从 custom 目录读取 index.html 文件作为首页
	Head              string // 添加到 index.html 的 HEAD 中
	ExternalPlayerUrl bool   // 是否开启外置播放器
	Crx               bool   // crx 美化
	ActorPlus         bool   // 过滤没有头像的演员和制作人员
	FanartShow        bool   // 显示同人图（fanart图）
	Danmaku           bool   // Web 弹幕
	VideoTogether     bool   // VideoTogether
}

// 客户端User-Agent过滤设置
type ClientFilterSetting struct {
	Enable     bool
	Mode       constants.FliterMode
	ClientList []string
}

// HTTPStrm播放设置
type HTTPStrmSetting struct {
	Enable     bool
	TransCode  bool // false->强制关闭转码 true->保持原有转码设置
	FinalURL   bool // 对 URL 进行重定向判断，找到非重定向地址再重定向给客户端，减少客户端重定向次数
	PrefixList []string
}

// AlistStrm具体设置
type AlistSetting struct {
	ADDR       string
	Username   string
	Password   string
	Token      *string
	PrefixList []string
}

// AlistStrm播放设置
type AlistStrmSetting struct {
	Enable    bool
	TransCode bool // false->强制关闭转码 true->保持原有转码设置
	RawURL    bool // 是否使用原始 URL
	List      []AlistSetting
}

// 字幕设置
type SubtitleSetting struct {
	Enable   bool
	SRT2ASS  bool // SRT 字幕转 ASS 字幕
	ASSStyle []string
	SubSet   bool // ASS 字幕字体子集化
}

// Strm302重定向设置
type Strm302Setting struct {
	Enable           bool     // 启用302重定向功能
	MediaMountPath   []string // 媒体挂载路径列表
	TranscodeEnable  bool     // 是否允许转码
	FallbackOriginal bool     // 失败时是否回退到原始链接
}

// Alist配置
type AlistConfig struct {
	Addr            string            // Alist服务器地址
	Token           string            // Alist访问令牌
	SignEnable      bool              // 是否启用签名
	SignExpireTime  int               // 签名过期时间(秒)
	PublicAddr      string            // Alist公网地址
	RawUrlMapping   map[string]string // 原始URL映射
}

// 路径映射规则
type PathMappingRule struct {
	From string // 源路径
	To   string // 目标路径
}

// 软链接规则
type SymlinkRule struct {
	Path   string // 软链接路径
	Target string // 目标路径
}

// 重定向配置
type RedirectConfig struct {
	Enable            bool               // 启用重定向
	CheckEnable       bool               // 启用重定向检查
	MediaPathMapping  []PathMappingRule  // 媒体路径映射规则
	SymlinkRules      []SymlinkRule      // 软链接规则
}
