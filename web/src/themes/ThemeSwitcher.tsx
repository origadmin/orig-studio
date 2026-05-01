/**
 * ThemeSwitcher - UI component for theme and color mode selection.
 *
 * Displays theme cards grouped by category with color preview swatches.
 * Integrates with useTheme hook for state management.
 */
import React, { useState } from 'react';
import { useTheme } from './useTheme';
import type { ThemeMeta, ThemeCategory } from './types';
import { Check, Loader2, Sun, Moon, Monitor } from 'lucide-react';

const CATEGORY_LABELS: Record<ThemeCategory, string> = {
  professional: '专业商务',
  social: '活力社交',
  creative: '创意设计',
  minimal: '极简暗色',
  custom: '项目专属',
};

const COLOR_MODE_OPTIONS = [
  { value: 'light' as const, label: '浅色', icon: Sun },
  { value: 'dark' as const, label: '深色', icon: Moon },
  { value: 'system' as const, label: '跟随系统', icon: Monitor },
];

export function ThemeSwitcher() {
  const { themeId, setTheme, colorMode, setColorMode, themes, isLoading } = useTheme();
  const [activeCategory, setActiveCategory] = useState<string>('all');
  const [switchingTo, setSwitchingTo] = useState<string | null>(null);

  // Group themes by category
  const grouped = themes.reduce<Record<string, ThemeMeta[]>>((acc, theme) => {
    const cat = theme.category;
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(theme);
    return acc;
  }, {});

  const categories = Object.keys(grouped) as ThemeCategory[];

  const handleThemeSwitch = async (id: string) => {
    if (id === themeId) return;
    setSwitchingTo(id);
    try {
      await setTheme(id);
    } finally {
      setSwitchingTo(null);
    }
  };

  const displayThemes =
    activeCategory === 'all' ? themes : grouped[activeCategory] || [];

  return (
    <div className="space-y-6">
      {/* Color mode toggle */}
      <div className="flex items-center gap-3">
        <span className="text-sm font-medium">外观模式</span>
        <div className="flex rounded-lg border p-1">
          {COLOR_MODE_OPTIONS.map((option) => {
            const Icon = option.icon;
            return (
              <button
                key={option.value}
                onClick={() => setColorMode(option.value)}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors ${
                  colorMode === option.value
                    ? 'bg-brand text-brand-foreground'
                    : 'hover:bg-muted'
                }`}
              >
                <Icon className="h-3.5 w-3.5" />
                {option.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Category filter */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setActiveCategory('all')}
          className={`px-3 py-1 text-sm rounded-full transition-colors ${
            activeCategory === 'all'
              ? 'bg-brand text-brand-foreground'
              : 'bg-muted hover:bg-muted/80'
          }`}
        >
          全部
        </button>
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setActiveCategory(cat)}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              activeCategory === cat
                ? 'bg-brand text-brand-foreground'
                : 'bg-muted hover:bg-muted/80'
            }`}
          >
            {CATEGORY_LABELS[cat] || cat}
          </button>
        ))}
      </div>

      {/* Theme grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {displayThemes.map((theme) => (
          <ThemeCard
            key={theme.id}
            theme={theme}
            isActive={themeId === theme.id}
            isSwitching={switchingTo === theme.id}
            onSelect={() => handleThemeSwitch(theme.id)}
          />
        ))}
      </div>

      {/* Loading state */}
      {isLoading && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          正在加载主题...
        </div>
      )}
    </div>
  );
}

const PRIMARY_SCALE_LEVELS = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950] as const;

const PALETTE_COLORS = [
  { key: 'accent', label: 'Accent' },
  { key: 'success', label: 'Success' },
  { key: 'warning', label: 'Warning' },
  { key: 'info', label: 'Info' },
  { key: 'destructive', label: 'Error' },
] as const;

function ThemeCard({
  theme,
  isActive,
  isSwitching,
  onSelect,
}: {
  theme: ThemeMeta;
  isActive: boolean;
  isSwitching: boolean;
  onSelect: () => void;
}) {
  const preview = theme.preview as Record<string, string>;
  const hasPrimaryScale = preview['primary-50'] != null;

  return (
    <button
      onClick={onSelect}
      disabled={isSwitching}
      className={`relative flex flex-col rounded-xl border-2 p-3 text-left transition-all hover:shadow-md ${
        isActive ? 'border-brand shadow-sm' : 'border-border hover:border-brand/30'
      } ${isSwitching ? 'opacity-70 cursor-wait' : ''}`}
    >
      {isActive && (
        <div className="absolute top-2 right-2 w-5 h-5 bg-brand rounded-full flex items-center justify-center">
          <Check className="w-3 h-3 text-brand-foreground" />
        </div>
      )}

      {isSwitching && (
        <div className="absolute top-2 right-2 w-5 h-5 bg-muted rounded-full flex items-center justify-center">
          <Loader2 className="w-3 h-3 animate-spin" />
        </div>
      )}

      {theme.thumbnail ? (
        <img
          src={`/themes/${theme.id}/${theme.thumbnail}`}
          alt={theme.name}
          className="w-full h-16 object-cover rounded-lg mb-3"
        />
      ) : (
        <div className="mb-3">
          {hasPrimaryScale ? (
            <div className="flex h-7 rounded-md overflow-hidden mb-2">
              {PRIMARY_SCALE_LEVELS.map((level, i) => (
                <div
                  key={level}
                  className={`flex-1 ${level === 600 ? 'ring-1 ring-inset ring-white/40' : ''} ${i === 0 ? 'rounded-l-md' : ''} ${i === PRIMARY_SCALE_LEVELS.length - 1 ? 'rounded-r-md' : ''}`}
                  style={{ backgroundColor: preview[`primary-${level}`] || '#888' }}
                />
              ))}
            </div>
          ) : (
            <div
              className="h-7 rounded-md mb-2"
              style={{ backgroundColor: preview.primary || '#888' }}
            />
          )}
          <div className="flex gap-1">
            {PALETTE_COLORS.map(({ key }) => (
              <div
                key={key}
                className="flex-1 h-5 rounded-sm"
                style={{ backgroundColor: preview[key] || '#888' }}
              />
            ))}
          </div>
        </div>
      )}

      <span className="text-sm font-medium">{theme.name}</span>
      {theme.description && (
        <span className="text-xs text-muted-foreground line-clamp-1">{theme.description}</span>
      )}
    </button>
  );
}
