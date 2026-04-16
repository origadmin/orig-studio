/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { api } from '@/api';
import { MOCK_MODE } from './config';

// Mock the MOCK_MODE for testing
jest.mock('./config', () => ({
  MOCK_MODE: true,
}));

describe('Mock API', () => {
  it('should return mock medias when MOCK_MODE is true', async () => {
    const response = await api.get('/medias');
    expect(response).toHaveProperty('items');
    expect(Array.isArray(response.items)).toBe(true);
    expect(response.items.length).toBeGreaterThan(0);
  });
});
