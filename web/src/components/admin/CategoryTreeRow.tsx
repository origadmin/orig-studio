/**
 * CategoryTreeRow molecule - renders a single row in the tree table
 * with proper indentation, expand toggle, child count badge,
 * cascading status hint, and action dropdown menu.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
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
  Eye,
  Edit,
  Plus,
  ToggleLeft,
  ToggleRight,
  Trash2,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { CategoryTreeNode } from '@/lib/utils/categoryTree';
import { ExpandToggle } from './ExpandToggle';
import { ChildCountBadge } from './ChildCountBadge';
import { CascadingStatusHint } from './CascadingStatusHint';

export interface CategoryTreeRowProps {
  node: CategoryTreeNode;
  isExpanded: boolean;
  onToggleExpand: (id: number) => void;
  onEdit: (category: CategoryTreeNode) => void;
  onDelete: (category: CategoryTreeNode) => void;
  onToggleStatus: (category: CategoryTreeNode) => void;
  onAddChild: (category: CategoryTreeNode) => void;
  onView: (category: CategoryTreeNode) => void;
}

export const CategoryTreeRow: React.FC<CategoryTreeRowProps> = React.memo(
  ({
    node,
    isExpanded,
    onToggleExpand,
    onEdit,
    onDelete,
    onToggleStatus,
    onAddChild,
    onView,
  }) => {
    const { t } = useTranslation();

    return (
      <TableRow
        className={cn(node.isAncestorDisabled && 'opacity-60')}
        aria-level={node.depth + 1}
      >
        {/* Name column with indent + expand toggle + badges */}
        <TableCell>
          <div
            className="flex items-center gap-1.5"
            style={{ paddingLeft: `${node.depth * 24 + 12}px` }}
          >
            <ExpandToggle
              hasChildren={node.hasChildren}
              isExpanded={isExpanded}
              onClick={() => onToggleExpand(node.id)}
              label={node.name}
            />
            <span className="font-medium">{node.name}</span>
            <ChildCountBadge count={node.children.length} />
            <CascadingStatusHint isAncestorDisabled={node.isAncestorDisabled} />
          </div>
        </TableCell>

        {/* Slug */}
        <TableCell>
          <code className="text-xs bg-muted px-2 py-1 rounded">{node.slug}</code>
        </TableCell>

        {/* Description */}
        <TableCell className="text-muted-foreground max-w-[200px] truncate">
          {node.description || '-'}
        </TableCell>

        {/* Media Count */}
        <TableCell className="text-right">{node.media_count || 0}</TableCell>

        {/* Order */}
        <TableCell>{node.order ?? 0}</TableCell>

        {/* Status */}
        <TableCell>
          <Badge variant={node.status === 1 ? 'default' : 'secondary'}>
            {node.status === 1
              ? t('admin.enabled') || 'Enabled'
              : t('admin.disabled') || 'Disabled'}
          </Badge>
        </TableCell>

        {/* Actions */}
        <TableCell className="text-right">
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
              <DropdownMenuItem onClick={() => onView(node)}>
                <Eye className="mr-2 h-4 w-4" />
                {t('admin.view') || 'View'}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onEdit(node)}>
                <Edit className="mr-2 h-4 w-4" />
                {t('admin.edit') || 'Edit'}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => onAddChild(node)}
                disabled={node.status !== 1}
              >
                <Plus className="mr-2 h-4 w-4" />
                {t('admin.addChild') || 'Add Child'}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onToggleStatus(node)}>
                {node.status === 1 ? (
                  <>
                    <ToggleRight className="mr-2 h-4 w-4" />
                    {t('admin.disable') || 'Disable'}
                  </>
                ) : (
                  <>
                    <ToggleLeft className="mr-2 h-4 w-4" />
                    {t('admin.enable') || 'Enable'}
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => onDelete(node)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {t('admin.delete') || 'Delete'}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    );
  }
);

CategoryTreeRow.displayName = 'CategoryTreeRow';
