import {memo, useMemo} from 'react';
import {ArrowLeft, Save, Play, MoreHorizontal, Trash2, CheckCircle, XCircle, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Separator} from '@/components/ui/separator';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {StatusDot, type StatusDotStatus} from '@/components/common/StatusDot';
import {useMediaQuery} from '@/hooks/useMediaQuery';
import {cn} from '@/lib/utils';

export type SaveState = 'idle' | 'saving' | 'success' | 'error';

export interface HeaderBadgeConfig {
  type: 'media-type' | 'state' | 'featured' | 'custom';
  variant: 'default' | 'secondary' | 'destructive' | 'outline';
  label: string;
  ariaLabel: string;
  className?: string;
  visible?: boolean;
}

export interface EncodingStatusConfig {
  status: StatusDotStatus;
  label?: string;
  ariaLabel?: string;
}

export interface EditPageHeaderProps {
  title: string;
  isDirty: boolean;
  isSaving: boolean;
  saveState: SaveState;
  onBack: () => void;
  onSave: () => void;
  onPreview?: () => void;
  onDelete: () => void;
  badges: HeaderBadgeConfig[];
  encodingStatus?: EncodingStatusConfig;
}

const BADGE_PRIORITY: Record<HeaderBadgeConfig['type'], number> = {
  'state': 0,
  'media-type': 1,
  'featured': 3,
  'custom': 2,
};

const BackNavigation = memo(function BackNavigation({onBack}: { onBack: () => void }) {
  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        onClick={onBack}
        aria-label="返回媒体列表"
      >
        <ArrowLeft className="w-4 h-4"/>
        <span className="hidden sm:inline">返回</span>
      </Button>
      <Separator orientation="vertical" className="h-6"/>
    </>
  );
});

function DirtyIndicator() {
  return (
    <span className="text-primary" aria-label="有未保存的更改">*</span>
  );
}

