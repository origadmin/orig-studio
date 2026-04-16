/**
 * Extracts list data from API response with defensive programming
 * Handles different response formats:
 * - Direct array: [{...}, {...}]
 * - Object with list field: {list: [...], total: N}
 * - Object with items field: {items: [...], total: N}
 * - Wrapped object: {code: 0, message: "ok", data: {...}}
 * - Empty object or null/undefined
 */
export function extractList<T>(response: unknown): T[] {
  if (Array.isArray(response)) {
    return response as T[];
  }
  
  if (response && typeof response === 'object') {
    const obj = response as Record<string, unknown>;
    
    // Check for wrapped response with code, message, data
    if ('data' in obj && typeof obj.data === 'object' && obj.data !== null) {
      const dataObj = obj.data as Record<string, unknown>;
      if (Array.isArray(dataObj.list)) {
        return dataObj.list as T[];
      }
      if (Array.isArray(dataObj.items)) {
        return dataObj.items as T[];
      }
      if (Array.isArray(dataObj.data)) {
        return dataObj.data as T[];
      }
    }
    
    // Check for direct list/items fields
    if (Array.isArray(obj.list)) {
      return obj.list as T[];
    }
    if (Array.isArray(obj.items)) {
      return obj.items as T[];
    }
  }
  
  return [];
}

/**
 * Extracts pagination info from API response
 */
export function extractPagination(response: unknown): {
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
} {
  if (response && typeof response === 'object') {
    const obj = response as Record<string, unknown>;
    
    // Check for wrapped response with code, message, data
    let target = obj;
    if ('data' in obj && typeof obj.data === 'object' && obj.data !== null) {
      target = obj.data as Record<string, unknown>;
    }
    
    return {
      total: typeof target.total === 'number' ? target.total : 0,
      page: typeof target.page === 'number' ? target.page : 1,
      pageSize: typeof target.page_size === 'number' ? target.page_size : 20,
      totalPages: typeof target.total_pages === 'number' ? target.total_pages : 0,
    };
  }
  
  return {
    total: 0,
    page: 1,
    pageSize: 20,
    totalPages: 0,
  };
}
