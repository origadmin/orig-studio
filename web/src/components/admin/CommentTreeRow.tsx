/**
 * CommentTreeRow molecule - renders a single row in the comment tree table
 * with proper indentation, expand toggle, status badge, and action menu.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { ChevronRight, ChevronDown, ThumbsUp, MessageCircle, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { CommentTreeNode } from '@/lib/utils/commentTree';
import { CommentContentCell } from './CommentContentCell';
import { CommentActionMenu } from './CommentActionMenu';

interface CommentTreeRowProps {
  node: CommentTreeNode;
  isExpanded: boolean;
  onToggleExpand: (id: string) => void;
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
  onBlock: (id: string) => void;
  onUnblock: (id: string) => void;
  onDelete: (id: string) => void;
  onViewReports: (id: string) => void;
  onDismissReports: (id: string) => void;
}

const STATUS_BADGE_VARIANTS: Record<string, 'default' | 'secondary' | 'outline' | 'destructive'> = {
  approved: 'default',
  pending: 'outline',
  rejected: 'secondary',
  blocked: 'destructive',
};

const STATUS_LABELS: Record<string, string> = {
  approved: 'Approved',
  pending: 'Pending',
  rejected: 'Rejected',
  blocked: 'Blocked',
};

export const CommentTreeRow: React.FC<CommentTreeRowProps> = React.memo(
  ({
    node,
    isExpanded,
    onToggleExpand,
    onApprove,
    onReject,
    onBlock,
    onUnblock,
    onDelete,
    onViewReports,
    onDismissReports,
  }) => {
    const { t } = useTranslation();
    const status = node.status || 'pending';

    return (
      <TableRow
        className={cn(status === 'blocked' && 'opacity-70')}
        aria-level={node.depth + 1}
        aria-expanded={node.hasReplies ? isExpanded : undefined}
        role="treeitem"
      >
        {/* Content column with indent + expand toggle */}
        <TableCell>
          <div
            className="flex items-center gap-1.5"
            style={{ paddingLeft: `${node.depth * 24 + 12}px` }}
          >
            {/* Expand toggle */}
            {node.hasReplies ? (
              <button
                onClick={() => onToggleExpand(node.id)}
                className="flex-shrink-0 p-0.5 rounded hover:bg-muted transition-colors"
                aria-label={isExpanded ? 'Collapse' : 'Expand'}
              >
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                )}
              </button>
            ) : (
              <span className="w-5 flex-shrink-0" />
            )}
            <CommentContentCell content={node.content || ''} status={status} />
          </div>
        </TableCell>

        {/* User */}
        <TableCell className="hidden md:table-cell">
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <AvatarImage src={node.avatar || undefined} />
              <AvatarFallback className="text-xs">
                {(node.username || 'U')[0]?.toUpperCase() || 'U'}
              </AvatarFallback>
            </Avatar>
            <span className="text-sm font-medium truncate max-w-[100px]">
              {node.username || 'Unknown'}
            </span>
          </div>
        </TableCell>

        {/* Media */}
        <TableCell className="hidden lg:table-cell">
          <span className="text-sm truncate max-w-[130px] block">
            {node.media?.title || '-'}
          </span>
        </TableCell>

        {/* Likes */}
        <TableCell className="hidden md:table-cell text-center">
          <div className="flex items-center justify-center gap-1">
            <ThumbsUp className="h-3 w-3 text-muted-foreground" />
            {node.like_count ?? 0}
          </div>
        </TableCell>

        {/* Replies */}
        <TableCell className="hidden md:table-cell text-center">
          <div className="flex items-center justify-center gap-1">
            <MessageCircle className="h-3 w-3 text-muted-foreground" />
            {node.reply_count ?? 0}
          </div>
        </TableCell>

        {/* Status */}
        <TableCell>
          <Badge variant={STATUS_BADGE_VARIANTS[status] || 'outline'}>
            {STATUS_LABELS[status] || status}
          </Badge>
        </TableCell>

        {/* Reports */}
        <TableCell className="text-center">
          <div className="flex items-center justify-center gap-1">
            {(node.report_count ?? 0) > 0 && node.has_pending_reports && (
              <AlertTriangle className="h-3 w-3 text-orange-500" />
            )}
            <span className={cn('text-sm', node.has_pending_reports && 'text-orange-600 font-medium')}>
              {node.report_count ?? 0}
            </span>
          </div>
        </TableCell>

        {/* Actions */}
        <TableCell className="text-right">
          <CommentActionMenu
            node={node}
            onApprove={onApprove}
            onReject={onReject}
            onBlock={onBlock}
            onUnblock={onUnblock}
            onDelete={onDelete}
            onViewReports={onViewReports}
            onDismissReports={onDismissReports}
          />
        </TableCell>
      </TableRow>
    );
  }
);

CommentTreeRow.displayName = 'CommentTreeRow';
