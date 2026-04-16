/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminCategories from '../../pages/admin/Categories';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/categories',
  component: () => <AdminCategories />,
});
