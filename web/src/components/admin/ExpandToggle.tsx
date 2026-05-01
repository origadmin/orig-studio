/**
 * ExpandToggle atom - chevron button for expanding/collapsing tree nodes.
 *
 * - Shows ChevronRight (collapsed) or ChevronDown (expanded) when node has children
 * - Invisible spacer when no children (maintains alignment)
 * - Accessible: aria-expanded and aria-label
 */
import React from 'react';
import { ChevronRight, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface ExpandToggleProps {
  hasChildren: boolean;
  isExpanded: boolean;
  onClick: () => void;
  label: string;
}

export const ExpandToggle: React.FC<ExpandToggleProps> = React.memo(
  ({ hasChildren, isExpanded, onClick, label }) => {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          'p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors shrink-0',
          !hasChildren && 'invisible'
        )}
        aria-expanded={hasChildren ? isExpanded : undefined}
        aria-label={
          hasChildren
            ? `${isExpanded ? 'Collapse' : 'Expand'} ${label}`
            : undefined
        }
      >
        {hasChildren ? (
          isExpanded ? (
            <ChevronDown className="w-4 h-4" />
          ) : (
            <ChevronRight className="w-4 h-4" />
          )
        ) : (
          <span className="w-4 h-4 inline-block" />
        )}
      </button>
    );
  }
);

ExpandToggle.displayName = 'ExpandToggle';
