import {formatViews, formatRelativeTime, formatDuration, formatDateTime} from './format';

describe('formatViews', () => {
    it('should return 0 for 0 views', () => {
        expect(formatViews(0)).toBe('0');
    });

    it('should return the same number for less than 1000 views', () => {
        expect(formatViews(999)).toBe('999');
    });

    it('should return 1.0K for 1000 views', () => {
        expect(formatViews(1000)).toBe('1.0K');
    });

    it('should return 1.5K for 1500 views', () => {
        expect(formatViews(1500)).toBe('1.5K');
    });

    it('should return 10.0K for 10000 views', () => {
        expect(formatViews(10000)).toBe('10.0K');
    });

    it('should return 100.0K for 100000 views', () => {
        expect(formatViews(100000)).toBe('100.0K');
    });

    it('should return 1.0M for 1000000 views', () => {
        expect(formatViews(1000000)).toBe('1.0M');
    });

    it('should return 1.5M for 1500000 views', () => {
        expect(formatViews(1500000)).toBe('1.5M');
    });
});

describe('formatRelativeTime', () => {
    it('should return "Just now" for 0 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000)).toBe('Just now');
    });

    it('should return "1m ago" for 60 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 60)).toBe('1m ago');
    });

    it('should return "5m ago" for 300 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 300)).toBe('5m ago');
    });

    it('should return "1h ago" for 3600 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 3600)).toBe('1h ago');
    });

    it('should return "5h ago" for 18000 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 18000)).toBe('5h ago');
    });

    it('should return "1d ago" for 86400 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 86400)).toBe('1d ago');
    });

    it('should return "5d ago" for 432000 seconds', () => {
        expect(formatRelativeTime(Date.now() / 1000 - 432000)).toBe('5d ago');
    });
});

describe('formatDuration', () => {
    it('should return "0:00" for 0 seconds', () => {
        expect(formatDuration(0)).toBe('0:00');
    });

    it('should return "0:05" for 5 seconds', () => {
        expect(formatDuration(5)).toBe('0:05');
    });

    it('should return "0:59" for 59 seconds', () => {
        expect(formatDuration(59)).toBe('0:59');
    });

    it('should return "1:00" for 60 seconds', () => {
        expect(formatDuration(60)).toBe('1:00');
    });

    it('should return "1:05" for 65 seconds', () => {
        expect(formatDuration(65)).toBe('1:05');
    });

    it('should return "5:00" for 300 seconds', () => {
        expect(formatDuration(300)).toBe('5:00');
    });

    it('should return "5:30" for 330 seconds', () => {
        expect(formatDuration(330)).toBe('5:30');
    });

    it('should return "1:00:00" for 3600 seconds', () => {
        expect(formatDuration(3600)).toBe('1:00:00');
    });
});

describe('formatDateTime', () => {
    it('should return "N/A" for null input', () => {
        expect(formatDateTime(null)).toBe('N/A');
    });

    it('should return "N/A" for undefined input', () => {
        expect(formatDateTime(undefined)).toBe('N/A');
    });

    it('should return "N/A" for empty string', () => {
        expect(formatDateTime('')).toBe('N/A');
    });

    it('should format ISO string with milliseconds to YYYY-MM-DD HH:mm:ss', () => {
        expect(formatDateTime('2026-05-01T12:34:56.789Z')).toBe('2026-05-01 12:34:56');
    });

    it('should format ISO string without milliseconds to YYYY-MM-DD HH:mm:ss', () => {
        expect(formatDateTime('2026-05-01T12:34:56Z')).toBe('2026-05-01 12:34:56');
    });

    it('should format ISO string with timezone offset', () => {
        expect(formatDateTime('2026-05-01T20:34:56+08:00')).toBe('2026-05-01 12:34:56');
    });

    it('should format protobuf Timestamp object {seconds, nanos}', () => {
        // 2026-05-01 12:34:56 UTC = 1777638896 seconds since epoch
        expect(formatDateTime({seconds: 1777638896, nanos: 0})).toBe('2026-05-01 12:34:56');
    });

    it('should format protobuf Timestamp with nanos (ignoring sub-second precision)', () => {
        // 2026-05-01 12:34:56.789 UTC
        expect(formatDateTime({seconds: 1777638896, nanos: 789000000})).toBe('2026-05-01 12:34:56');
    });

    it('should format numeric timestamp (seconds since epoch)', () => {
        // 2026-05-01 12:34:56 UTC = 1777638896
        expect(formatDateTime(1777638896)).toBe('2026-05-01 12:34:56');
    });

    it('should pad single-digit month/day/hour/minute/second', () => {
        expect(formatDateTime('2026-01-05T03:07:09.123Z')).toBe('2026-01-05 03:07:09');
    });

    it('should return "N/A" for invalid date string', () => {
        expect(formatDateTime('not-a-date')).toBe('N/A');
    });

    it('should return "N/A" for invalid object without seconds', () => {
        expect(formatDateTime({} as any)).toBe('N/A');
    });

    it('should handle Date object input', () => {
        const date = new Date('2026-05-01T12:34:56.789Z');
        expect(formatDateTime(date)).toBe('2026-05-01 12:34:56');
    });
});
