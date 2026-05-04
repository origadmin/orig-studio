import React from 'react';
import {Button} from '@/components/ui/button';
import {Skeleton} from '@/components/ui/skeleton';
import {useTranslation} from 'react-i18next';

interface Category {
    id: number | string;
    name: string;
}

interface CategoryFilterProps {
    categories: Category[];
    activeId: number | string | null;
    onSelect: (id: number | string | null) => void;
    className?: string;
}

const CategoryFilter: React.FC<CategoryFilterProps> = ({
    categories,
    activeId,
    onSelect,
    className = '',
}) => {
    const {t} = useTranslation();

    return (
        <div
            className={`flex gap-2 py-3 md:py-4 overflow-x-auto scrollbar-hide ${className}`}
            role="tablist"
            aria-label={t('featured.categoryFilter', 'Category filter')}
            style={{
                scrollbarWidth: 'none',
                msOverflowStyle: 'none',
            }}
        >
            <Button
                variant={activeId === null ? 'default' : 'outline'}
                size="sm"
                className="rounded-full text-sm font-medium px-4 py-1.5 shrink-0"
                role="tab"
                aria-selected={activeId === null}
                onClick={() => onSelect(null)}
            >
                {t('featured.allCategories')}
            </Button>
            {categories.map((cat) => (
                <Button
                    key={cat.id}
                    variant={activeId === cat.id ? 'default' : 'outline'}
                    size="sm"
                    className="rounded-full text-sm font-medium px-4 py-1.5 shrink-0"
                    role="tab"
                    aria-selected={activeId === cat.id}
                    onClick={() => onSelect(cat.id)}
                >
                    {cat.name}
                </Button>
            ))}
        </div>
    );
};

export const CategoryFilterSkeleton: React.FC = () => (
    <div className="flex gap-2 py-3 md:py-4 overflow-hidden">
        {Array.from({length: 6}).map((_, i) => (
            <Skeleton key={i} className="h-8 w-16 rounded-full shrink-0"/>
        ))}
    </div>
);

export default CategoryFilter;
