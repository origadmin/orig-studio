/**
 * CascadingStatusHint atom - shows amber text when an ancestor is disabled.
 *
 * Displays "(Parent Disabled)" next to the category name when
 * the isAncestorDisabled flag is true.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';

export interface CascadingStatusHintProps {
  isAncestorDisabled: boolean;
}

export const CascadingStatusHint: React.FC<CascadingStatusHintProps> = React.memo(
  ({ isAncestorDisabled }) => {
    const { t } = useTranslation();

    if (!isAncestorDisabled) return null;

    return (
      <span className="text-xs text-warning whitespace-nowrap">
        ({t('admin.parentDisabled') || 'Parent Disabled'})
      </span>
    );
  }
);

CascadingStatusHint.displayName = 'CascadingStatusHint';
