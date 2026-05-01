/**
 * CategoryStatsCards molecule - displays statistics cards for the
 * category tree page.
 *
 * Cards: Total Categories, Active Categories, Top-Level Categories, Sub-Categories
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import type { CategoryTreeStats } from '@/hooks/useCategoryTree';

export interface CategoryStatsCardsProps {
  stats: CategoryTreeStats;
}

export const CategoryStatsCards: React.FC<CategoryStatsCardsProps> = React.memo(
  ({ stats }) => {
    const { t } = useTranslation();

    return (
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* Total Categories */}
        <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-info dark:text-blue-400">
              {stats.total}
            </div>
            <p className="text-sm text-muted-foreground">
              {t('admin.totalCategories') || 'Total Categories'}
            </p>
          </CardContent>
          <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10" />
        </Card>

        {/* Active Categories */}
        <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-success dark:text-green-400">
              {stats.active}
            </div>
            <p className="text-sm text-muted-foreground">
              {t('admin.activeCategories') || 'Active Categories'}
            </p>
          </CardContent>
          <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10" />
        </Card>

        {/* Top-Level Categories */}
        <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-cyan-600 dark:text-cyan-400">
              {stats.topLevel}
            </div>
            <p className="text-sm text-muted-foreground">
              {t('admin.topLevelCategories') || 'Top-Level Categories'}
            </p>
          </CardContent>
          <div className="absolute bottom-0 left-0 h-1 bg-cyan-500 w-full opacity-10" />
        </Card>

        {/* Sub-Categories */}
        <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-warning dark:text-amber-400">
              {stats.subCategories}
            </div>
            <p className="text-sm text-muted-foreground">
              {t('admin.subCategories') || 'Sub-Categories'}
            </p>
          </CardContent>
          <div className="absolute bottom-0 left-0 h-1 bg-amber-500 w-full opacity-10" />
        </Card>
      </div>
    );
  }
);

CategoryStatsCards.displayName = 'CategoryStatsCards';
