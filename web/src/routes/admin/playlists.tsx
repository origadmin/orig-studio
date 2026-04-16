/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminPlaylists from '../../pages/admin/Playlists';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/playlists',
  component: () => <AdminPlaylists />,
});
