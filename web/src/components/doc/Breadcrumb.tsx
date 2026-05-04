/**
 * Breadcrumb - Breadcrumb navigation for article pages.
 * Shows: Home > Category > Article Title
 */
import React from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronRight, Home } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface BreadcrumbItem {
  label: string;
  to?: string;
  params?: Record<string, string>;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

const Breadcrumb: React.FC<BreadcrumbProps> = ({ items }) => {
  const { t } = useTranslation();

  return (
    <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 text-sm text-muted-foreground mb-4">
      <Link to="/" className="flex items-center gap-1 hover:text-foreground transition-colors">
        <Home size={14} />
        <span className="hidden sm:inline">{t('nav.home')}</span>
      </Link>
      {items.map((item, index) => (
        <React.Fragment key={index}>
          <ChevronRight size={14} className="shrink-0" />
          {item.to ? (
            <Link
              to={item.to}
              params={item.params}
              className="hover:text-foreground transition-colors truncate max-w-[200px]"
            >
              {item.label}
            </Link>
          ) : (
            <span className="text-foreground truncate max-w-[200px]">{item.label}</span>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
};

export default Breadcrumb;
