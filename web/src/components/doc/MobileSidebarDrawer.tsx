/**
 * MobileSidebarDrawer - Slide-out sidebar drawer for mobile devices.
 * Shows the same DocNavTree as the desktop sidebar.
 * Includes overlay backdrop that closes the drawer on click.
 */
import React from 'react';
import { X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Spinner } from '@/components/ui/spinner';
import { useDocCategoryTree } from '@/hooks/useDocNav';
import DocNavTree from './DocNavTree';

interface MobileSidebarDrawerProps {
  open: boolean;
  onClose: () => void;
}

const MobileSidebarDrawer: React.FC<MobileSidebarDrawerProps> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const { data: tree, isLoading } = useDocCategoryTree();

  if (!open) return null;

  return (
    <>
      {/* Backdrop overlay */}
      <div
        className="fixed inset-0 bg-black/50 z-40 md:hidden"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer panel */}
      <div className="fixed top-0 left-0 bottom-0 w-72 bg-white dark:bg-gray-900 z-50 md:hidden shadow-xl transform transition-transform duration-300">
        {/* Drawer header */}
        <div className="flex items-center justify-between px-4 h-12 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-sm font-semibold">{t('doc.browseByCategory')}</h2>
          <button
            onClick={onClose}
            className="p-1.5 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
            aria-label="Close sidebar"
          >
            <X size={18} />
          </button>
        </div>

        {/* Navigation tree */}
        <div className="overflow-y-auto h-[calc(100%-48px)] py-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Spinner className="h-5 w-5" />
            </div>
          ) : (
            <DocNavTree tree={tree ?? []} />
          )}
        </div>
      </div>
    </>
  );
};

export default MobileSidebarDrawer;
