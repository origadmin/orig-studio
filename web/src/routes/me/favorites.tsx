/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import FavoritesPage from '../../pages/home/me/Favorites';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/favorites',
  component: () => <FavoritesPage />,
});
