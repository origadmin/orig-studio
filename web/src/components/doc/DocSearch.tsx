/**
 * DocSearch - Documentation-style search component.
 * Used in both DocHeader (inline) and HeroSection (standalone).
 * This is a simple search input that navigates to /search?q=xxx.
 */
import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface DocSearchProps {
  variant?: 'header' | 'hero';
  className?: string;
}

const DocSearch: React.FC<DocSearchProps> = ({ variant = 'header', className = '' }) => {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      navigate({ to: '/search', search: { q: search } });
    }
  };

  if (variant === 'hero') {
    return (
      <form onSubmit={handleSearch} className={`w-full max-w-md mx-auto ${className}`}>
        <div className="relative">
          <Search
            size={18}
            className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground"
          />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('doc.searchDocs')}
            className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl pl-11 pr-4 py-3 text-sm focus:ring-2 focus:ring-brand focus:border-brand transition-all outline-none shadow-sm"
          />
        </div>
      </form>
    );
  }

  return (
    <form onSubmit={handleSearch} className={`flex-1 max-w-lg mx-auto ${className}`}>
      <div className="relative">
        <Search
          size={15}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
        />
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t('doc.searchDocs')}
          className="w-full bg-gray-100 dark:bg-gray-800 border-0 rounded-full pl-9 pr-4 py-1.5 text-sm focus:ring-2 focus:ring-brand focus:bg-white dark:focus:bg-gray-700 transition-all outline-none"
        />
      </div>
    </form>
  );
};

export default DocSearch;
