/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import UploadPage from '../../pages/home/me/Upload';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/upload',
  component: () => <UploadPage />,
});
