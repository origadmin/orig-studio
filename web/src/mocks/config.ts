/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Mock mode configuration
export const MOCK_MODE = process.env.NODE_ENV === 'development' && process.env.REACT_APP_MOCK_MODE === 'true';
