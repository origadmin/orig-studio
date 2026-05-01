/**
 * Custom hook for managing category tree state, expand/collapse,
 * and derived data (visible nodes, statistics).
 */
import { useState, useMemo, useCallback } from 'react';
import { adminCategoryApi, type Category } from '@/lib/api/category';
import {
  buildCategoryTree,
  flattenCategoryTree,
  type CategoryTreeNode,
  type ExpandState,
} from '@/lib/utils/categoryTree';
import { PAGINATION_CONFIG } from '@/config/pagination';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface CategoryTreeStats {
  total: number;
  active: number;
  topLevel: number;
  subCategories: number;
}

export interface UseCategoryTreeReturn {
  // Raw data
  categories: Category[];
  loading: boolean;
  total: number;

  // Tree data
  tree: CategoryTreeNode[];
  visibleNodes: CategoryTreeNode[];
  expandedIds: ExpandState;

  // Actions
  loadCategories: (params?: { page?: number; page_size?: number; keyword?: string }) => Promise<void>;
  toggleExpand: (id: number) => void;
  expandAll: () => void;
  collapseAll: () => void;
  expandNode: (id: number) => void;

  // Stats
  stats: CategoryTreeStats;
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useCategoryTree(): UseCategoryTreeReturn {
  const [categories, setCategories] = useState<Category[]>([]);
  const [expandedIds, setExpandedIds] = useState<ExpandState>(new Set());
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);

  // Build tree from flat list (memoized)
  const tree = useMemo(
    () => buildCategoryTree(categories),
    [categories]
  );

  // Flatten tree based on expand state (memoized)
  const visibleNodes = useMemo(
    () => flattenCategoryTree(tree, expandedIds),
    [tree, expandedIds]
  );

  // Calculate stats from tree (memoized)
  const stats = useMemo<CategoryTreeStats>(() => {
    // Flatten all nodes (fully expanded) for stats calculation
    const allNodes = flattenCategoryTree(
      tree,
      new Set(tree.map(n => n.id))
    );
    return {
      total: allNodes.length,
      active: allNodes.filter(n => n.status === 1).length,
      topLevel: tree.length,
      subCategories: allNodes.filter(n => n.depth > 0).length,
    };
  }, [tree]);

  const toggleExpand = useCallback((id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const expandNode = useCallback((id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      next.add(id);
      return next;
    });
  }, []);

  const expandAll = useCallback(() => {
    setExpandedIds(
      new Set(
        flattenCategoryTree(
          tree,
          new Set(tree.map(n => n.id))
        )
          .filter(n => n.hasChildren)
          .map(n => n.id)
      )
    );
  }, [tree]);

  const collapseAll = useCallback(() => {
    setExpandedIds(new Set());
  }, []);

  const loadCategories = useCallback(
    async (params?: { page?: number; page_size?: number; keyword?: string }) => {
      setLoading(true);
      try {
        const response = await adminCategoryApi.list({
          page: params?.page ?? 1,
          page_size: params?.page_size ?? PAGINATION_CONFIG.HARD_LIMIT,
        });
        const list = Array.isArray(response?.items) ? response.items : [];
        setCategories(list);
        if (response?.total !== undefined) setTotal(response.total);
      } catch (error) {
        console.error('Failed to fetch categories:', error);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return {
    categories,
    loading,
    total,
    tree,
    visibleNodes,
    expandedIds,
    loadCategories,
    toggleExpand,
    expandAll,
    collapseAll,
    expandNode,
    stats,
  };
}
