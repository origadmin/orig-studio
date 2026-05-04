/**
 * RecentlyUpdated - List of recently updated articles for the Doc Home page.
 * Shows the latest 5 articles with title, time, and summary.
 */
import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Clock, FileText } from 'lucide-react';
import { useDocLatestArticles } from '@/hooks/useDocNav';
import { Spinner } from '@/components/ui/spinner';
import { formatDate } from '@/lib/format';

const RecentlyUpdated: React.FC = () => {
  const { t } = useTranslation();
  const { data: articles, isLoading } = useDocLatestArticles(5);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Spinner className="h-5 w-5" />
      </div>
    );
  }

  const items = articles ?? [];

  if (items.length === 0) {
    return (
      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t('doc.recentlyUpdated')}</h2>
        <div className="text-center py-8">
          <FileText className="h-10 w-10 mx-auto mb-3 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{t('doc.noArticles')}</p>
          <p className="text-xs text-muted-foreground mt-1">{t('doc.noArticlesDesc')}</p>
        </div>
      </section>
    );
  }

  return (
    <section className="mb-12">
      <h2 className="text-xl font-semibold mb-4">{t('doc.recentlyUpdated')}</h2>
      <div className="divide-y divide-gray-200 dark:divide-gray-800">
        {items.map((article) => (
          <Link
            key={article.id}
            to="/articles/$slug"
            params={{ slug: article.slug }}
            className="block py-3 hover:bg-muted/50 -mx-2 px-2 rounded-lg transition-colors"
          >
            <h3 className="font-medium text-sm mb-1 line-clamp-1">{article.title}</h3>
            {article.summary && (
              <p className="text-sm text-muted-foreground line-clamp-2 mb-1">{article.summary}</p>
            )}
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Clock size={12} />
              <span>{formatDate(article.update_time || article.create_time)}</span>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
};

export default RecentlyUpdated;
