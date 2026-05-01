/**
 * CommentActionMenu molecule - context-aware dropdown menu that shows
 * different actions based on comment status.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  MoreHorizontal,
  Check,
  X,
  ShieldOff,
  ShieldCheck,
  Flag,
  EyeOff,
  Trash2,
  AlertCircle,
} from 'lucide-react';
import type { CommentTreeNode } from '@/lib/utils/commentTree';

interface CommentActionMenuProps {
  node: CommentTreeNode;
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
  onBlock: (id: string) => void;
  onUnblock: (id: string) => void;
  onDelete: (id: string) => void;
  onViewReports: (id: string) => void;
  onDismissReports: (id: string) => void;
}

export const CommentActionMenu: React.FC<CommentActionMenuProps> = React.memo(
  ({
    node,
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
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            title="More Actions"
          >
            <MoreHorizontal className="h-3 w-3" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {/* Status-based actions */}
          {status === 'pending' && (
            <>
              <DropdownMenuItem onClick={() => onApprove(node.id)}>
                <Check className="mr-2 h-4 w-4" />
                {t('admin.approve') || 'Approve'}
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onReject(node.id)}>
                <X className="mr-2 h-4 w-4" />
                {t('admin.reject') || 'Reject'}
              </DropdownMenuItem>
            </>
          )}
          {status === 'approved' && (
            <>
              <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onReject(node.id)}>
                <X className="mr-2 h-4 w-4" />
                {t('admin.reject') || 'Reject'}
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onBlock(node.id)}>
                <ShieldOff className="mr-2 h-4 w-4" />
                Block
              </DropdownMenuItem>
            </>
          )}
          {status === 'rejected' && (
            <DropdownMenuItem onClick={() => onApprove(node.id)}>
              <Check className="mr-2 h-4 w-4" />
              {t('admin.approve') || 'Approve'}
            </DropdownMenuItem>
          )}
          {status === 'blocked' && (
            <DropdownMenuItem onClick={() => onUnblock(node.id)}>
              <ShieldCheck className="mr-2 h-4 w-4" />
              Unblock
            </DropdownMenuItem>
          )}

          {/* Report-related actions */}
          {(node.report_count ?? 0) > 0 && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onViewReports(node.id)}>
                <Flag className="mr-2 h-4 w-4" />
                View Reports ({node.report_count})
              </DropdownMenuItem>
              {node.has_pending_reports && (
                <DropdownMenuItem onClick={() => onDismissReports(node.id)}>
                  <EyeOff className="mr-2 h-4 w-4" />
                  Dismiss Reports
                </DropdownMenuItem>
              )}
            </>
          )}

          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-destructive focus:text-destructive"
            onClick={() => onDelete(node.id)}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            {t('admin.delete') || 'Delete'}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }
);

CommentActionMenu.displayName = 'CommentActionMenu';
