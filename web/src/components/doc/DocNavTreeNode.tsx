/**
 * DocNavTreeNode - Recursive tree node for the documentation sidebar.
 * Renders a category with expand/collapse, and its child articles.
 */
import React, { useState } from 'react';
import { Link, useLocation } from '@tanstack/react-router';
import { ChevronRight, FolderOpen, Folder } from 'lucide-react';
import type { CategoryTreeNode } from '@/lib/utils/categoryTree';
import { useDocCategoryArticles } from '@/hooks/useDocNav';

interface DocNavTreeNodeProps {
  node: CategoryTreeNode;
  currentSlug?: string;
}

const DocNavTreeNode: React.FC<DocNavTreeNodeProps> = ({ node, currentSlug }) => {
  const [expanded, setExpanded] = useState(false);
  const location = useLocation();

  // Lazy-load articles when expanding
  const { data: articlesData } = useDocCategoryArticles(expanded ? node.id : undefined);
  const articles = articlesData?.items ?? [];

  const isCategoryActive = location.pathname.includes(`/categories/${node.slug}`);
  const hasChildren = node.hasChildren || articles.length > 0;

  return (
    <li role="treeitem" aria-expanded={hasChildren ? expanded : undefined}>
      {/* Category link */}
      <div
        className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm cursor-pointer transition-colors ${
          isCategoryActive
            ? 'bg-primary/10 text-primary font-medium'
            : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
        }`}
        style={{ paddingLeft: `${node.depth * 12 + 12}px` }}
      >
        {hasChildren && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
            className="shrink-0 w-4 h-4 flex items-center justify-center"
            aria-label={expanded ? 'Collapse' : 'Expand'}
          >
            <ChevronRight
              size={14}
              className={`transition-transform ${expanded ? 'rotate-90' : ''}`}
            />
          </button>
        )}
        {!hasChildren && <span className="w-4 shrink-0" />}

        <Link
          to="/categories/$slug"
          params={{ slug: node.slug }}
          className="flex items-center gap-1.5 flex-1 min-w-0"
        >
          {expanded ? (
            <FolderOpen size={14} className="shrink-0 text-muted-foreground" />
          ) : (
            <Folder size={14} className="shrink-0 text-muted-foreground" />
          )}
          <span className="truncate">{node.name}</span>
        </Link>
      </div>

      {/* Articles under this category */}
      {expanded && articles.length > 0 && (
        <ul role="group" className="space-y-0.5">
          {articles.map((article) => {
            const isArticleActive = currentSlug === article.slug;
            return (
              <li key={article.id} role="treeitem">
                <Link
                  to="/articles/$slug"
                  params={{ slug: article.slug }}
                  className={`block px-3 py-1 rounded-md text-sm transition-colors ${
                    isArticleActive
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
                  }`}
                  style={{ paddingLeft: `${(node.depth + 1) * 12 + 12}px` }}
                >
                  {article.title}
                </Link>
              </li>
            );
          })}
        </ul>
      )}

      {/* Child categories */}
      {expanded && node.children.length > 0 && (
        <ul role="group" className="space-y-0.5">
          {node.children.map((child) => (
            <DocNavTreeNode
              key={child.id}
              node={child}
              currentSlug={currentSlug}
            />
          ))}
        </ul>
      )}
    </li>
  );
};

export default DocNavTreeNode;
