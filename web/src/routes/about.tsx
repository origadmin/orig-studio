/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import AboutPage from '../pages/home/About';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/about',
  component: () => <AboutPage />,
});
