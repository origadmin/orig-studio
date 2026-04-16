/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import PlaylistsPage from '../../pages/home/me/Playlists';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/playlists',
  component: () => <PlaylistsPage />,
});
