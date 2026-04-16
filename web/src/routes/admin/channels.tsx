/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as AdminLayoutRoute } from './_layout';
import AdminChannels from '../../pages/admin/Channels';

export const Route = createRoute({
  getParentRoute: () => AdminLayoutRoute,
  path: '/admin/channels',
  component: () => <AdminChannels />,
});
