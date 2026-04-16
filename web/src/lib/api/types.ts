// Unified paginated response type
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages?: number;
}

// Generic API response wrapper
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}
