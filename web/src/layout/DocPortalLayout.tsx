/**
 * DocPortalLayout - Three-column documentation layout.
 * Left: DocSidebar (w-64, fixed)
 * Center: Content area (flex-1, max-w-4xl)
 * Right: DocToc (w-48, fixed, xl+ only)
 *
 * Responsive:
 * - xl+: Three columns (sidebar + content + TOC)
 * - md-xl: Two columns (sidebar + content, TOC hidden)
 * - <md: Single column (content full-width, sidebar as drawer)
 */
import React, { useState, useEffect } from 'react';
import { Outlet } from '@tanstack/react-router';
import DocHeader from '@/components/doc/DocHeader';
import DocSidebar from '@/components/doc/DocSidebar';
import DocToc from '@/components/doc/DocToc';
import MobileSidebarDrawer from '@/components/doc/MobileSidebarDrawer';
import { DocTocProvider } from '@/contexts/DocTocContext';
import { useTheme } from '@/themes';

const DocPortalLayout: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isDesktop, setIsDesktop] = useState(false);
  const { isDark, toggleDark } = useTheme();

  useEffect(() => {
    const checkScreen = () => setIsDesktop(window.innerWidth >= 768);
    checkScreen();
    window.addEventListener('resize', checkScreen);
    return () => window.removeEventListener('resize', checkScreen);
  }, []);

  return (
    <DocTocProvider>
      <div className="min-h-screen bg-background transition-colors">
        <DocHeader
          onToggleSidebar={() => setSidebarOpen(!sidebarOpen)}
          darkMode={isDark}
          onToggleDarkMode={toggleDark}
        />

        <div className="pt-12 flex">
          {/* Desktop sidebar */}
          <DocSidebar />

          {/* Content area */}
          <main className="flex-1 min-w-0 md:ml-64 xl:mr-48 px-6 lg:px-8">
            <div className="max-w-4xl mx-auto py-8">
              <Outlet />
            </div>
          </main>

          {/* Right-side TOC (xl+ only) */}
          <DocToc />
        </div>

        {/* Mobile sidebar drawer */}
        {!isDesktop && (
          <MobileSidebarDrawer
            open={sidebarOpen}
            onClose={() => setSidebarOpen(false)}
          />
        )}
      </div>
    </DocTocProvider>
  );
};

export default DocPortalLayout;
