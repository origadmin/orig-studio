import { type LucideIcon } from 'lucide-react';

export interface NavItem {
  id: string;
  label: string;
  to: string;
  params?: Record<string, string>;
  icon?: LucideIcon;
  module?: 'articles' | 'videos' | 'music';
  badge?: string | number;
  badgeVariant?: 'default' | 'primary' | 'warning' | 'danger';
  disabled?: boolean;
  external?: boolean;
  /** When true, the `to` path contains `__dynamic__` placeholder that must be resolved at render time */
  isDynamic?: boolean;
}

export interface NavSection {
  id: string;
  title: string;
  requiresAuth?: boolean;
  requiresAdmin?: boolean;
  items: NavItem[];
}
