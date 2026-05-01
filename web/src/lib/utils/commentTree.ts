/**
 * Comment tree utility functions for building, flattening, and querying
 * hierarchical comment structures.
 *
 * All functions are pure (no side effects) and can be unit-tested independently.
 */
import type { AdminComment } from '@/lib/api/comment';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/**
 * Represents a comment node in the tree structure.
 * Extends AdminComment with tree metadata (children, depth, descendant count).
 */
export interface CommentTreeNode extends AdminComment {
  /** Direct children of this node */
  children: CommentTreeNode[];

  /** Depth in the tree (0 = top-level, 1 = first reply, etc.) */
  depth: number;

  /** Whether this node has at least one child */
  hasReplies: boolean;

  /** Total number of descendants (recursive) */
  descendantCount: number;
}

/**
 * Tracks which tree nodes are expanded.
 * Uses Set for O(1) lookup.
 */
export type ExpandState = Set<string>;

// ---------------------------------------------------------------------------
// buildCommentTree
// ---------------------------------------------------------------------------

/**
 * Builds a tree structure from a flat list of comments.
 * Time complexity: O(n).
 * Space complexity: O(n).
 *
 * @param flatList - Flat array of AdminComment from API
 * @returns Array of root-level CommentTreeNode (with children nested)
 */
export function buildCommentTree(flatList: AdminComment[]): CommentTreeNode[] {
  if (!flatList || flatList.length === 0) return [];

  // Step 1: Create node map
  const nodeMap = new Map<string, CommentTreeNode>();
  for (const c of flatList) {
    nodeMap.set(c.id, {
      ...c,
      children: [],
      depth: 0,
      hasReplies: false,
      descendantCount: 0,
    });
  }

  // Step 2: Build parent-child links
  const rootNodes: CommentTreeNode[] = [];

  for (const node of nodeMap.values()) {
    const parentId = node.parent_id;

    // No parent, or parent not found -> root node
    if (!parentId || !nodeMap.has(parentId)) {
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
    parent.hasReplies = true;
  }

  // Step 3: Sort children by create_time
  function sortChildren(nodes: CommentTreeNode[]): void {
    nodes.sort((a, b) => {
      const timeA = a.create_time ? new Date(a.create_time).getTime() : 0;
      const timeB = b.create_time ? new Date(b.create_time).getTime() : 0;
      return timeA - timeB;
    });
    for (const node of nodes) {
      sortChildren(node.children);
    }
  }
  sortChildren(rootNodes);

  // Step 4: Calculate depth
  function calculateDepth(nodes: CommentTreeNode[], parentDepth: number): void {
    for (const node of nodes) {
      node.depth = parentDepth;
      calculateDepth(node.children, parentDepth + 1);
    }
  }
  calculateDepth(rootNodes, 0);

  // Step 5: Calculate descendantCount (post-order)
  function calculateDescendantCount(nodes: CommentTreeNode[]): void {
    for (const node of nodes) {
      calculateDescendantCount(node.children);
      node.descendantCount = node.children.reduce(
        (sum, child) => sum + 1 + child.descendantCount,
        0
      );
    }
  }
  calculateDescendantCount(rootNodes);

  return rootNodes;
}

// ---------------------------------------------------------------------------
// flattenCommentTree
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
 * @returns Flat array of CommentTreeNode in display order
 */
export function flattenCommentTree(
  tree: CommentTreeNode[],
  expandedIds: ExpandState
): CommentTreeNode[] {
  const result: CommentTreeNode[] = [];

  function dfs(nodes: CommentTreeNode[]): void {
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
// serverTreeToNodes
// ---------------------------------------------------------------------------

/**
 * Converts server tree response (with nested children) to CommentTreeNode[].
 * The server returns pre-built tree structure when tree=true is used.
 * This function adds depth, hasReplies, and descendantCount metadata.
 *
 * @param serverData - Array of comments from server with nested children
 * @returns Array of CommentTreeNode with tree metadata
 */
export function serverTreeToNodes(serverData: AdminComment[]): CommentTreeNode[] {
  if (!serverData || serverData.length === 0) return [];

  function convertNode(item: AdminComment, depth: number): CommentTreeNode {
    const children = (item.children || []).map(child => convertNode(child, depth + 1));

    return {
      ...item,
      children,
      depth,
      hasReplies: children.length > 0,
      descendantCount: children.reduce(
        (sum, child) => sum + 1 + child.descendantCount,
        0
      ),
    };
  }

  return serverData.map(item => convertNode(item, 0));
}

// ---------------------------------------------------------------------------
// getAllExpandableIds
// ---------------------------------------------------------------------------

/**
 * Gets all IDs of nodes that have children (expandable).
 * Used for "Expand All" functionality.
 *
 * @param tree - Root-level tree nodes
 * @returns Array of expandable node IDs
 */
export function getAllExpandableIds(tree: CommentTreeNode[]): string[] {
  const result: string[] = [];

  function dfs(nodes: CommentTreeNode[]): void {
    for (const node of nodes) {
      if (node.children.length > 0) {
        result.push(node.id);
      }
      dfs(node.children);
    }
  }

  dfs(tree);
  return result;
}
