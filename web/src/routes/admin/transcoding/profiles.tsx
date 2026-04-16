/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from '../_layout';
import AdminTranscodingProfiles from '../../../pages/admin/TranscodingProfiles';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/transcoding/profiles',
  component: () => <AdminTranscodingProfiles />,
});