function BadgeOverflow({count, items}: { count: number; items: HeaderBadgeConfig[] }) {
  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge
            variant="outline"
            className="text-xs cursor-default"
            aria-label={`${count} more badges`}
          >
            +{count}
          </Badge>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <div className="flex flex-col gap-1">
            {items.map((badge) => (
              <span key={badge.type} className="text-xs">
                {badge.label}
              </span>
            ))}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

const TitleWithBadges = memo(function TitleWithBadges({
  title,
  isDirty,
  badges,
  encodingStatus,
  maxBadges,
  showEncodingStatus,
}: {
  title: string;
  isDirty: boolean;
  badges: HeaderBadgeConfig[];
  encodingStatus?: EncodingStatusConfig;
  maxBadges: number;
  showEncodingStatus: boolean;
}) {
  const sortedBadges = useMemo(() =>
    badges
      .filter(b => b.visible !== false)
      .sort((a, b) => BADGE_PRIORITY[a.type] - BADGE_PRIORITY[b.type]),
    [badges]
  );

  const visibleBadges = maxBadges < sortedBadges.length
    ? sortedBadges.slice(0, maxBadges)
    : sortedBadges;
  const overflowBadges = maxBadges < sortedBadges.length
    ? sortedBadges.slice(maxBadges)
    : [];
  const overflowCount = overflowBadges.length;

  return (
    <div className="flex items-center gap-2 min-w-0">
      <h1 className="text-base font-semibold truncate">
        {title || '未命名媒体'}
        {isDirty && <DirtyIndicator/>}
      </h1>
      {(visibleBadges.length > 0 || (showEncodingStatus && encodingStatus) || overflowCount > 0) && (
        <div className="flex items-center gap-1.5 shrink-0">
          {visibleBadges.map((badge) => (
            <Badge
              key={badge.type}
              variant={badge.variant}
              className={cn('text-xs', badge.className)}
              aria-label={badge.ariaLabel}
            >
              {badge.label}
            </Badge>
          ))}
          {showEncodingStatus && encodingStatus && (
            <StatusDot
              status={encodingStatus.status}
              label={encodingStatus.label}
            />
          )}
          {overflowCount > 0 && (
            <BadgeOverflow count={overflowCount} items={overflowBadges}/>
          )}
        </div>
      )}
    </div>
  );
});

function SaveButtonIcon({saveState}: { saveState: SaveState }) {
  switch (saveState) {
    case 'saving':
      return <Loader2 className="w-4 h-4 animate-spin"/>;
    case 'success':
      return <CheckCircle className="w-4 h-4 text-success"/>;
    case 'error':
      return <XCircle className="w-4 h-4 text-destructive"/>;
    default:
      return <Save className="w-4 h-4"/>;
  }
}

function getSaveButtonText(saveState: SaveState): string {
  switch (saveState) {
    case 'saving':
      return '保存中...';
    case 'success':
      return '已保存';
    case 'error':
      return '保存失败';
    default:
      return '保存';
  }
}

const HeaderActions = memo(function HeaderActions({
  saveState,
  isDirty,
  onSave,
  onPreview,
  onDelete,
  hasPreview,
}: {
  saveState: SaveState;
  isDirty: boolean;
  onSave: () => void;
  onPreview?: () => void;
  onDelete: () => void;
  hasPreview: boolean;
}) {
  const isSaving = saveState === 'saving';
  const saveDisabled = isSaving;

  return (
    <div className="flex items-center gap-2 shrink-0">
      <TooltipProvider delayDuration={300}>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              onClick={onSave}
              disabled={saveDisabled}
              className={cn(isDirty && saveState === 'idle' && 'ring-2 ring-primary/30')}
              aria-label="保存 (Ctrl+S)"
            >
              <SaveButtonIcon saveState={saveState}/>
              <span className="hidden md:inline">{getSaveButtonText(saveState)}</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Ctrl+S</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      {hasPreview && onPreview && (
        <Button
          variant="outline"
          size="sm"
          onClick={onPreview}
          className="hidden sm:inline-flex"
          aria-label="预览 (新窗口)"
        >
          <Play className="w-4 h-4"/>
          <span className="hidden md:inline">预览</span>
        </Button>
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" aria-label="更多操作">
            <MoreHorizontal className="w-4 h-4"/>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {hasPreview && onPreview && (
            <>
              <DropdownMenuItem onClick={onPreview} className="sm:hidden">
                <Play className="w-4 h-4 mr-2"/>
                预览
              </DropdownMenuItem>
              <DropdownMenuSeparator className="sm:hidden"/>
            </>
          )}
          <DropdownMenuItem
            onClick={onDelete}
            className="text-destructive focus:text-destructive"
            aria-label="删除媒体"
          >
            <Trash2 className="w-4 h-4 mr-2"/>
            删除
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
});

export function EditPageHeader({
  title,
  isDirty,
  isSaving: _isSaving,
  saveState,
  onBack,
  onSave,
  onPreview,
  onDelete,
  badges,
  encodingStatus,
}: EditPageHeaderProps) {
  const isSm = useMediaQuery('(min-width: 640px)');
  const isLg = useMediaQuery('(min-width: 1024px)');

  const maxBadges = useMemo(() => {
    if (isLg) return Infinity;
    if (isSm) return 2;
    return 1;
  }, [isSm, isLg]);

  const showEncodingStatus = isSm;

  return (
    <div
      className="sticky top-0 z-10 border-b bg-card backdrop-blur supports-[backdrop-filter]:bg-card/95 h-12 sm:h-14"
    >
      <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-6 h-full flex items-center justify-between">
        <div className="flex items-center gap-4 min-w-0">
          <BackNavigation onBack={onBack}/>
          <TitleWithBadges
            title={title}
            isDirty={isDirty}
            badges={badges}
            encodingStatus={encodingStatus}
            maxBadges={maxBadges}
            showEncodingStatus={showEncodingStatus}
          />
        </div>
        <HeaderActions
          saveState={saveState}
          isDirty={isDirty}
          onSave={onSave}
          onPreview={onPreview}
          onDelete={onDelete}
          hasPreview={!!onPreview}
        />
      </div>
    </div>
  );
}
