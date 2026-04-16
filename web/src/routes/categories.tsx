/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import SearchPage from '../pages/home/Search';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/categories',
  component: () => <SearchPage />,
});
