/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import ChannelPage from '../pages/home/Channel';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/c/$id',
  component: () => <ChannelPage />,
});
