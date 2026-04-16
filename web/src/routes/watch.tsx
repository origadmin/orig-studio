/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import WatchPage from '../pages/home/Watch';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/watch',
  validateSearch: (search: Record<string, unknown>) => {
    // 清理 v 参数，移除引号和空格
    const v = search.v ? String(search.v).replace(/["']/g, '').trim() : undefined;
    return { v };
  },
  component: () => <WatchPage />,
});
