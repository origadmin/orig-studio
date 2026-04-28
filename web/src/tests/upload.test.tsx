import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock API calls
vi.mock('@/lib/api/upload', () => ({
  uploadApi: {
    InitiateMultipartUpload: vi.fn(),
    UploadPart: vi.fn(),
    CompleteMultipartUpload: vi.fn(),
    AbortMultipartUpload: vi.fn(),
    ListParts: vi.fn(),
  },
}));

vi.mock('@/lib/api/media', () => ({
  mediaApi: {
    ListMedia: vi.fn(),
    GetMedia: vi.fn(),
  },
}));

vi.mock('@/lib/api/encoding', () => ({
  encodingApi: {
    ListEncodingTasks: vi.fn(),
    RetryEncodingTask: vi.fn(),
  },
}));

import { uploadApi } from '@/lib/api/upload';
import { mediaApi } from '@/lib/api/media';
import { encodingApi } from '@/lib/api/encoding';
import { UploadPage } from '@/pages/admin/Upload';
import { MediaPage } from '@/pages/admin/Media';
import { TranscodingStatusPage } from '@/pages/admin/TranscodingStatus';
import { TranscodingProfilesPage } from '@/pages/admin/TranscodingProfiles';

// Mock file for testing
const createMockFile = (name: string, size: number, type: string): File => {
  const blob = new Blob(['mock content'.repeat(size / 12)], { type });
  return new File([blob], name, { type });
};

// Mock upload session data
const mockUploadSession = {
  upload_id: 'test-upload-id',
  total_parts: 2,
  chunk_size: 5 * 1024 * 1024, // 5MB
};

// Mock media data
const mockMedia = {
  id: 'test-media-id',
  title: 'Test Video',
  description: 'A test video',
  url: 'test.mp4',
  size: 10 * 1024 * 1024, // 10MB
  mime_type: 'video/mp4',
  thumbnail: 'thumbnails/test.jpg',
  encoding_status: 'pending',
  hls_file: '',
  created_at: new Date().toISOString(),
};

