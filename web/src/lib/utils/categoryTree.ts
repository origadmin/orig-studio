/**
 * Category tree utility functions for building, flattening, and querying
 * hierarchical category structures.
 *
 * All functions are pure (no side effects) and can be unit-tested independently.
 */
import type { Category } from '@/lib/api/category';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/**
 * Represents a category node in the tree structure.
 * Extends Category with tree metadata (children, depth, parent status).
 */
export interface CategoryTreeNode extends Category {
  /** Direct children of this node, sorted by `order` */
  children: CategoryTreeNode[];

  /** Depth in the tree (0 = top-level, 1 = first child, etc.) */
  depth: number;

  /** Whether this node has at least one child */
  hasChildren: boolean;

  /** Total number of descendants (recursive) */
  descendantCount: number;

  /** Whether any ancestor is disabled (status !== 1) */
  isAncestorDisabled: boolean;

  /** The IDs of all ancestors from root to this node (exclusive) */
  ancestorIds: number[];
}

/**
 * Tracks which tree nodes are expanded.
 * Uses Set for O(1) lookup.
 */
export type ExpandState = Set<number>;

/**
 * Option for the tree-based parent category selector.
 * Includes depth for visual indentation in the dropdown.
 */
export interface TreeSelectOption {
  id: number;
  name: string;
  slug: string;
  depth: number;
  isDisabled: boolean;
  hasChildren: boolean;
}

// ---------------------------------------------------------------------------
// buildCategoryTree
// ---------------------------------------------------------------------------

/**
 * Builds a tree structure from a flat list of categories.
 * Time complexity: O(n log n) (dominated by sorting).
 * Space complexity: O(n).
 *
 * @param flatList - Flat array of Category from API
 * @returns Array of root-level CategoryTreeNode (with children nested)
 */
export function buildCategoryTree(flatList: Category[]): CategoryTreeNode[] {
  if (!flatList || flatList.length === 0) return [];

  // Step 1: Create node map
  const nodeMap = new Map<number, CategoryTreeNode>();
  for (const cat of flatList) {
    nodeMap.set(cat.id, {
      ...cat,
      children: [],
      depth: 0,
      hasChildren: false,
      descendantCount: 0,
      isAncestorDisabled: false,
      ancestorIds: [],
    });
  }

  // Step 2: Build parent-child links
  const rootNodes: CategoryTreeNode[] = [];
  const visited = new Set<number>();

  for (const node of nodeMap.values()) {
    const parentId = node.parent_id;

    // No parent, parent_id is 0, or parent not found -> root node
    if (!parentId || parentId === 0 || !nodeMap.has(parentId)) {
      rootNodes.push(node);
      continue;
    }

    // Circular reference detection
    if (parentId === node.id) {
      rootNodes.push(node);
      continue;
    }

    const parent = nodeMap.get(parentId)!;
    parent.children.push(node);
    parent.hasChildren = true;
  }

  // Step 3: Sort children by `order` field
  function sortChildren(nodes: CategoryTreeNode[]): void {
    nodes.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
    for (const node of nodes) {
      sortChildren(node.children);
    }
  }
  sortChildren(rootNodes);

  // Step 4: Calculate depth and ancestorIds
  function calculateDepthAndAncestors(nodes: CategoryTreeNode[], parentDepth: number, parentAncestorIds: number[]): void {
    for (const node of nodes) {
      node.depth = parentDepth;
      node.ancestorIds = [...parentAncestorIds];
      calculateDepthAndAncestors(node.children, parentDepth + 1, [...parentAncestorIds, node.id]);
    }
  }
  calculateDepthAndAncestors(rootNodes, 0, []);

  // Step 5: Calculate descendantCount (post-order)
  function calculateDescendantCount(nodes: CategoryTreeNode[]): void {
    for (const node of nodes) {
      calculateDescendantCount(node.children);
      node.descendantCount = node.children.reduce(
        (sum, child) => sum + 1 + child.descendantCount,
        0
      );
    }
  }
  calculateDescendantCount(rootNodes);

  // Step 6: Calculate isAncestorDisabled (pre-order)
  function markAncestorDisabled(nodes: CategoryTreeNode[], parentDisabled: boolean): void {
    for (const node of nodes) {
      node.isAncestorDisabled = parentDisabled;
      const thisNodeDisabled = parentDisabled || node.status !== 1;
      markAncestorDisabled(node.children, thisNodeDisabled);
    }
  }
  markAncestorDisabled(rootNodes, false);

  return rootNodes;
}

