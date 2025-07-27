// Jellyfin CRX 通用工具函数
// 此文件为占位符，用于解决构建问题

// Jellyfin 通用工具函数
const JellyfinUtils = {
    // 工具函数占位符
    init: function() {
        console.log('JellyfinUtils 已初始化');
    },
    
    // API 请求工具
    apiRequest: function(url, options) {
        // 功能占位符
        return Promise.resolve({});
    },
    
    // DOM 操作工具
    createElement: function(tag, attributes, content) {
        // 功能占位符
        const element = document.createElement(tag);
        return element;
    }
};

// 导出模块
if (typeof module !== 'undefined' && module.exports) {
    module.exports = JellyfinUtils;
} else if (typeof window !== 'undefined') {
    window.JellyfinUtils = JellyfinUtils;
}