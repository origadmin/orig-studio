/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import TestPage from '../pages/test';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/test',
  component: () => <TestPage />,
});
