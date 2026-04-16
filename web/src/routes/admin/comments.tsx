/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminComments from '../../pages/admin/Comments';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/comments',
  component: () => <AdminComments />,
});
