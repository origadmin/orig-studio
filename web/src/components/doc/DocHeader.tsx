/**
 * DocHeader - Documentation-style top navigation bar.
 * Simpler than the YouTube-style Header: no QuickLinks, no Upload/Write buttons.
 * Includes: Logo + Site name | Search | Language switch | Dark mode toggle | User menu
 */
import React, { useState, useRef, useEffect } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { Search, Menu, Sun, Moon, LogIn, User, LogOut, Shield } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '@/hooks/useAuth';
import { useModuleState } from '@/contexts/ModuleConfigContext';
import LanguageSwitcher from '@/components/common/LanguageSwitcher';

interface DocHeaderProps {
  onToggleSidebar?: () => void;
  darkMode?: boolean;
  onToggleDarkMode?: () => void;
}

const DocHeader: React.FC<DocHeaderProps> = ({ onToggleSidebar, darkMode, onToggleDarkMode }) => {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { isAuthenticated, user, logout, isAdmin } = useAuth();
  const { site } = useModuleState();

  // Close user menu on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) navigate({ to: '/search', search: { q: search } });
  };

  return (
    <header className="fixed top-0 left-0 right-0 h-12 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 z-50">
      <div className="h-full flex items-center px-4 gap-3">
        {/* Mobile sidebar toggle */}
        <button
          onClick={onToggleSidebar}
          className="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors shrink-0 md:hidden"
          aria-label={t('doc.toggleSidebar')}
        >
          <Menu size={18} />
        </button>

        {/* Logo + Site name */}
        <Link to="/" className="flex items-center gap-2 shrink-0">
          <img src="/logo.svg" alt="" className="h-7 w-7" />
          <span className="text-base font-bold text-gray-900 dark:text-white hidden sm:inline">
            {site.site_name || 'OrigStudio'}
          </span>
        </Link>

        {/* Search - centered, wider than YouTube style */}
        <form onSubmit={handleSearch} className="flex-1 max-w-lg mx-auto">
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

        {/* Right side controls */}
        <div className="flex items-center gap-1 shrink-0">
          {onToggleDarkMode && (
            <button
              onClick={onToggleDarkMode}
              className="w-9 h-9 flex items-center justify-center text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
              title={darkMode ? t('nav.toggleLight') : t('nav.toggleDark')}
            >
              {darkMode ? <Sun size={16} className="text-amber-500" /> : <Moon size={16} />}
            </button>
          )}
          <LanguageSwitcher buttonClassName="text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800" />

          {isAuthenticated && user ? (
            <div className="relative" ref={userMenuRef}>
              <button
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="w-9 h-9 flex items-center justify-center rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              >
                {user.avatarUrl ? (
                  <img
                    src={user.avatarUrl}
                    alt={user.displayName}
                    loading="lazy"
                    className="w-7 h-7 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-7 h-7 bg-brand/10 dark:bg-brand/20 rounded-full flex items-center justify-center">
                    <User size={14} className="text-brand dark:text-brand" />
                  </div>
                )}
              </button>

              {userMenuOpen && (
                <div className="absolute right-0 top-full mt-1 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg py-1">
                  <div className="px-4 py-2 border-b border-gray-100 dark:border-gray-700">
                    <p className="text-sm font-medium text-gray-900 dark:text-white">
                      {user.displayName || user.username}
                    </p>
                  </div>
                  {isAdmin && (
                    <Link
                      to="/admin"
                      onClick={() => setUserMenuOpen(false)}
                      className="flex items-center gap-2 px-4 py-2 text-sm text-brand dark:text-brand hover:bg-brand/10 dark:hover:bg-brand/20"
                    >
                      <Shield size={14} /> {t('nav.admin')}
                    </Link>
                  )}
                  <button
                    onClick={() => {
                      setUserMenuOpen(false);
                      logout();
                      navigate({ to: '/' });
                    }}
                    className="w-full flex items-center gap-2 px-4 py-2 text-sm text-destructive dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
                  >
                    <LogOut size={14} /> {t('nav.logout')}
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link
              to="/auth/signin"
              className="flex items-center gap-1.5 h-9 px-3 bg-brand text-white text-sm font-medium rounded-full hover:bg-brand/90 transition-colors"
            >
              <LogIn size={14} />
              <span className="hidden sm:inline">{t('nav.login')}</span>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
};

export default DocHeader;
