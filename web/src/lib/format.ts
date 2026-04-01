import i18n from 'i18next';

export function formatDuration(seconds: number): string {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return h > 0
        ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
        : `${m}:${String(s).padStart(2, '0')}`;
}

export function formatViews(count: number): string {
    const lng = i18n.language;
    if (lng === 'zh') {
        if (count >= 10000) return `${(count / 10000).toFixed(1)}万`;
    }
    if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
    return String(count);
}

export function formatDate(dateStr: string): string {
    const lng = i18n.language;
    const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 86400000);
    if (diff === 0) return lng === 'zh' ? '今天' : 'Today';
    if (diff === 1) return lng === 'zh' ? '昨天' : 'Yesterday';
    if (diff < 7) return lng === 'zh' ? `${diff} 天前` : `${diff}d ago`;
    if (diff < 30) return lng === 'zh' ? `${Math.floor(diff / 7)} 周前` : `${Math.floor(diff / 7)}w ago`;
    if (diff < 365) return lng === 'zh' ? `${Math.floor(diff / 30)} 个月前` : `${Math.floor(diff / 30)}mo ago`;
    return lng === 'zh' ? `${Math.floor(diff / 365)} 年前` : `${Math.floor(diff / 365)}y ago`;
}

export function formatFileSize(bytes: number): string {
    if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`;
    if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`;
    if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${bytes} B`;
}

export function formatRelativeTime(dateStr: string): string {
    const lng = i18n.language;
    const diff = Date.now() - new Date(dateStr).getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    if (minutes < 1) return lng === 'zh' ? '刚刚' : 'Just now';
    if (minutes < 60) return lng === 'zh' ? `${minutes} 分钟前` : `${minutes}m ago`;
    if (hours < 24) return lng === 'zh' ? `${hours} 小时前` : `${hours}h ago`;
    if (days < 7) return lng === 'zh' ? `${days} 天前` : `${days}d ago`;
    const locale = lng === 'zh' ? 'zh-CN' : 'en-US';
    return new Date(dateStr).toLocaleDateString(locale, {month: 'long', day: 'numeric'});
}