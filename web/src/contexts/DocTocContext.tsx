/**
 * DocTocContext - Provides TOC (Table of Contents) data for the Doc Layout.
 * Article pages set headings via setHeadings(), and DocToc reads them for display.
 * IntersectionObserver updates activeId as the user scrolls.
 */
import React, { createContext, useContext, useState, useCallback } from 'react';
import type { DocTocContextValue, TocHeading } from '@/types/doc';

const DocTocContext = createContext<DocTocContextValue>({
  headings: [],
  activeId: '',
  setHeadings: () => {},
  setActiveId: () => {},
});

export function DocTocProvider({ children }: { children: React.ReactNode }) {
  const [headings, setHeadingsState] = useState<TocHeading[]>([]);
  const [activeId, setActiveId] = useState('');

  const setHeadings = useCallback((newHeadings: TocHeading[]) => {
    setHeadingsState(newHeadings);
  }, []);

  const setActiveIdCallback = useCallback((id: string) => {
    setActiveId(id);
  }, []);

  return (
    <DocTocContext.Provider value={{ headings, activeId, setHeadings, setActiveId: setActiveIdCallback }}>
      {children}
    </DocTocContext.Provider>
  );
}

export function useDocTocContext(): DocTocContextValue {
  return useContext(DocTocContext);
}
