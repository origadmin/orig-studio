/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminTags from '../../pages/admin/Tags';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/tags',
  component: () => <AdminTags />,
});
