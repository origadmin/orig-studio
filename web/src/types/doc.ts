/**
 * Type definitions for the Doc Layout feature (F026).
 * Includes TOC headings, DocToc context value, and navigation tree node types.
 */

/** TOC heading item extracted from article content */
export interface TocHeading {
  /** Heading id (used for anchor navigation) */
  id: string;
  /** Heading text content */
  text: string;
  /** Heading level (2 = h2, 3 = h3) */
  level: 2 | 3;
}

/** DocToc Context value */
export interface DocTocContextValue {
  /** Headings extracted from the current article */
  headings: TocHeading[];
  /** Currently visible heading id (highlighted in TOC) */
  activeId: string;
  /** Set headings (called by article page on mount) */
  setHeadings: (headings: TocHeading[]) => void;
  /** Set active heading id (called by IntersectionObserver) */
  setActiveId: (id: string) => void;
}

/** Article summary shown in DocSidebar navigation */
export interface DocNavArticle {
  id: string;
  title: string;
  slug: string;
}
