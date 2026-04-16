/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from '../_layout';
import AdminTranscodingStatus from '../../../pages/admin/TranscodingStatus';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/transcoding/status',
  component: () => <AdminTranscodingStatus />,
});
