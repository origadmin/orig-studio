/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import MembersPage from '../pages/home/Members';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/members',
  component: () => <MembersPage />,
});
