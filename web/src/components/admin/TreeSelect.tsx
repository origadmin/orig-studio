/**
 * TreeSelect molecule - tree-based parent category selector with
 * indentation in the dropdown.
 *
 * Uses shadcn Select component with custom SelectItem rendering.
 * Each option is padded by depth * 16px for visual indentation.
 */
import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { CategoryTreeNode } from '@/lib/utils/categoryTree';
import {
  getTreeSelectOptions,
  getDescendantIds,
  type TreeSelectOption,
} from '@/lib/utils/categoryTree';

export interface TreeSelectProps {
  tree: CategoryTreeNode[];
  value?: number;
  onChange: (value: number | undefined) => void;
  excludeId?: number;
  placeholder?: string;
  disabled?: boolean;
}

const NONE_VALUE = '__none__';

export const TreeSelect: React.FC<TreeSelectProps> = React.memo(
  ({ tree, value, onChange, excludeId, placeholder, disabled }) => {
    const { t } = useTranslation();

    // Calculate descendant IDs to exclude (for circular reference prevention)
    const excludeDescendantIds = useMemo(() => {
      if (excludeId === undefined) return [];
      return getDescendantIds(tree, excludeId);
    }, [tree, excludeId]);

    // Generate tree select options
    const options: TreeSelectOption[] = useMemo(
      () => getTreeSelectOptions(tree, excludeId, excludeDescendantIds),
      [tree, excludeId, excludeDescendantIds]
    );

    const selectValue =
      value !== undefined && value !== null ? String(value) : NONE_VALUE;

    const handleChange = (v: string) => {
      onChange(v === NONE_VALUE ? undefined : Number(v));
    };

    return (
      <Select
        value={selectValue}
        onValueChange={handleChange}
        disabled={disabled}
      >
        <SelectTrigger>
          <SelectValue
            placeholder={placeholder || t('admin.selectParentCategory') || 'Select parent category'}
          />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={NONE_VALUE}>
            {t('admin.noParent') || 'No Parent (Top Level)'}
          </SelectItem>
          {options.map(option => (
            <SelectItem
              key={option.id}
              value={String(option.id)}
              disabled={option.isDisabled && option.id !== value}
            >
              <span style={{ paddingLeft: `${option.depth * 16}px` }}>
                {option.isDisabled
                  ? `${option.name} (${t('admin.disabled') || 'Disabled'})`
                  : option.name}
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }
);

TreeSelect.displayName = 'TreeSelect';
