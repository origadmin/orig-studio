/**
 * Custom hook for managing comment tree state, expand/collapse,
 * and derived data (visible nodes, statistics).
 */
import { useState, useMemo, useCallback } from 'react';
import { adminCommentApi, type AdminComment } from '@/lib/api/comment';
import {
  buildCommentTree,
  flattenCommentTree,
  serverTreeToNodes,
  getAllExpandableIds,
  type CommentTreeNode,
  type ExpandState,
} from '@/lib/utils/commentTree';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface CommentTreeStats {
  total: number;
  approved: number;
  pending: number;
  blocked: number;
  reportedPending: number;
}

export interface UseCommentTreeReturn {
  // Raw data
  comments: AdminComment[];
  loading: boolean;
  total: number;

  // Tree data
  tree: CommentTreeNode[];
  visibleNodes: CommentTreeNode[];
  expandedIds: ExpandState;

  // Actions
  loadComments: (params?: {
    page?: number;
    page_size?: number;
    media_id?: string;
    status?: string;
    report_status?: string;
    tree?: boolean;
  }) => Promise<void>;
  toggleExpand: (id: string) => void;
  expandAll: () => void;
  collapseAll: () => void;
  expandNode: (id: string) => void;

  // Stats
  stats: CommentTreeStats;
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useCommentTree(): UseCommentTreeReturn {
  const [comments, setComments] = useState<AdminComment[]>([]);
  const [expandedIds, setExpandedIds] = useState<ExpandState>(new Set());
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [isTreeMode, setIsTreeMode] = useState(false);

  // Build tree from flat list or convert server tree (memoized)
  const tree = useMemo<CommentTreeNode[]>(() => {
    if (isTreeMode) {
      // Server already returns tree structure
      return serverTreeToNodes(comments);
    }
    return buildCommentTree(comments);
  }, [comments, isTreeMode]);

  // Flatten tree based on expand state (memoized)
  const visibleNodes = useMemo(
    () => flattenCommentTree(tree, expandedIds),
    [tree, expandedIds]
  );

  // Calculate stats from tree (memoized)
  const stats = useMemo<CommentTreeStats>(() => {
    const allNodes = flattenCommentTree(
      tree,
      new Set(getAllExpandableIds(tree))
    );
    return {
      total: allNodes.length,
      approved: allNodes.filter(n => n.status === 'approved').length,
      pending: allNodes.filter(n => n.status === 'pending').length,
      blocked: allNodes.filter(n => n.status === 'blocked').length,
      reportedPending: allNodes.filter(n => n.has_pending_reports).length,
    };
  }, [tree]);

  const toggleExpand = useCallback((id: string) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const expandNode = useCallback((id: string) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      next.add(id);
      return next;
    });
  }, []);

  const expandAll = useCallback(() => {
    setExpandedIds(new Set(getAllExpandableIds(tree)));
  }, [tree]);

  const collapseAll = useCallback(() => {
    setExpandedIds(new Set());
  }, []);

  const loadComments = useCallback(
    async (params?: {
      page?: number;
      page_size?: number;
      media_id?: string;
      status?: string;
      report_status?: string;
      tree?: boolean;
    }) => {
      setLoading(true);
      try {
        const useTree = params?.tree ?? true;
        setIsTreeMode(useTree);
        const response = await adminCommentApi.list({
          page: params?.page ?? 1,
          page_size: params?.page_size ?? 20,
          media_id: params?.media_id,
          status: params?.status,
          report_status: params?.report_status,
          tree: useTree,
        });
        const list = Array.isArray(response?.items) ? response.items : [];
        setComments(list);
        if (response?.total !== undefined) setTotal(response.total);
      } catch (error) {
        console.error('Failed to fetch comments:', error);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return {
    comments,
    loading,
    total,
    tree,
    visibleNodes,
    expandedIds,
    loadComments,
    toggleExpand,
    expandAll,
    collapseAll,
    expandNode,
    stats,
  };
}
