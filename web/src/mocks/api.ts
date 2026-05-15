/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { MediaItem } from '../types/media';

// Mock media data
export const mockMedias: MediaItem[] = [
  {
    id: 1,
    title: 'Introduction to OrigAdmin',
    thumbnail_url: 'https://neeko-copilot.bytedance.net/api/text2image?prompt=Introduction%20to%20OrigAdmin&size=1280x720',
    duration: 300,
    create_time: new Date().toISOString(),
  },
  {
    id: 2,
    title: 'Getting Started with OrigStudio',
    thumbnail_url: 'https://neeko-copilot.bytedance.net/api/text2image?prompt=Getting%20Started%20with%20OrigStudio&size=1280x720',
    duration: 420,
    create_time: new Date().toISOString(),
  },
  {
    id: 3,
    title: 'Advanced Features of OrigAdmin',
    thumbnail_url: 'https://neeko-copilot.bytedance.net/api/text2image?prompt=Advanced%20Features%20of%20OrigAdmin&size=1280x720',
    duration: 540,
    create_time: new Date().toISOString(),
  },
];

// Mock API response
export const mockApi = {
  getMedias: () => {
    return new Promise<{ items: Media[]; total: number }>((resolve) => {
      setTimeout(() => {
        resolve({ items: mockMedias, total: mockMedias.length });
      }, 500);
    });
  },
  getMedia: (id: number) => {
    return new Promise<Media | null>((resolve) => {
      setTimeout(() => {
        const media = mockMedias.find((m) => m.id === id);
        resolve(media || null);
      }, 300);
    });
  },
};
