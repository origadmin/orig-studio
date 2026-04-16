/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import TagsPage from '../pages/home/Tags';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/tags',
  component: () => <TagsPage />,
});
