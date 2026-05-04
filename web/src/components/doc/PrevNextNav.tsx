/**
 * PrevNextNav - Previous/Next article navigation at the bottom of article pages.
 * Shows links to the previous and next articles in the same category.
 */
import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PrevNextNavProps {
  prev?: { slug: string; title: string } | null;
  next?: { slug: string; title: string } | null;
}

const PrevNextNav: React.FC<PrevNextNavProps> = ({ prev, next }) => {
  const { t } = useTranslation();

  if (!prev && !next) return null;

  return (
    <nav className="flex items-center justify-between gap-4 pt-8 mt-8 border-t border-gray-200 dark:border-gray-800">
      {prev ? (
        <Link
          to="/articles/$slug"
          params={{ slug: prev.slug }}
          className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors min-w-0 flex-1"
        >
          <ChevronLeft size={16} className="shrink-0" />
          <div className="min-w-0">
            <div className="text-xs text-muted-foreground">{t('doc.previousArticle')}</div>
            <div className="truncate">{prev.title}</div>
          </div>
        </Link>
      ) : (
        <div className="flex-1" />
      )}

      {next ? (
        <Link
          to="/articles/$slug"
          params={{ slug: next.slug }}
          className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors min-w-0 flex-1 justify-end text-right"
        >
          <div className="min-w-0">
            <div className="text-xs text-muted-foreground">{t('doc.nextArticle')}</div>
            <div className="truncate">{next.title}</div>
          </div>
          <ChevronRight size={16} className="shrink-0" />
        </Link>
      ) : (
        <div className="flex-1" />
      )}
    </nav>
  );
};

export default PrevNextNav;
