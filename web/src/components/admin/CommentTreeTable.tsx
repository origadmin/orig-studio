/**
 * CommentTreeTable organism - renders the tree table with expand/collapse,
 * indentation, and row actions.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChevronsDown, ChevronsUp } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import type { CommentTreeNode } from '@/lib/utils/commentTree';
import { CommentTreeRow } from './CommentTreeRow';

export interface CommentTreeTableProps {
  nodes: CommentTreeNode[];
  expandedIds: Set<string>;
  onToggleExpand: (id: string) => void;
  onExpandAll: () => void;
  onCollapseAll: () => void;
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
  onBlock: (id: string) => void;
  onUnblock: (id: string) => void;
  onDelete: (id: string) => void;
  onViewReports: (id: string) => void;
  onDismissReports: (id: string) => void;
  loading?: boolean;
}

export const CommentTreeTable: React.FC<CommentTreeTableProps> = React.memo(
  ({
    nodes,
    expandedIds,
    onToggleExpand,
    onExpandAll,
    onCollapseAll,
    onApprove,
    onReject,
    onBlock,
    onUnblock,
    onDelete,
    onViewReports,
    onDismissReports,
    loading,
  }) => {
    const { t } = useTranslation();

    return (
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('admin.commentList') || 'Comment List'}</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" onClick={onExpandAll}>
                <ChevronsDown className="mr-2 h-4 w-4" />
                {t('admin.expandAll') || 'Expand All'}
              </Button>
              <Button size="sm" variant="outline" onClick={onCollapseAll}>
                <ChevronsUp className="mr-2 h-4 w-4" />
                {t('admin.collapseAll') || 'Collapse All'}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="py-12 text-center">
              <Spinner className="mx-auto" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="min-w-[300px]">{t('admin.commentContent') || 'Content'}</TableHead>
                  <TableHead className="hidden md:table-cell">{t('admin.user') || 'User'}</TableHead>
                  <TableHead className="hidden lg:table-cell">{t('admin.belongVideo') || 'Media'}</TableHead>
                  <TableHead className="hidden md:table-cell text-center">{t('admin.likes') || 'Likes'}</TableHead>
                  <TableHead className="hidden md:table-cell text-center">{t('admin.replies') || 'Replies'}</TableHead>
                  <TableHead>{t('admin.status') || 'Status'}</TableHead>
                  <TableHead className="text-center">Reports</TableHead>
                  <TableHead className="text-right">{t('admin.actions') || 'Actions'}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes.length > 0 ? (
                  nodes.map(node => (
                    <CommentTreeRow
                      key={node.id}
                      node={node}
                      isExpanded={expandedIds.has(node.id)}
                      onToggleExpand={onToggleExpand}
                      onApprove={onApprove}
                      onReject={onReject}
                      onBlock={onBlock}
                      onUnblock={onUnblock}
                      onDelete={onDelete}
                      onViewReports={onViewReports}
                      onDismissReports={onDismissReports}
                    />
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8">
                      {t('admin.noComments') || 'No comments found'}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    );
  }
);

CommentTreeTable.displayName = 'CommentTreeTable';
