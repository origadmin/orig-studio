/**
 * Centralized pagination configuration for OrigStudio.
 *
 * These values MUST align with the backend PaginationConfig:
 *   - default_page_size: 20
 *   - max_page_size: 100
 *   - hard_limit: 1000
 *
 * When backend config changes, update here accordingly.
 */

export const PAGINATION_CONFIG = {
  /** Default page size for list queries */
  DEFAULT_PAGE_SIZE: 20,

  /** Maximum page size allowed for normal queries */
  MAX_PAGE_SIZE: 100,

  /** Hard limit for no-paging / export scenarios */
  HARD_LIMIT: 1000,

  /** Default starting page number */
  DEFAULT_PAGE: 1,

  /** Available page size options for UI selectors */
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100] as const,
} as const;

/** Type for valid page size option values */
export type PageSizeOption = typeof PAGINATION_CONFIG.PAGE_SIZE_OPTIONS[number];

/**
 * Backward-compatible alias for PAGINATION_CONFIG.
 * Prefer using PAGINATION_CONFIG directly in new code.
 * @deprecated Use PAGINATION_CONFIG instead.
 */
export const PAGINATION = PAGINATION_CONFIG;

/**
 * Validates and normalizes pagination parameters.
 * Mirrors the backend NormalizePagination logic.
 *
 * @param page - Requested page number (1-based)
 * @param pageSize - Requested page size
 * @returns Normalized [page, pageSize] tuple
 */
export function normalizePagination(
  page?: number | null,
  pageSize?: number | null
): [number, number] {
  let p = page ?? PAGINATION.DEFAULT_PAGE;
  let ps = pageSize ?? PAGINATION.DEFAULT_PAGE_SIZE;

  if (p < 1) {
    p = 1;
  }
  if (ps <= 0) {
    ps = PAGINATION.DEFAULT_PAGE_SIZE;
  }
  if (ps > PAGINATION.MAX_PAGE_SIZE) {
    ps = PAGINATION.MAX_PAGE_SIZE;
  }

  return [p, ps];
}

/**
 * Builds pagination query parameters for API requests.
 * Ensures page_size never exceeds MAX_PAGE_SIZE.
 */
export function buildPaginationParams(
  page?: number | null,
  pageSize?: number | null
): { page: number; page_size: number } {
  const [p, ps] = normalizePagination(page, pageSize);
  return { page: p, page_size: ps };
}
