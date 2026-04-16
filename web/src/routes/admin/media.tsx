/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminMedia from '../../pages/admin/Media';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/media',
  component: () => <AdminMedia />,
});
