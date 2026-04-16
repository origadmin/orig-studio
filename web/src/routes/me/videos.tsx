/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import MyVideosPage from '../../pages/home/me/MyVideos';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/videos',
  component: () => <MyVideosPage />,
});
