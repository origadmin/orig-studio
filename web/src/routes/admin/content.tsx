/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminContent from '../../pages/admin/Content';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/content',
  component: () => <AdminContent />,
});
