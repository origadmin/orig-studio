/**
 * CommentContentCell molecule - displays comment content with
 * expand/collapse for long text.
 */
import React, { useState } from 'react';
import { BlockedContentPlaceholder } from './BlockedContentPlaceholder';

interface CommentContentCellProps {
  content: string;
  status?: string;
  maxLength?: number;
}

export const CommentContentCell: React.FC<CommentContentCellProps> = React.memo(
  ({ content, status, maxLength = 150 }) => {
    const [expanded, setExpanded] = useState(false);

    // Show blocked placeholder for blocked comments
    if (status === 'blocked') {
      return <BlockedContentPlaceholder />;
    }

    if (!content) {
      return <span className="text-muted-foreground italic">No content</span>;
    }

    const isTruncatable = content.length > maxLength;
    const displayText = expanded || !isTruncatable
      ? content
      : content.slice(0, maxLength) + '...';

    return (
      <div
        className="min-w-[300px] max-w-full"
        onClick={() => isTruncatable && setExpanded(!expanded)}
        role={isTruncatable ? 'button' : undefined}
        tabIndex={isTruncatable ? 0 : undefined}
        onKeyDown={(e) => {
          if (isTruncatable && (e.key === 'Enter' || e.key === ' ')) {
            e.preventDefault();
            setExpanded(!expanded);
          }
        }}
      >
        <p className={`text-sm whitespace-pre-wrap break-words ${isTruncatable ? 'cursor-pointer hover:text-foreground' : ''}`}>
          {displayText}
        </p>
        {expanded && content.length > 500 && (
          <div className="max-h-[200px] overflow-y-auto mt-1">
            {/* Already shown above, this is just for very long content scroll */}
          </div>
        )}
      </div>
    );
  }
);

CommentContentCell.displayName = 'CommentContentCell';