// ---------------------------------------------------------------------------
// flattenCategoryTree
// ---------------------------------------------------------------------------

/**
 * Flattens the tree into a renderable array based on expand state.
 * Only includes nodes that should be visible:
 * - All root nodes (always visible)
 * - Direct children of expanded nodes
 * - Does NOT include children of collapsed nodes
 *
 * @param tree - Root-level tree nodes
 * @param expandedIds - Set of currently expanded node IDs
 * @returns Flat array of CategoryTreeNode in display order
 */
export function flattenCategoryTree(
  tree: CategoryTreeNode[],
  expandedIds: ExpandState
): CategoryTreeNode[] {
  const result: CategoryTreeNode[] = [];

  function dfs(nodes: CategoryTreeNode[]): void {
    for (const node of nodes) {
      result.push(node);
      if (expandedIds.has(node.id)) {
        dfs(node.children);
      }
    }
  }

  dfs(tree);
  return result;
}

// ---------------------------------------------------------------------------
// getTreeSelectOptions
// ---------------------------------------------------------------------------

/**
 * Generates options for the tree-based parent category selector.
 * Returns a flat array with depth info for visual indentation.
 *
 * @param tree - Root-level tree nodes
 * @param excludeId - Category ID to exclude (current category in edit mode)
 * @param excludeDescendantIds - IDs of descendants to exclude (prevent circular)
 * @returns Flat array of TreeSelectOption in tree display order
 */
export function getTreeSelectOptions(
  tree: CategoryTreeNode[],
  excludeId?: number,
  excludeDescendantIds?: number[]
): TreeSelectOption[] {
  const result: TreeSelectOption[] = [];
  const excludeSet = new Set<number>();
  if (excludeId !== undefined) excludeSet.add(excludeId);
  if (excludeDescendantIds) {
    for (const id of excludeDescendantIds) excludeSet.add(id);
  }

  function dfs(nodes: CategoryTreeNode[]): void {
    for (const node of nodes) {
      if (excludeSet.has(node.id)) continue;
      result.push({
        id: node.id,
        name: node.name,
        slug: node.slug,
        depth: node.depth,
        isDisabled: node.status !== 1,
        hasChildren: node.hasChildren,
      });
      dfs(node.children);
    }
  }

  dfs(tree);
  return result;
}

// ---------------------------------------------------------------------------
// getDescendantIds
// ---------------------------------------------------------------------------

/**
 * Gets all descendant IDs of a given category.
 * Used to prevent circular parent assignment in edit dialog.
 *
 * @param tree - Root-level tree nodes
 * @param categoryId - The category whose descendants to find
 * @returns Array of descendant category IDs
 */
export function getDescendantIds(
  tree: CategoryTreeNode[],
  categoryId: number
): number[] {
  const result: number[] = [];

  function findAndCollect(nodes: CategoryTreeNode[]): boolean {
    for (const node of nodes) {
      if (node.id === categoryId) {
        collectDescendants(node);
        return true;
      }
      if (findAndCollect(node.children)) {
        return true;
      }
    }
    return false;
  }

  function collectDescendants(node: CategoryTreeNode): void {
    for (const child of node.children) {
      result.push(child.id);
      collectDescendants(child);
    }
  }

  findAndCollect(tree);
  return result;
}
