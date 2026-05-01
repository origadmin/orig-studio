/**
 * Unit tests for category tree utility functions.
 * Covers: buildCategoryTree, flattenCategoryTree, getTreeSelectOptions, getDescendantIds
 */
import {
  buildCategoryTree,
  flattenCategoryTree,
  getTreeSelectOptions,
  getDescendantIds,
  type CategoryTreeNode,
} from '@/lib/utils/categoryTree';
import type { Category } from '@/lib/api/category';

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

function makeCategory(overrides: Partial<Category> & { id: number }): Category {
  return {
    name: `Category ${overrides.id}`,
    slug: `cat-${overrides.id}`,
    description: '',
    parent_id: 0,
    order: 0,
    status: 1,
    media_count: 0,
    create_time: '2026-01-01T00:00:00Z',
    update_time: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// buildCategoryTree
// ---------------------------------------------------------------------------

describe('buildCategoryTree', () => {
  it('returns empty array for empty list', () => {
    const result = buildCategoryTree([]);
    expect(result).toEqual([]);
  });

  it('returns single root node with no children', () => {
    const flatList = [makeCategory({ id: 1 })];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
    expect(result[0].children).toHaveLength(0);
    expect(result[0].depth).toBe(0);
    expect(result[0].hasChildren).toBe(false);
    expect(result[0].descendantCount).toBe(0);
    expect(result[0].isAncestorDisabled).toBe(false);
    expect(result[0].ancestorIds).toEqual([]);
  });

  it('returns multiple root nodes sorted by order', () => {
    const flatList = [
      makeCategory({ id: 1, order: 2 }),
      makeCategory({ id: 2, order: 1 }),
      makeCategory({ id: 3, order: 3 }),
    ];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(3);
    expect(result[0].id).toBe(2); // order 1
    expect(result[1].id).toBe(1); // order 2
    expect(result[2].id).toBe(3); // order 3
  });

  it('builds two-level tree with correct depth', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0 }),
      makeCategory({ id: 2, parent_id: 1, order: 1 }),
      makeCategory({ id: 3, parent_id: 1, order: 2 }),
    ];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
    expect(result[0].children).toHaveLength(2);
    expect(result[0].hasChildren).toBe(true);
    expect(result[0].descendantCount).toBe(2);

    expect(result[0].children[0].id).toBe(2);
    expect(result[0].children[0].depth).toBe(1);
    expect(result[0].children[0].ancestorIds).toEqual([1]);

    expect(result[0].children[1].id).toBe(3);
    expect(result[0].children[1].depth).toBe(1);
  });

  it('builds three-level tree with correct depth', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0 }),
      makeCategory({ id: 2, parent_id: 1 }),
      makeCategory({ id: 3, parent_id: 2 }),
    ];
    const result = buildCategoryTree(flatList);

    expect(result[0].depth).toBe(0);
    expect(result[0].children[0].depth).toBe(1);
    expect(result[0].children[0].children[0].depth).toBe(2);
    expect(result[0].children[0].children[0].ancestorIds).toEqual([1, 2]);
  });

  it('treats orphan category (non-existent parent) as root', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 999 }), // parent 999 does not exist
    ];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
    expect(result[0].depth).toBe(0);
  });

  it('handles self-referencing category (circular to self) as root', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 1 }), // self-referencing
    ];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
    expect(result[0].depth).toBe(0);
  });

  it('sorts children by order field ascending', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0 }),
      makeCategory({ id: 2, parent_id: 1, order: 3 }),
      makeCategory({ id: 3, parent_id: 1, order: 1 }),
      makeCategory({ id: 4, parent_id: 1, order: 2 }),
    ];
    const result = buildCategoryTree(flatList);

    const childrenIds = result[0].children.map(c => c.id);
    expect(childrenIds).toEqual([3, 4, 2]); // sorted by order: 1, 2, 3
  });

  it('calculates descendantCount correctly for multi-level tree', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0 }),
      makeCategory({ id: 2, parent_id: 1 }),
      makeCategory({ id: 3, parent_id: 1 }),
      makeCategory({ id: 4, parent_id: 2 }),
      makeCategory({ id: 5, parent_id: 2 }),
    ];
    const result = buildCategoryTree(flatList);

    // Root has 2 direct children + 2 grandchildren = 4 descendants
    expect(result[0].descendantCount).toBe(4);
    // Node 2 has 2 direct children = 2 descendants
    expect(result[0].children[0].descendantCount).toBe(2);
    // Leaf nodes have 0 descendants
    expect(result[0].children[0].children[0].descendantCount).toBe(0);
  });

  it('marks isAncestorDisabled=true when any ancestor has status !== 1', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0, status: 2 }), // disabled parent
      makeCategory({ id: 2, parent_id: 1, status: 1 }), // enabled child
      makeCategory({ id: 3, parent_id: 2, status: 1 }), // grandchild
    ];
    const result = buildCategoryTree(flatList);

    // Root is disabled but isAncestorDisabled=false (no ancestors)
    expect(result[0].isAncestorDisabled).toBe(false);
    // Child has disabled ancestor
    expect(result[0].children[0].isAncestorDisabled).toBe(true);
    // Grandchild also has disabled ancestor (recursive)
    expect(result[0].children[0].children[0].isAncestorDisabled).toBe(true);
  });

  it('marks isAncestorDisabled=false when all ancestors are enabled', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: 0, status: 1 }),
      makeCategory({ id: 2, parent_id: 1, status: 1 }),
    ];
    const result = buildCategoryTree(flatList);

    expect(result[0].isAncestorDisabled).toBe(false);
    expect(result[0].children[0].isAncestorDisabled).toBe(false);
  });

  it('handles undefined parent_id as root', () => {
    const flatList = [
      makeCategory({ id: 1, parent_id: undefined as unknown as number }),
    ];
    const result = buildCategoryTree(flatList);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
    expect(result[0].depth).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// flattenCategoryTree
// ---------------------------------------------------------------------------

describe('flattenCategoryTree', () => {
  function buildSampleTree(): CategoryTreeNode[] {
    return buildCategoryTree([
      makeCategory({ id: 1, parent_id: 0, order: 1 }),
      makeCategory({ id: 2, parent_id: 1, order: 1 }),
      makeCategory({ id: 3, parent_id: 1, order: 2 }),
      makeCategory({ id: 4, parent_id: 2, order: 1 }),
      makeCategory({ id: 5, parent_id: 0, order: 2 }),
    ]);
  }

  it('returns only root nodes when all collapsed', () => {
    const tree = buildSampleTree();
    const result = flattenCategoryTree(tree, new Set());

    expect(result.map(n => n.id)).toEqual([1, 5]);
  });

  it('returns root + direct children when one node expanded', () => {
    const tree = buildSampleTree();
    const result = flattenCategoryTree(tree, new Set([1]));

    expect(result.map(n => n.id)).toEqual([1, 2, 3, 5]);
  });

  it('returns all nodes when all expanded', () => {
    const tree = buildSampleTree();
    const result = flattenCategoryTree(tree, new Set([1, 2]));

    expect(result.map(n => n.id)).toEqual([1, 2, 4, 3, 5]);
  });

  it('returns empty array for empty tree', () => {
    const result = flattenCategoryTree([], new Set());
    expect(result).toEqual([]);
  });

  it('collapsing parent hides all descendants', () => {
    const tree = buildSampleTree();
    // Expand 1 and 2, then collapse 1
    const expanded = new Set<number>();
    expanded.add(1);
    expanded.add(2);
    const fullResult = flattenCategoryTree(tree, expanded);
    expect(fullResult.map(n => n.id)).toEqual([1, 2, 4, 3, 5]);

    // Now collapse node 1
    expanded.delete(1);
    const collapsedResult = flattenCategoryTree(tree, expanded);
    expect(collapsedResult.map(n => n.id)).toEqual([1, 5]);
  });
});

// ---------------------------------------------------------------------------
// getTreeSelectOptions
// ---------------------------------------------------------------------------

describe('getTreeSelectOptions', () => {
  function buildSampleTree(): CategoryTreeNode[] {
    return buildCategoryTree([
      makeCategory({ id: 1, parent_id: 0, status: 1 }),
      makeCategory({ id: 2, parent_id: 1, status: 1 }),
      makeCategory({ id: 3, parent_id: 1, status: 2 }), // disabled
      makeCategory({ id: 4, parent_id: 0, status: 1 }),
    ]);
  }

  it('returns all categories when no exclusion', () => {
    const tree = buildSampleTree();
    const options = getTreeSelectOptions(tree);

    expect(options.map(o => o.id)).toEqual([1, 2, 3, 4]);
  });

  it('excludes specified category ID and its children (for circular prevention)', () => {
    const tree = buildSampleTree();
    // When excluding node 1, its children (2, 3) should also be excluded
    // because they are descendants of the excluded node
    const descendantsOf1 = getDescendantIds(tree, 1);
    const options = getTreeSelectOptions(tree, 1, descendantsOf1);

    expect(options.map(o => o.id)).toEqual([4]);
  });

  it('excludes only specified category ID without descendants when no descendantIds provided', () => {
    const tree = buildSampleTree();
    // When only excluding node 1 (no descendant exclusion), children are still
    // excluded because they are under the excluded node in the tree traversal
    const options = getTreeSelectOptions(tree, 1);

    expect(options.map(o => o.id)).toEqual([4]);
  });

  it('marks disabled categories correctly', () => {
    const tree = buildSampleTree();
    const options = getTreeSelectOptions(tree);

    const disabledOption = options.find(o => o.id === 3);
    expect(disabledOption?.isDisabled).toBe(true);

    const enabledOption = options.find(o => o.id === 1);
    expect(enabledOption?.isDisabled).toBe(false);
  });

  it('preserves depth information', () => {
    const tree = buildSampleTree();
    const options = getTreeSelectOptions(tree);

    expect(options.find(o => o.id === 1)?.depth).toBe(0);
    expect(options.find(o => o.id === 2)?.depth).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// getDescendantIds
// ---------------------------------------------------------------------------

describe('getDescendantIds', () => {
  function buildSampleTree(): CategoryTreeNode[] {
    return buildCategoryTree([
      makeCategory({ id: 1, parent_id: 0 }),
      makeCategory({ id: 2, parent_id: 1 }),
      makeCategory({ id: 3, parent_id: 1 }),
      makeCategory({ id: 4, parent_id: 2 }),
      makeCategory({ id: 5, parent_id: 0 }),
    ]);
  }

  it('returns all descendant IDs for a node', () => {
    const tree = buildSampleTree();
    const descendants = getDescendantIds(tree, 1);

    expect(descendants).toEqual([2, 4, 3]);
  });

  it('returns empty array for a leaf node', () => {
    const tree = buildSampleTree();
    const descendants = getDescendantIds(tree, 4);

    expect(descendants).toEqual([]);
  });

  it('returns empty array for non-existent ID', () => {
    const tree = buildSampleTree();
    const descendants = getDescendantIds(tree, 999);

    expect(descendants).toEqual([]);
  });

  it('returns only direct children for a node with no grandchildren', () => {
    const tree = buildSampleTree();
    const descendants = getDescendantIds(tree, 2);

    expect(descendants).toEqual([4]);
  });
});
