/**
 * useDocToc - Hook for extracting headings from article content DOM
 * and tracking the currently visible heading via IntersectionObserver.
 *
 * Usage: Call in article page component with a ref to the content container.
 */
import { useEffect, useRef, useCallback } from 'react';
import { useDocTocContext } from '@/contexts/DocTocContext';
import type { TocHeading } from '@/types/doc';

/**
 * Slugify text for use as an HTML id attribute.
 * Handles CJK characters, alphanumeric, and hyphens.
 */
function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\u4e00-\u9fff]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export function useDocToc(contentRef: React.RefObject<HTMLElement | null>) {
  const { setHeadings, setActiveId } = useDocTocContext();
  const observerRef = useRef<IntersectionObserver | null>(null);

  // Extract headings from the content DOM
  const extractHeadings = useCallback(() => {
    if (!contentRef.current) return undefined;

    const elements = contentRef.current.querySelectorAll('h2, h3');
    const headings: TocHeading[] = [];

    elements.forEach((el) => {
      // Auto-generate id if missing
      if (!el.id) {
        el.id = slugify(el.textContent || '');
      }
      headings.push({
        id: el.id,
        text: el.textContent || '',
        level: el.tagName === 'H2' ? 2 : 3,
      });
    });

    setHeadings(headings);
    return headings;
  }, [contentRef, setHeadings]);

  // Set up IntersectionObserver to track visible headings
  useEffect(() => {
    const headings = extractHeadings();
    if (!headings || headings.length === 0) return;

    const elements = headings
      .map((h) => document.getElementById(h.id))
      .filter(Boolean) as HTMLElement[];
    if (elements.length === 0) return;

    observerRef.current = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
            break;
          }
        }
      },
      { rootMargin: '-80px 0px -70% 0px' }
    );

    elements.forEach((el) => observerRef.current?.observe(el));

    return () => {
      observerRef.current?.disconnect();
    };
  }, [extractHeadings, setActiveId]);
}
