/**
 * CategoryDialog organism - unified create/edit/add-child dialog
 * with tree-based parent selector.
 *
 * Modes:
 * - create: "New Category" with enabled parent selector
 * - edit: "Edit Category" with parent selector excluding self + descendants
 * - addChild: "Add Child Category" with locked parent selector
 */
import React, { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import type { Category } from '@/lib/api/category';
import type { CategoryTreeNode } from '@/lib/utils/categoryTree';
import { TreeSelect } from './TreeSelect';

export type CategoryDialogMode = 'create' | 'edit' | 'addChild';

export interface CategoryDialogProps {
  mode: CategoryDialogMode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tree: CategoryTreeNode[];
  currentCategory?: CategoryTreeNode | null;
  onSubmit: (data: Partial<Category>) => Promise<void>;
}

export const CategoryDialog: React.FC<CategoryDialogProps> = ({
  mode,
  open,
  onOpenChange,
  tree,
  currentCategory,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const [formData, setFormData] = React.useReducer(
    (state: Partial<Category>, action: Partial<Category>) => ({
      ...state,
      ...action,
    }),
    {}
  );
  const [submitting, setSubmitting] = React.useState(false);

  // Reset form when dialog opens or mode changes
  useEffect(() => {
    if (!open) return;

    if (mode === 'create') {
      setFormData({
        name: '',
        slug: '',
        description: '',
        parent_id: undefined,
        order: 0,
      });
    } else if (mode === 'edit' && currentCategory) {
      setFormData({
        name: currentCategory.name,
        slug: currentCategory.slug,
        description: currentCategory.description || '',
        parent_id: currentCategory.parent_id ?? undefined,
        order: currentCategory.order ?? 0,
      });
    } else if (mode === 'addChild' && currentCategory) {
      setFormData({
        name: '',
        slug: '',
        description: '',
        parent_id: currentCategory.id,
        order: 0,
      });
    }
  }, [open, mode, currentCategory]);

  const handleSubmit = async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      await onSubmit(formData);
      onOpenChange(false);
    } catch (err) {
      console.error('Failed to submit category:', err);
    } finally {
      setSubmitting(false);
    }
  };

  const getTitle = (): string => {
    switch (mode) {
      case 'create':
        return t('admin.newCategory') || 'New Category';
      case 'edit':
        return t('admin.editCategory') || 'Edit Category';
      case 'addChild':
        return t('admin.addChildCategory') || 'Add Child Category';
    }
  };

  const getDescription = (): string => {
    switch (mode) {
      case 'create':
        return t('admin.createNewCategory') || 'Create a new category';
      case 'edit':
        return t('admin.editCategoryDesc') || 'Edit category details';
      case 'addChild':
        return currentCategory
          ? `Create a sub-category under ${currentCategory.name}`
          : t('admin.createNewCategory') || 'Create a new category';
    }
  };

  const getSubmitLabel = (): string => {
    switch (mode) {
      case 'create':
      case 'addChild':
        return t('admin.create') || 'Create';
      case 'edit':
        return t('admin.save') || 'Save';
    }
  };

  const isParentSelectorDisabled = mode === 'addChild';
  const excludeId = mode === 'edit' ? currentCategory?.id : undefined;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{getTitle()}</DialogTitle>
          <DialogDescription>{getDescription()}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div>
            <label className="block text-sm font-medium mb-1">
              {t('admin.name') || 'Name'} *
            </label>
            <Input
              value={formData.name || ''}
              onChange={e => setFormData({ name: e.target.value })}
              placeholder={t('admin.enterCategoryName') || 'Enter category name'}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              {t('admin.slug') || 'Slug'} *
            </label>
            <Input
              value={formData.slug || ''}
              onChange={e => setFormData({ slug: e.target.value })}
              placeholder={t('admin.enterCategorySlug') || 'Enter category slug'}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              {t('admin.description') || 'Description'}
            </label>
            <Textarea
              value={formData.description || ''}
              onChange={e => setFormData({ description: e.target.value })}
              placeholder={
                t('admin.enterCategoryDescription') || 'Enter category description'
              }
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              {t('admin.parent') || 'Parent Category'}
            </label>
            {isParentSelectorDisabled && currentCategory ? (
              <Input
                value={currentCategory.name}
                disabled
                className="bg-muted"
              />
            ) : (
              <TreeSelect
                tree={tree}
                value={formData.parent_id}
                onChange={parentId => setFormData({ parent_id: parentId })}
                excludeId={excludeId}
                placeholder={
                  t('admin.selectParentCategory') || 'Select parent category'
                }
              />
            )}
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              {t('admin.order') || 'Order'}
            </label>
            <Input
              type="number"
              value={formData.order ?? 0}
              onChange={e =>
                setFormData({ order: parseInt(e.target.value) || 0 })
              }
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('admin.cancel') || 'Cancel'}
          </Button>
          <Button onClick={handleSubmit} disabled={submitting}>
            {submitting ? '...' : getSubmitLabel()}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
