/**
 * DocSidebar - Left sidebar for the Doc Layout.
 * Contains the category navigation tree.
 * Fixed position, w-64, independent scrolling.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Spinner } from '@/components/ui/spinner';
import { useDocCategoryTree } from '@/hooks/useDocNav';
import DocNavTree from './DocNavTree';

const DocSidebar: React.FC = () => {
  const { t } = useTranslation();
  const { data: tree, isLoading, error } = useDocCategoryTree();

  return (
    <aside className="hidden md:block w-64 fixed top-12 bottom-0 left-0 overflow-y-auto border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
      <div className="py-4">
        <div className="px-4 mb-3">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {t('doc.browseByCategory')}
          </h2>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Spinner className="h-5 w-5" />
          </div>
        ) : error ? (
          <div className="px-4 py-4 text-sm text-destructive">
            Failed to load categories
          </div>
        ) : (
          <DocNavTree tree={tree ?? []} />
        )}
      </div>
    </aside>
  );
};

export default DocSidebar;
