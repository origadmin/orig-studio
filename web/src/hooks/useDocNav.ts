/**
 * useDocNav - Hook for loading category tree and article lists
 * for the DocSidebar navigation.
 *
 * Uses TanStack Query for caching and stale-while-revalidate.
 */
import { useQuery } from '@tanstack/react-query';
import { categoryApi } from '@/lib/api/category';
import { articleApi } from '@/lib/api/article';
import { buildCategoryTree, type CategoryTreeNode } from '@/lib/utils/categoryTree';

/** Load the full category tree for sidebar navigation */
export function useDocCategoryTree() {
  return useQuery({
    queryKey: ['categories', 'tree'],
    queryFn: async () => {
      const res = await categoryApi.getAll({ page: 1, page_size: 200 });
      return buildCategoryTree(res.items || []);
    },
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
  });
}

/** Load articles for a specific category (lazy-loaded on expand) */
export function useDocCategoryArticles(categoryId: number | undefined) {
  return useQuery({
    queryKey: ['articles', 'category', categoryId],
    queryFn: () =>
      articleApi.list({
        page: 1,
        page_size: 50,
        category_id: categoryId,
        state: 'published',
      }),
    enabled: categoryId !== undefined,
    staleTime: 30_000,
    gcTime: 5 * 60_000,
  });
}

/** Load latest articles for the doc home page */
export function useDocLatestArticles(limit = 5) {
  return useQuery({
    queryKey: ['articles', 'latest', limit],
    queryFn: () => articleApi.latest(limit),
    staleTime: 30_000,
    gcTime: 5 * 60_000,
  });
}

export type { CategoryTreeNode };
