/**
 * CategoryTreeTable organism - renders the tree table with expand/collapse,
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
import { Plus, ChevronsDown, ChevronsUp } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import type { CategoryTreeNode, ExpandState } from '@/lib/utils/categoryTree';
import { CategoryTreeRow } from './CategoryTreeRow';

export interface CategoryTreeTableProps {
  nodes: CategoryTreeNode[];
  expandedIds: ExpandState;
  onToggleExpand: (id: number) => void;
  onExpandAll: () => void;
  onCollapseAll: () => void;
  onCreate: () => void;
  onEdit: (category: CategoryTreeNode) => void;
  onDelete: (category: CategoryTreeNode) => void;
  onToggleStatus: (category: CategoryTreeNode) => void;
  onAddChild: (category: CategoryTreeNode) => void;
  onView: (category: CategoryTreeNode) => void;
  loading?: boolean;
}

export const CategoryTreeTable: React.FC<CategoryTreeTableProps> = React.memo(
  ({
    nodes,
    expandedIds,
    onToggleExpand,
    onExpandAll,
    onCollapseAll,
    onCreate,
    onEdit,
    onDelete,
    onToggleStatus,
    onAddChild,
    onView,
    loading,
  }) => {
    const { t } = useTranslation();

    return (
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('admin.categoryList') || 'Category List'}</CardTitle>
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
              <Button size="sm" onClick={onCreate}>
                <Plus className="mr-2 h-4 w-4" />
                {t('admin.newCategory') || 'New Category'}
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
                  <TableHead>{t('admin.name') || 'Name'}</TableHead>
                  <TableHead>Slug</TableHead>
                  <TableHead>{t('admin.description') || 'Description'}</TableHead>
                  <TableHead className="text-right">
                    {t('admin.mediaCount') || 'Media'}
                  </TableHead>
                  <TableHead>{t('admin.order') || 'Order'}</TableHead>
                  <TableHead>{t('admin.status') || 'Status'}</TableHead>
                  <TableHead className="text-right">
                    {t('admin.actions') || 'Actions'}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes.length > 0 ? (
                  nodes.map(node => (
                    <CategoryTreeRow
                      key={node.id}
                      node={node}
                      isExpanded={expandedIds.has(node.id)}
                      onToggleExpand={onToggleExpand}
                      onEdit={onEdit}
                      onDelete={onDelete}
                      onToggleStatus={onToggleStatus}
                      onAddChild={onAddChild}
                      onView={onView}
                    />
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      {t('admin.noCategoriesFound') || 'No categories found'}
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

CategoryTreeTable.displayName = 'CategoryTreeTable';