// Mock encoding task data
const mockEncodingTask = {
  id: 'test-task-id',
  media_id: 'test-media-id',
  profile_id: 1,
  status: 'pending',
  output_path: '',
  error_message: '',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('Frontend Upload Tests', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    vi.clearAllMocks();
  });

  afterEach(() => {
    // Reset all mocks after each test
    vi.resetAllMocks();
  });

  describe('UploadPage', () => {
    it('should render upload form', () => {
      render(<UploadPage />);
      expect(screen.getByText('Upload Media')).toBeInTheDocument();
      expect(screen.getByLabelText('File')).toBeInTheDocument();
      expect(screen.getByLabelText('Title')).toBeInTheDocument();
      expect(screen.getByLabelText('Description')).toBeInTheDocument();
    });

    it('should handle file selection', () => {
      render(<UploadPage />);
      const fileInput = screen.getByLabelText('File') as HTMLInputElement;
      const mockFile = createMockFile('test.mp4', 10 * 1024 * 1024, 'video/mp4');
      
      fireEvent.change(fileInput, {
        target: { files: [mockFile] },
      });
      
      expect(fileInput.files?.length).toBe(1);
      expect(fileInput.files?.[0].name).toBe('test.mp4');
    });

    it('should handle upload initiation', async () => {
      // Mock API response
      (uploadApi.InitiateMultipartUpload as vi.Mock).mockResolvedValue({
        data: mockUploadSession,
      });

      render(<UploadPage />);
      
      // Select file
      const fileInput = screen.getByLabelText('File') as HTMLInputElement;
      const mockFile = createMockFile('test.mp4', 10 * 1024 * 1024, 'video/mp4');
      fireEvent.change(fileInput, {
        target: { files: [mockFile] },
      });

      // Fill form
      fireEvent.change(screen.getByLabelText('Title'), {
        target: { value: 'Test Video' },
      });

      // Submit form
      fireEvent.click(screen.getByText('Upload'));

      // Wait for API call
      await waitFor(() => {
        expect(uploadApi.InitiateMultipartUpload).toHaveBeenCalled();
      });
    });

    it('should handle upload progress', async () => {
      // Mock API responses
      (uploadApi.InitiateMultipartUpload as vi.Mock).mockResolvedValue({
        data: mockUploadSession,
      });
      
      (uploadApi.UploadPart as vi.Mock).mockResolvedValue({
        data: {
          etag: 'test-etag',
          size: 5 * 1024 * 1024,
        },
      });
      
      (uploadApi.CompleteMultipartUpload as vi.Mock).mockResolvedValue({
        data: {
          media: mockMedia,
        },
      });

      render(<UploadPage />);
      
      // Select file
      const fileInput = screen.getByLabelText('File') as HTMLInputElement;
      const mockFile = createMockFile('test.mp4', 10 * 1024 * 1024, 'video/mp4');
      fireEvent.change(fileInput, {
        target: { files: [mockFile] },
      });

      // Fill form
      fireEvent.change(screen.getByLabelText('Title'), {
        target: { value: 'Test Video' },
      });

      // Submit form
      fireEvent.click(screen.getByText('Upload'));

      // Wait for upload to complete
      await waitFor(() => {
        expect(uploadApi.CompleteMultipartUpload).toHaveBeenCalled();
      }, { timeout: 10000 });

      // Check success message
      expect(screen.getByText('Upload completed successfully!')).toBeInTheDocument();
    });

    it('should handle upload failure', async () => {
      // Mock API responses
      (uploadApi.InitiateMultipartUpload as vi.Mock).mockResolvedValue({
        data: mockUploadSession,
      });
      
      (uploadApi.UploadPart as vi.Mock).mockRejectedValue(new Error('Upload failed'));

      render(<UploadPage />);
      
      // Select file
      const fileInput = screen.getByLabelText('File') as HTMLInputElement;
      const mockFile = createMockFile('test.mp4', 10 * 1024 * 1024, 'video/mp4');
      fireEvent.change(fileInput, {
        target: { files: [mockFile] },
      });

      // Fill form
      fireEvent.change(screen.getByLabelText('Title'), {
        target: { value: 'Test Video' },
      });

      // Submit form
      fireEvent.click(screen.getByText('Upload'));

      // Wait for error
      await waitFor(() => {
        expect(screen.getByText('Upload failed')).toBeInTheDocument();
      });
    });
  });

  describe('MediaPage', () => {
    it('should render media list', async () => {
      // Mock API response
      (mediaApi.ListMedia as vi.Mock).mockResolvedValue({
        data: {
          items: [mockMedia],
          total: 1,
        },
      });

      render(<MediaPage />);

      // Wait for media to load
      await waitFor(() => {
        expect(screen.getByText('Test Video')).toBeInTheDocument();
      });
    });

    it('should show encoding status', async () => {
      // Mock API response with pending status
      const pendingMedia = {
        ...mockMedia,
        encoding_status: 'pending',
      };
      
      (mediaApi.ListMedia as vi.Mock).mockResolvedValue({
        data: {
          items: [pendingMedia],
          total: 1,
        },
      });

      render(<MediaPage />);

      // Wait for media to load
      await waitFor(() => {
        expect(screen.getByText('pending')).toBeInTheDocument();
      });
    });

    it('should show success status', async () => {
      // Mock API response with success status
      const successMedia = {
        ...mockMedia,
        encoding_status: 'success',
        hls_file: 'hls/test/master.m3u8',
      };
      
      (mediaApi.ListMedia as vi.Mock).mockResolvedValue({
        data: {
          items: [successMedia],
          total: 1,
        },
      });

      render(<MediaPage />);

      // Wait for media to load
      await waitFor(() => {
        expect(screen.getByText('success')).toBeInTheDocument();
      });
    });

    it('should show failed status', async () => {
      // Mock API response with failed status
      const failedMedia = {
        ...mockMedia,
        encoding_status: 'failed',
      };
      
      (mediaApi.ListMedia as vi.Mock).mockResolvedValue({
        data: {
          items: [failedMedia],
          total: 1,
        },
      });

      render(<MediaPage />);

      // Wait for media to load
      await waitFor(() => {
        expect(screen.getByText('failed')).toBeInTheDocument();
      });
    });
  });

  describe('TranscodingStatusPage', () => {
    it('should render transcoding tasks', async () => {
      // Mock API response
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [mockEncodingTask],
          total: 1,
        },
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('test-media-id')).toBeInTheDocument();
      });
    });

    it('should show pending status', async () => {
      // Mock API response with pending status
      const pendingTask = {
        ...mockEncodingTask,
        status: 'pending',
      };
      
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [pendingTask],
          total: 1,
        },
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('pending')).toBeInTheDocument();
      });
    });

    it('should show processing status', async () => {
      // Mock API response with processing status
      const processingTask = {
        ...mockEncodingTask,
        status: 'processing',
      };
      
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [processingTask],
          total: 1,
        },
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('processing')).toBeInTheDocument();
      });
    });

    it('should show success status', async () => {
      // Mock API response with success status
      const successTask = {
        ...mockEncodingTask,
        status: 'success',
        output_path: 'hls/test/720p/index.m3u8',
      };
      
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [successTask],
          total: 1,
        },
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('success')).toBeInTheDocument();
      });
    });

    it('should show failed status', async () => {
      // Mock API response with failed status
      const failedTask = {
        ...mockEncodingTask,
        status: 'failed',
        error_message: 'Transcode failed',
      };
      
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [failedTask],
          total: 1,
        },
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('failed')).toBeInTheDocument();
        expect(screen.getByText('Transcode failed')).toBeInTheDocument();
      });
    });

    it('should handle retry functionality', async () => {
      // Mock API response with failed status
      const failedTask = {
        ...mockEncodingTask,
        status: 'failed',
        error_message: 'Transcode failed',
      };
      
      (encodingApi.ListEncodingTasks as vi.Mock).mockResolvedValue({
        data: {
          items: [failedTask],
          total: 1,
        },
      });
      
      (encodingApi.RetryEncodingTask as vi.Mock).mockResolvedValue({
        data: {},
      });

      render(<TranscodingStatusPage />);

      // Wait for tasks to load
      await waitFor(() => {
        expect(screen.getByText('failed')).toBeInTheDocument();
      });

      // Click retry button
      fireEvent.click(screen.getByText('Retry'));

      // Wait for retry API call
      await waitFor(() => {
        expect(encodingApi.RetryEncodingTask).toHaveBeenCalled();
      });
    });
  });

  describe('TranscodingProfilesPage', () => {
    it('should render transcoding profiles', () => {
      render(<TranscodingProfilesPage />);
      expect(screen.getByText('Transcoding Profiles')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      // Mock API response with error
      (mediaApi.ListMedia as vi.Mock).mockRejectedValue(new Error('Network error'));

      render(<MediaPage />);

      // Wait for error
      await waitFor(() => {
        expect(screen.getByText('Error loading media')).toBeInTheDocument();
      });
    });

    it('should handle empty state', async () => {
      // Mock API response with empty list
      (mediaApi.ListMedia as vi.Mock).mockResolvedValue({
        data: {
          items: [],
          total: 0,
        },
      });

      render(<MediaPage />);

      // Wait for empty state
      await waitFor(() => {
        expect(screen.getByText('No media found')).toBeInTheDocument();
      });
    });
  });

  describe('Loading States', () => {
    it('should show loading spinner during data fetch', async () => {
      // Mock API with delay
      (mediaApi.ListMedia as vi.Mock).mockImplementation(() => {
        return new Promise(resolve => {
          setTimeout(() => {
            resolve({
              data: {
                items: [mockMedia],
                total: 1,
              },
            });
          }, 500);
        });
      });

      render(<MediaPage />);

      // Check loading state
      expect(screen.getByText('Loading...')).toBeInTheDocument();

      // Wait for data to load
      await waitFor(() => {
        expect(screen.getByText('Test Video')).toBeInTheDocument();
      });
    });
  });
});
