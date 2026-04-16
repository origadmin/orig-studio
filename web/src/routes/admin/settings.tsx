/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminSettings from '../../pages/admin/Settings';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/settings',
  component: () => <AdminSettings />,
});
