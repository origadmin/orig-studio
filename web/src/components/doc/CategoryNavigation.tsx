/**
 * CategoryNavigation - Full category navigation cards for the Doc Home page.
 * Shows all categories with article counts in a grid.
 */
import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { FileText, Folder, FolderOpen } from 'lucide-react';
import { useDocCategoryTree } from '@/hooks/useDocNav';
import { Spinner } from '@/components/ui/spinner';

const CategoryNavigation: React.FC = () => {
  const { t } = useTranslation();
  const { data: tree, isLoading } = useDocCategoryTree();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Spinner className="h-5 w-5" />
      </div>
    );
  }

  const categories = tree ?? [];

  if (categories.length === 0) {
    return null;
  }

  return (
    <section className="mb-12">
      <h2 className="text-xl font-semibold mb-4">{t('doc.browseByCategory')}</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {categories.map((cat) => (
          <Link
            key={cat.id}
            to="/categories/$slug"
            params={{ slug: cat.slug }}
            className="flex items-start gap-3 p-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-primary/50 dark:hover:border-primary/50 hover:bg-primary/5 transition-colors"
          >
            <FolderOpen size={20} className="text-primary shrink-0 mt-0.5" />
            <div className="min-w-0">
              <div className="font-medium text-sm truncate">{cat.name}</div>
              {cat.description && (
                <p className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
                  {cat.description}
                </p>
              )}
              <div className="flex items-center gap-2 mt-1.5 text-xs text-muted-foreground">
                {cat.media_count !== undefined && (
                  <span className="flex items-center gap-1">
                    <FileText size={11} />
                    {cat.media_count} articles
                  </span>
                )}
                {cat.hasChildren && (
                  <span className="flex items-center gap-1">
                    <Folder size={11} />
                    {cat.descendantCount} subcategories
                  </span>
                )}
              </div>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
};

export default CategoryNavigation;
