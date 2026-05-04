/**
 * DocToc - Right-side Table of Contents for the Doc Layout.
 * Reads headings from DocTocContext and highlights the active one.
 * Only visible on xl+ screens. Hidden on article pages with no headings.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { useDocTocContext } from '@/contexts/DocTocContext';

const DocToc: React.FC = () => {
  const { t } = useTranslation();
  const { headings, activeId } = useDocTocContext();

  if (headings.length === 0) {
    return null;
  }

  const handleClick = (id: string) => {
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  };

  return (
    <aside className="hidden xl:block w-48 fixed top-12 right-0 bottom-0 overflow-y-auto py-6 px-4">
      <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
        {t('doc.tocTitle')}
      </h4>
      <nav aria-label="Table of Contents">
        <ul className="space-y-1">
          {headings.map((h) => (
            <li key={h.id}>
              <button
                onClick={() => handleClick(h.id)}
                className={`block w-full text-left text-sm py-0.5 transition-colors ${
                  h.level === 3 ? 'pl-3' : ''
                } ${
                  activeId === h.id
                    ? 'text-primary font-medium'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {h.text}
              </button>
            </li>
          ))}
        </ul>
      </nav>
    </aside>
  );
};

export default DocToc;
