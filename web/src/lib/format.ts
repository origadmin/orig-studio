import i18n from 'i18next';

/** Supported date input types across the application */
type DateInput = string | number | Date | { seconds: number; nanos: number } | null | undefined;

/**
 * Parse various date input formats into a Date object.
 * Supports: ISO 8601 strings, unix timestamps (seconds), Date objects,
 * and protobuf Timestamp objects ({seconds, nanos}).
 * Returns null for invalid or empty input.
 */
function parseDateInput(dateInput: DateInput): Date | null {
    if (dateInput === null || dateInput === undefined || dateInput === '') {
        return null;
    }

    let date: Date;

    if (dateInput instanceof Date) {
        date = dateInput;
    } else if (typeof dateInput === 'string') {
        date = new Date(dateInput);
    } else if (typeof dateInput === 'number') {
        // Handle unix timestamp in seconds
        date = new Date(dateInput * 1000);
    } else if (typeof dateInput === 'object' && dateInput.seconds !== undefined) {
        // Handle Protocol Buffers Timestamp format: {seconds, nanos}
        date = new Date(Number(dateInput.seconds) * 1000 + Math.floor(Number(dateInput.nanos) / 1_000_000));
    } else {
        return null;
    }

    if (isNaN(date.getTime())) {
        return null;
    }

    return date;
}

export function formatDuration(seconds: number | null | undefined): string {
    if (seconds === null || seconds === undefined || isNaN(seconds)) {
        return '0:00';
    }

    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return h > 0
        ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
        : `${m}:${String(s).padStart(2, '0')}`;
}

export function formatViews(count: number | undefined): string {
    const lng = i18n.language;
    if (count === undefined || count === null) return '0';
    if (lng === 'zh') {
        if (count >= 10000) return `${(count / 10000).toFixed(1)}万`;
    }
    if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
    return String(count);
}

export function formatDate(dateInput: DateInput): string {
    const lng = i18n.language;
    const date = parseDateInput(dateInput);

    if (!date) {
        return lng === 'zh' ? '未知时间' : 'Unknown time';
    }

    const diff = Math.floor((Date.now() - date.getTime()) / 86400000);
    if (diff === 0) return lng === 'zh' ? '今天' : 'Today';
    if (diff === 1) return lng === 'zh' ? '昨天' : 'Yesterday';
    if (diff < 7) return lng === 'zh' ? `${diff} 天前` : `${diff}d ago`;
    if (diff < 30) return lng === 'zh' ? `${Math.floor(diff / 7)} 周前` : `${Math.floor(diff / 7)}w ago`;
    if (diff < 365) return lng === 'zh' ? `${Math.floor(diff / 30)} 个月前` : `${Math.floor(diff / 30)}mo ago`;
    return lng === 'zh' ? `${Math.floor(diff / 365)} 年前` : `${Math.floor(diff / 365)}y ago`;
}

/**
 * Format a date to YYYY-MM-DD HH:mm:ss (second-precision, UTC).
 * Designed for admin pages that need precise, consistent datetime display.
 * Handles: ISO 8601 strings, unix timestamps (seconds), Date objects,
 * and protobuf Timestamp objects ({seconds, nanos}).
 */
export function formatDateTime(dateInput: DateInput): string {
    const date = parseDateInput(dateInput);

    if (!date) {
        return 'N/A';
    }

    const year = date.getUTCFullYear();
    const month = String(date.getUTCMonth() + 1).padStart(2, '0');
    const day = String(date.getUTCDate()).padStart(2, '0');
    const hours = String(date.getUTCHours()).padStart(2, '0');
    const minutes = String(date.getUTCMinutes()).padStart(2, '0');
    const seconds = String(date.getUTCSeconds()).padStart(2, '0');

    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

export function formatFileSize(bytes: number): string {
    if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`;
    if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`;
    if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${bytes} B`;
}

export function formatRelativeTime(dateInput: DateInput): string {
    const lng = i18n.language;
    const date = parseDateInput(dateInput);

    if (!date) {
        return lng === 'zh' ? '刚刚' : 'Just now';
    }

    const diff = Date.now() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return lng === 'zh' ? '刚刚' : 'Just now';
    if (minutes < 60) return lng === 'zh' ? `${minutes} 分钟前` : `${minutes}m ago`;
    if (hours < 24) return lng === 'zh' ? `${hours} 小时前` : `${hours}h ago`;
    if (days < 7) return lng === 'zh' ? `${days} 天前` : `${days}d ago`;

    const locale = lng === 'zh' ? 'zh-CN' : 'en-US';
    return date.toLocaleDateString(locale, {month: 'long', day: 'numeric'});
}

export function formatNumber(num: number): string {
    const lng = i18n.language;
    if (num === undefined || num === null) return '0';
    if (lng === 'zh') {
        if (num >= 100000000) return `${(num / 100000000).toFixed(1)}亿`;
        if (num >= 10000) return `${(num / 10000).toFixed(1)}万`;
    }
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return String(num);
}