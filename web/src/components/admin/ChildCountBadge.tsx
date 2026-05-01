/**
 * ChildCountBadge atom - shows the number of direct children for a tree node.
 *
 * Only rendered when count > 0.
 */
import React from 'react';
import { Badge } from '@/components/ui/badge';

export interface ChildCountBadgeProps {
  count: number;
}

export const ChildCountBadge: React.FC<ChildCountBadgeProps> = React.memo(
  ({ count }) => {
    if (count <= 0) return null;

    return (
      <Badge
        variant="outline"
        className="text-xs px-1.5 py-0 h-5 font-normal text-muted-foreground"
      >
        {count}
      </Badge>
    );
  }
);

ChildCountBadge.displayName = 'ChildCountBadge';
