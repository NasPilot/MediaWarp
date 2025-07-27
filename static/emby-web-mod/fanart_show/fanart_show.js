// Emby Web Mod - Fanart Show 模块
// 此文件为占位符，用于解决构建问题

(function() {
    'use strict';
    
    // Fanart Show 模块
    const FanartShow = {
        init: function() {
            console.log('FanartShow 模块已初始化');
        },
        
        // 粉丝艺术展示功能占位符
        showFanart: function(element, options) {
            // 功能占位符
            return {
                display: function() {},
                hide: function() {},
                update: function() {}
            };
        }
    };
    
    // 模块导出
    if (typeof module !== 'undefined' && module.exports) {
        module.exports = FanartShow;
    } else if (typeof window !== 'undefined') {
        window.FanartShow = FanartShow;
    }
    
})();