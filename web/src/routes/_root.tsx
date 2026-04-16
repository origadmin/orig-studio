/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRootRoute } from '@tanstack/react-router';

import { Outlet } from '@tanstack/react-router';

export const Route = createRootRoute({
  component: () => <Outlet />,
});
