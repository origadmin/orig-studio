/**
 * QuickStartGrid - Grid of category cards for quick access.
 * Shows up to 8 categories in a 2x4 grid.
 */
import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Folder } from 'lucide-react';
import { useDocCategoryTree } from '@/hooks/useDocNav';
import { Spinner } from '@/components/ui/spinner';

const QuickStartGrid: React.FC = () => {
  const { t } = useTranslation();
  const { data: tree, isLoading } = useDocCategoryTree();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Spinner className="h-5 w-5" />
      </div>
    );
  }

  const categories = (tree ?? []).slice(0, 8);

  if (categories.length === 0) {
    return null;
  }

  return (
    <section className="mb-12">
      <h2 className="text-xl font-semibold mb-4">{t('doc.quickStart')}</h2>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {categories.map((cat) => (
          <Link
            key={cat.id}
            to="/categories/$slug"
            params={{ slug: cat.slug }}
            className="flex items-center gap-3 p-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-primary/50 dark:hover:border-primary/50 hover:bg-primary/5 transition-colors"
          >
            <Folder size={20} className="text-primary shrink-0" />
            <div className="min-w-0">
              <div className="font-medium text-sm truncate">{cat.name}</div>
              {cat.media_count !== undefined && cat.media_count > 0 && (
                <div className="text-xs text-muted-foreground">
                  {cat.media_count} articles
                </div>
              )}
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
};

export default QuickStartGrid;
