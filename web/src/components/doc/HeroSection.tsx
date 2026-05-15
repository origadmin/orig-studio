/**
 * HeroSection - Hero area for the Doc Home page.
 * Displays site name, description, and a search input.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { useModuleState } from '@/contexts/ModuleConfigContext';
import DocSearch from '@/components/doc/DocSearch';

const HeroSection: React.FC = () => {
  const { t } = useTranslation();
  const { site } = useModuleState();

  return (
    <section className="text-center py-12 md:py-16">
      <h1 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-3">
        {site.site_name || 'OrigStudio'}
      </h1>
      {site.site_description && (
        <p className="text-lg text-muted-foreground mb-8 max-w-2xl mx-auto">
          {site.site_description}
        </p>
      )}
      <DocSearch variant="hero" />
    </section>
  );
};

export default HeroSection;
