/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminUsers from '../../pages/admin/Users';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/users',
  component: () => <AdminUsers />,
});
