/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Outlet } from '@tanstack/react-router';
import PortalLayout from '../layout/PortalLayout';

import { Route as RootRoute } from './__root';

export const Route = createRoute({
  getParentRoute: () => RootRoute,
  id: 'portal',
  component: () => <PortalLayout />,
});
