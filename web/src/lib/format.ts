export function formatDuration(seconds: number): string {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return h > 0
        ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
        : `${m}:${String(s).padStart(2, '0')}`;
}

export function formatViews(count: number): string {
    if (count >= 10000) return `${(count / 10000).toFixed(1)}万`;
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
    return String(count);
}

export function formatDate(dateStr: string): string {
    const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 86400000);
    if (diff === 0) return '今天';
    if (diff === 1) return '昨天';
    if (diff < 7) return `${diff} 天前`;
    if (diff < 30) return `${Math.floor(diff / 7)} 周前`;
    if (diff < 365) return `${Math.floor(diff / 30)} 个月前`;
    return `${Math.floor(diff / 365)} 年前`;
}
