/**
 * DocHomeLayout - Documentation mode home page.
 * Displays Hero section, Quick Start grid, Recently Updated articles,
 * and full Category Navigation.
 */
import React from 'react';
import HeroSection from '@/components/doc/HeroSection';
import QuickStartGrid from '@/components/doc/QuickStartGrid';
import RecentlyUpdated from '@/components/doc/RecentlyUpdated';
import CategoryNavigation from '@/components/doc/CategoryNavigation';

const DocHomeLayout: React.FC = () => {
  return (
    <div className="max-w-4xl mx-auto">
      <HeroSection />
      <QuickStartGrid />
      <RecentlyUpdated />
      <CategoryNavigation />
    </div>
  );
};

export default DocHomeLayout;
