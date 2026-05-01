import { Spinner } from '@/components/ui/spinner';
import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Search, RotateCcw } from 'lucide-react';
import { adminCategoryApi, type Category } from '@/lib/api/category';
import type { CategoryTreeNode } from '@/lib/utils/categoryTree';
import { useCategoryTree } from '@/hooks/useCategoryTree';
import { CategoryTreeTable } from '@/components/admin/CategoryTreeTable';
import { CategoryDialog, type CategoryDialogMode } from '@/components/admin/CategoryDialog';
import { CategoryStatsCards } from '@/components/admin/CategoryStatsCards';

const Categories: React.FC = () => {
  const { t } = useTranslation();
  const {
    tree,
    visibleNodes,
    expandedIds,
    loading,
    loadCategories,
    toggleExpand,
    expandAll,
    collapseAll,
    expandNode,
    stats,
  } = useCategoryTree();

  const [searchKeyword, setSearchKeyword] = useState('');

  // Dialog state
  const [dialogMode, setDialogMode] = useState<CategoryDialogMode>('create');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [currentCategory, setCurrentCategory] = useState<CategoryTreeNode | null>(null);

  // Delete dialog state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<CategoryTreeNode | null>(null);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  // ---- Dialog handlers ----

  const openCreateDialog = useCallback(() => {
    setDialogMode('create');
    setCurrentCategory(null);
    setDialogOpen(true);
  }, []);

  const openEditDialog = useCallback((category: CategoryTreeNode) => {
    setDialogMode('edit');
    setCurrentCategory(category);
    setDialogOpen(true);
  }, []);

  const openAddChildDialog = useCallback((category: CategoryTreeNode) => {
    setDialogMode('addChild');
    setCurrentCategory(category);
    setDialogOpen(true);
  }, []);

  const openDeleteDialog = useCallback((category: CategoryTreeNode) => {
    setDeleteTarget(category);
    setDeleteDialogOpen(true);
  }, []);

  // ---- Action handlers ----

  const handleDialogSubmit = useCallback(
    async (data: Partial<Category>) => {
      if (dialogMode === 'create' || dialogMode === 'addChild') {
        await adminCategoryApi.create(data);
        await loadCategories();
        // Auto-expand parent after adding child
        if (dialogMode === 'addChild' && data.parent_id) {
          expandNode(data.parent_id);
        }
      } else if (dialogMode === 'edit' && currentCategory) {
        await adminCategoryApi.update(currentCategory.id, data);
        await loadCategories();
      }
    },
    [dialogMode, currentCategory, loadCategories, expandNode]
  );

  const handleToggleStatus = useCallback(
    async (category: CategoryTreeNode) => {
      const newStatus = category.status === 1 ? 2 : 1;
      try {
        await adminCategoryApi.patch(category.id, { status: newStatus });
        await loadCategories();
      } catch (err) {
        console.error('Failed to toggle category status:', err);
      }
    },
    [loadCategories]
  );

  const handleDelete = useCallback(async () => {
    if (!deleteTarget) return;
    try {
      await adminCategoryApi.delete(deleteTarget.id);
      await loadCategories();
      setDeleteDialogOpen(false);
      setDeleteTarget(null);
    } catch (err) {
      console.error('Failed to delete category:', err);
    }
  }, [deleteTarget, loadCategories]);

  const handleView = useCallback((category: CategoryTreeNode) => {
    window.open(`/categories/${category.slug}`, '_blank');
  }, []);

  const handleSearch = useCallback(() => {
    // Server-side search returns flat list; reload with keyword
    loadCategories({ keyword: searchKeyword });
  }, [searchKeyword, loadCategories]);

  const handleReset = useCallback(() => {
    setSearchKeyword('');
    loadCategories();
  }, [loadCategories]);

  // ---- Delete warning text ----

  const deleteWarningText = deleteTarget
    ? deleteTarget.children.length > 0
      ? t('admin.deleteCategoryWithChildrenWarning', {
          count: deleteTarget.children.length,
        }) ||
        `This category has ${deleteTarget.children.length} sub-categories. They will become top-level categories. Are you sure you want to delete this category? This action cannot be undone.`
      : t('admin.deleteCategoryConfirm') ||
        'Are you sure you want to delete this category? This action cannot be undone.'
    : '';

  return (
    <div className="space-y-4 p-4 md:p-6">
      {/* Toolbar */}
      <Card className="overflow-hidden">
        <CardContent className="p-6">
          <div className="flex flex-col gap-4">
            {/* Page title */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
              <div>
                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">
                  {t('admin.categories')}
                </h2>
                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                  Manage your content categories
                </p>
              </div>
            </div>

            {/* Divider */}
            <div className="border-t border-slate-200 dark:border-slate-800 my-2" />

            {/* Search and filters */}
            <div className="flex flex-col lg:flex-row gap-4">
              <div className="flex-1 min-w-[120px] max-w-[400px]">
                <div className="relative w-full">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder={t('admin.search') || t('admin.categories') + '...'}
                    value={searchKeyword}
                    onChange={e => setSearchKeyword(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && handleSearch()}
                    className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                  />
                </div>
              </div>
              <div className="flex items-center gap-2 ml-auto lg:ml-0">
                <Button variant="outline" size="sm" onClick={handleReset}>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Reset
                </Button>
                <Button variant="default" size="sm" onClick={handleSearch}>
                  <Search className="h-4 w-4 mr-2" />
                  Search
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats cards */}
      <CategoryStatsCards stats={stats} />

      {/* Tree table */}
      <CategoryTreeTable
        nodes={visibleNodes}
        expandedIds={expandedIds}
        onToggleExpand={toggleExpand}
        onExpandAll={expandAll}
        onCollapseAll={collapseAll}
        onCreate={openCreateDialog}
        onEdit={openEditDialog}
        onDelete={openDeleteDialog}
        onToggleStatus={handleToggleStatus}
        onAddChild={openAddChildDialog}
        onView={handleView}
        loading={loading}
      />

      {/* Create/Edit/AddChild Dialog */}
      <CategoryDialog
        mode={dialogMode}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        tree={tree}
        currentCategory={currentCategory}
        onSubmit={handleDialogSubmit}
      />

      {/* Delete Alert Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t('admin.deleteCategory') || 'Delete Category'}
            </AlertDialogTitle>
            <AlertDialogDescription>{deleteWarningText}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>
              {t('admin.cancel') || 'Cancel'}
            </AlertDialogCancel>
            <AlertDialogAction
              className="bg-red-600 hover:bg-red-700"
              onClick={handleDelete}
            >
              {t('admin.delete') || 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default Categories;
