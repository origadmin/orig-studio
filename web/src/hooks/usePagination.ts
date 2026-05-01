import { useState, useCallback, useMemo } from 'react';
import { PAGINATION_CONFIG, normalizePagination } from '@/config/pagination';

/** Check if running in development mode (compatible with both Rsbuild and Jest) */
const isDev = process.env.NODE_ENV === 'development';

/**
 * Pagination state and controls returned by usePagination.
 */
export interface PaginationState {
  /** Current page number (1-based) */
  page: number;
  /** Current page size */
  pageSize: number;
  /** Total item count from server response */
  total: number;
  /** Computed total pages */
  totalPages: number;
  /** Whether current page is the first page */
  isFirstPage: boolean;
  /** Whether current page is the last page */
  isLastPage: boolean;
}

/**
 * Pagination actions returned by usePagination.
 */
export interface PaginationActions {
  /** Go to a specific page (auto-clamped to >= 1) */
  setPage: (page: number) => void;
  /** Change page size (auto-clamped to [1, MAX_PAGE_SIZE], resets to page 1) */
  setPageSize: (pageSize: number) => void;
  /** Go to next page */
  nextPage: () => void;
  /** Go to previous page */
  prevPage: () => void;
  /** Update total count from server response */
  setTotal: (total: number) => void;
  /** Reset to default page and page size */
  reset: () => void;
  /** Get validated pagination params for API requests */
  getParams: () => { page: number; page_size: number };
}

export interface UsePaginationOptions {
  /** Initial page number, defaults to 1 */
  initialPage?: number;
  /** Initial page size, defaults to PAGINATION_CONFIG.DEFAULT_PAGE_SIZE */
  initialPageSize?: number;
}

/**
 * usePagination provides centralized pagination state management.
 *
 * Features:
 * - Enforces PAGINATION_CONFIG.DEFAULT_PAGE_SIZE and PAGINATION_CONFIG.MAX_PAGE_SIZE constraints
 * - Auto-resets to page 1 when page size changes
 * - Provides computed totalPages, isFirstPage, isLastPage
 * - getParams() returns validated API query params (never sends page<1 or pageSize>MAX)
 *
 * @example
 * ```tsx
 * const { page, pageSize, totalPages, nextPage, prevPage, setTotal, getParams } = usePagination();
 *
 * const { data } = useQuery({
 *   queryKey: ['items', getParams()],
 *   queryFn: () => api.list(getParams()),
 * });
 *
 * // After receiving response:
 * setTotal(data.total);
 * ```
 */
export function usePagination(
  options: UsePaginationOptions = {}
): PaginationState & PaginationActions {
  const { initialPage, initialPageSize } = options;

  const [normalizedPage, normalizedPageSize] = normalizePagination(
    initialPage,
    initialPageSize
  );

  const [page, setPageInternal] = useState(normalizedPage);
  const [pageSize, setPageSizeInternal] = useState(normalizedPageSize);
  const [total, setTotal] = useState(0);

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(total / pageSize)),
    [total, pageSize]
  );

  const isFirstPage = page <= 1;
  const isLastPage = page >= totalPages;

  const setPage = useCallback((p: number) => {
    // Clamp page to >= 1
    if (p < 1) {
      if (isDev) {
        console.warn(`[usePagination] page ${p} corrected to 1`);
      }
      p = 1;
    }
    setPageInternal(p);
  }, []);

  const setPageSize = useCallback((ps: number) => {
    // Enforce constraints: clamp to [1, MAX_PAGE_SIZE]
    if (ps <= 0) {
      if (isDev) {
        console.warn(`[usePagination] pageSize ${ps} corrected to ${PAGINATION_CONFIG.DEFAULT_PAGE_SIZE}`);
      }
      ps = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE;
    }
    if (ps > PAGINATION_CONFIG.MAX_PAGE_SIZE) {
      if (isDev) {
        console.warn(`[usePagination] pageSize ${ps} corrected to ${PAGINATION_CONFIG.MAX_PAGE_SIZE}`);
      }
      ps = PAGINATION_CONFIG.MAX_PAGE_SIZE;
    }
    setPageSizeInternal(ps);
    // Reset to page 1 when page size changes
    setPageInternal(1);
  }, []);

  const nextPage = useCallback(() => {
    setPageInternal((prev) => prev + 1);
  }, []);

  const prevPage = useCallback(() => {
    setPageInternal((prev) => Math.max(1, prev - 1));
  }, []);

  const reset = useCallback(() => {
    setPageInternal(normalizedPage);
    setPageSizeInternal(normalizedPageSize);
    setTotal(0);
  }, [normalizedPage, normalizedPageSize]);

  const getParams = useCallback(
    // Double-check: ensure returned params are always valid
    // This is a safety net in case state is corrupted
    () => {
      const validatedPage = Math.max(1, page);
      const validatedPageSize = Math.min(
        Math.max(1, pageSize),
        PAGINATION_CONFIG.MAX_PAGE_SIZE
      );
      return { page: validatedPage, page_size: validatedPageSize };
    },
    [page, pageSize]
  );

  return {
    page,
    pageSize,
    total,
    totalPages,
    isFirstPage,
    isLastPage,
    setPage,
    setPageSize,
    nextPage,
    prevPage,
    setTotal,
    reset,
    getParams,
  };
}
