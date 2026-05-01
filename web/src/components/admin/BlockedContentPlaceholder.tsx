/**
 * BlockedContentPlaceholder atom - shows "This comment has been blocked"
 * with a shield-off icon.
 */
import React from 'react';
import { ShieldOff } from 'lucide-react';

export const BlockedContentPlaceholder: React.FC = React.memo(() => (
  <div className="flex items-center gap-2 text-muted-foreground italic py-1">
    <ShieldOff className="h-4 w-4" />
    <span>This comment has been blocked</span>
  </div>
));

BlockedContentPlaceholder.displayName = 'BlockedContentPlaceholder';
