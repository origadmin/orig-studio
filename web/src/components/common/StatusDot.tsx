import {cn} from '@/lib/utils';

/**
 * Status dot status type
 */
export type StatusDotStatus = 'success' | 'processing' | 'pending' | 'failed';

/**
 * StatusDot component props
 * Used for encoding status and similar state indicators
 */
export interface StatusDotProps {
  /** Status value */
  status: StatusDotStatus;
  /** Custom display label (overrides default mapping) */
  label?: string;
  /** Additional CSS class name */
  className?: string;
}

const STATUS_DOT_MAP: Record<StatusDotStatus, {
  color: string;
  animation: string;
  defaultLabel: string;
  ariaLabel: string;
}> = {
  success: {
    color: 'bg-success',
    animation: '',
    defaultLabel: '编码完成',
    ariaLabel: '编码状态: 编码完成',
  },
  processing: {
    color: 'bg-info',
    animation: 'animate-pulse',
    defaultLabel: '编码中',
    ariaLabel: '编码状态: 编码中',
  },
  pending: {
    color: 'bg-warning',
    animation: '',
    defaultLabel: '等待编码',
    ariaLabel: '编码状态: 等待编码',
  },
  failed: {
    color: 'bg-destructive',
    animation: '',
    defaultLabel: '编码失败',
    ariaLabel: '编码状态: 编码失败',
  },
};

/**
 * StatusDot - A lightweight status indicator with a colored dot and label
 * Used for encoding status and similar state display scenarios
 */
export function StatusDot({status, label, className}: StatusDotProps) {
  const config = STATUS_DOT_MAP[status];
  const displayLabel = label || config.defaultLabel;

  return (
    <span
      className={cn('inline-flex items-center gap-1 text-xs', className)}
      role="status"
      aria-label={config.ariaLabel}
    >
      <span
        className={cn(
          'w-1.5 h-1.5 rounded-full shrink-0',
          config.color,
          config.animation,
        )}
        aria-hidden="true"
      />
      <span>{displayLabel}</span>
    </span>
  );
}
