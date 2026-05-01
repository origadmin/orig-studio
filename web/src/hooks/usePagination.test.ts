/**
 * Unit tests for usePagination hook
 *
 * Covers AC-06 (frontend pagination default config) and AC-07 (frontend request param validation)
 * as defined in F014 AC v3.0.
 */
import { renderHook, act } from '@testing-library/react';
import { usePagination } from './usePagination';
import { PAGINATION_CONFIG } from '@/config/pagination';

describe('usePagination', () => {
  // AC-06: Default pagination state
  describe('default state', () => {
    it('should return default page=1 and pageSize=20', () => {
      const { result } = renderHook(() => usePagination());
      expect(result.current.page).toBe(1);
      expect(result.current.pageSize).toBe(20);
    });

    it('should return total=0 and totalPages=1 by default', () => {
      const { result } = renderHook(() => usePagination());
      expect(result.current.total).toBe(0);
      expect(result.current.totalPages).toBe(1);
    });

    it('should indicate first page and last page when total=0', () => {
      const { result } = renderHook(() => usePagination());
      expect(result.current.isFirstPage).toBe(true);
      expect(result.current.isLastPage).toBe(true);
    });
  });

  // AC-06: Custom initial page size
  describe('custom initialPageSize', () => {
    it('should accept custom initialPageSize=50', () => {
      const { result } = renderHook(() => usePagination({ initialPageSize: 50 }));
      expect(result.current.page).toBe(1);
      expect(result.current.pageSize).toBe(50);
    });

    it('should clamp initialPageSize=200 to MAX_PAGE_SIZE=100', () => {
      const { result } = renderHook(() => usePagination({ initialPageSize: 200 }));
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.MAX_PAGE_SIZE);
    });

    it('should correct initialPageSize=0 to DEFAULT_PAGE_SIZE=20', () => {
      const { result } = renderHook(() => usePagination({ initialPageSize: 0 }));
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
    });

    it('should correct negative initialPageSize to DEFAULT_PAGE_SIZE=20', () => {
      const { result } = renderHook(() => usePagination({ initialPageSize: -5 }));
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
    });
  });

  // Custom initialPage
  describe('custom initialPage', () => {
    it('should accept custom initialPage=3', () => {
      const { result } = renderHook(() => usePagination({ initialPage: 3 }));
      expect(result.current.page).toBe(3);
    });

    it('should correct initialPage=0 to 1', () => {
      const { result } = renderHook(() => usePagination({ initialPage: 0 }));
      expect(result.current.page).toBe(1);
    });

    it('should correct negative initialPage to 1', () => {
      const { result } = renderHook(() => usePagination({ initialPage: -1 }));
      expect(result.current.page).toBe(1);
    });
  });

  // AC-07: getParams() returns validated params
  describe('getParams', () => {
    it('should return {page: 1, page_size: 20} by default', () => {
      const { result } = renderHook(() => usePagination());
      expect(result.current.getParams()).toEqual({ page: 1, page_size: 20 });
    });

    it('should return correct params after page change', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(3);
      });
      expect(result.current.getParams()).toEqual({ page: 3, page_size: 20 });
    });

    it('should always return page >= 1', () => {
      const { result } = renderHook(() => usePagination());
      // Even if we try to set page to 0, getParams should return valid values
      act(() => {
        result.current.setPage(0);
      });
      // setPage clamps to 1, so page should be 1
      expect(result.current.page).toBe(1);
      expect(result.current.getParams().page).toBe(1);
    });

    it('should always return page_size <= MAX_PAGE_SIZE', () => {
      const { result } = renderHook(() => usePagination());
      // pageSize is clamped by setPageSize
      act(() => {
        result.current.setPageSize(200);
      });
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.MAX_PAGE_SIZE);
      expect(result.current.getParams().page_size).toBe(PAGINATION_CONFIG.MAX_PAGE_SIZE);
    });
  });

  // setPage
  describe('setPage', () => {
    it('should update page to a valid value', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(5);
      });
      expect(result.current.page).toBe(5);
    });

    it('should clamp page=0 to 1', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(0);
      });
      expect(result.current.page).toBe(1);
    });

    it('should clamp page=-1 to 1', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(-1);
      });
      expect(result.current.page).toBe(1);
    });
  });

  // setPageSize
  describe('setPageSize', () => {
    it('should update pageSize and reset to page 1', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(3);
      });
      act(() => {
        result.current.setPageSize(50);
      });
      expect(result.current.pageSize).toBe(50);
      expect(result.current.page).toBe(1);
    });

    it('should correct pageSize=0 to DEFAULT_PAGE_SIZE', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPageSize(0);
      });
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
    });

    it('should correct pageSize=200 to MAX_PAGE_SIZE=100', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPageSize(200);
      });
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.MAX_PAGE_SIZE);
    });

    it('should correct negative pageSize to DEFAULT_PAGE_SIZE', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPageSize(-10);
      });
      expect(result.current.pageSize).toBe(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
    });
  });

  // nextPage / prevPage
  describe('nextPage / prevPage', () => {
    it('should increment page with nextPage', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.nextPage();
      });
      expect(result.current.page).toBe(2);
    });

    it('should decrement page with prevPage but not below 1', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.prevPage();
      });
      expect(result.current.page).toBe(1);
    });

    it('should decrement page correctly when page > 1', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(3);
      });
      act(() => {
        result.current.prevPage();
      });
      expect(result.current.page).toBe(2);
    });
  });

  // setTotal
  describe('setTotal', () => {
    it('should update total and recalculate totalPages', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setTotal(100);
      });
      expect(result.current.total).toBe(100);
      expect(result.current.totalPages).toBe(5); // 100 / 20 = 5
    });

    it('should calculate totalPages=1 when total < pageSize', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setTotal(15);
      });
      expect(result.current.totalPages).toBe(1);
    });

    it('should calculate totalPages correctly with remainder', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setTotal(25);
      });
      expect(result.current.totalPages).toBe(2); // ceil(25/20) = 2
    });
  });

  // isFirstPage / isLastPage
  describe('isFirstPage / isLastPage', () => {
    it('should correctly indicate first and last page', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setTotal(60);
      });
      // page=1, totalPages=3
      expect(result.current.isFirstPage).toBe(true);
      expect(result.current.isLastPage).toBe(false);

      act(() => {
        result.current.setPage(2);
      });
      expect(result.current.isFirstPage).toBe(false);
      expect(result.current.isLastPage).toBe(false);

      act(() => {
        result.current.setPage(3);
      });
      expect(result.current.isFirstPage).toBe(false);
      expect(result.current.isLastPage).toBe(true);
    });
  });

  // reset
  describe('reset', () => {
    it('should reset to default values', () => {
      const { result } = renderHook(() => usePagination());
      act(() => {
        result.current.setPage(5);
        result.current.setPageSize(50);
        result.current.setTotal(200);
      });
      act(() => {
        result.current.reset();
      });
      expect(result.current.page).toBe(1);
      expect(result.current.pageSize).toBe(20);
      expect(result.current.total).toBe(0);
    });

    it('should reset to custom initial values', () => {
      const { result } = renderHook(() =>
        usePagination({ initialPage: 2, initialPageSize: 50 })
      );
      act(() => {
        result.current.setPage(5);
        result.current.setPageSize(100);
        result.current.setTotal(200);
      });
      act(() => {
        result.current.reset();
      });
      expect(result.current.page).toBe(2);
      expect(result.current.pageSize).toBe(50);
      expect(result.current.total).toBe(0);
    });
  });
});

describe('PAGINATION_CONFIG constants', () => {
  it('should have DEFAULT_PAGE=1', () => {
    expect(PAGINATION_CONFIG.DEFAULT_PAGE).toBe(1);
  });

  it('should have DEFAULT_PAGE_SIZE=20', () => {
    expect(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE).toBe(20);
  });

  it('should have MAX_PAGE_SIZE=100', () => {
    expect(PAGINATION_CONFIG.MAX_PAGE_SIZE).toBe(100);
  });

  it('should have PAGE_SIZE_OPTIONS=[10, 20, 50, 100]', () => {
    expect(PAGINATION_CONFIG.PAGE_SIZE_OPTIONS).toEqual([10, 20, 50, 100]);
  });
});
