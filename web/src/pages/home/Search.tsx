import React, {useState} from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {Filter, Eye, Play, Loader2, Search as SearchIcon} from 'lucide-react';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {useMediaList} from '@/hooks/queries';
import {getFullUrl} from '@/lib/utils';
import {Input} from '@/components/ui/input';
import {Button} from '@/components/ui/button';
import {useCategoryList} from '@/hooks/queries';

const SearchPage = () => {
    const {t} = useTranslation();
    const location = useLocation();
    const searchParams = new URLSearchParams(location.search);
    const initialQuery = searchParams.get('q') || '';
    const initialCategoryId = searchParams.get('category_id') || '';

    const [searchQuery, setSearchQuery] = useState(initialQuery);
    const [inputValue, setInputValue] = useState(initialQuery);
    const [selectedCategory, setSelectedCategory] = useState<string>(initialCategoryId);
    const [page, setPage] = useState(1);
    const [showFilters, setShowFilters] = useState(false);
    const pageSize = 10;

    const {data: categories} = useCategoryList();

    const {data, isLoading, error} = useMediaList({
        page,
        page_size: pageSize,
        status: 'active',
        keyword: searchQuery,
        category_id: selectedCategory ? Number(selectedCategory) : undefined
    });

    const searchResults = data?.items || [];
    const totalResults = data?.total || 0;

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        setSearchQuery(inputValue);
        setPage(1);
        // Update URL without reloading
        const params = new URLSearchParams();
        if (inputValue) params.set('q', inputValue);
        if (selectedCategory) params.set('category_id', selectedCategory);
        window.history.replaceState({}, '', `/search?${params.toString()}`);
    };

    const handleCategoryChange = (categoryId: string) => {
        setSelectedCategory(categoryId);
        setPage(1);
        const params = new URLSearchParams();
        if (searchQuery) params.set('q', searchQuery);
        if (categoryId) params.set('category_id', categoryId);
        window.history.replaceState({}, '', `/search?${params.toString()}`);
    };

    const clearFilters = () => {
        setSearchQuery('');
        setInputValue('');
        setSelectedCategory('');
        setPage(1);
        window.history.replaceState({}, '', '/search');
    };

    return (
        <div className="space-y-8">
            {/* Search Header */}
            <div className="pb-6 border-b border-slate-100 dark:border-gray-700">
                <form onSubmit={handleSearch} className="flex gap-3">
                    <div className="relative flex-1">
                        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"/>
                        <Input
                            type="text"
                            placeholder={t('search.placeholder') || 'Search videos...'}
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            className="pl-10 h-12 text-base"
                        />
                    </div>
                    <Button type="submit" className="h-12 px-6 bg-emerald-600 hover:bg-emerald-700">
                        <SearchIcon className="w-5 h-5 mr-2"/>
                        {t('search.search') || 'Search'}
                    </Button>
                    <Button
                        type="button"
                        variant="outline"
                        className="h-12 px-4"
                        onClick={() => setShowFilters(!showFilters)}
                    >
                        <Filter size={18} className={showFilters ? 'text-emerald-600' : ''}/>
                    </Button>
                </form>

                {/* Filters */}
                {showFilters && (
                    <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-xl">
                        <div className="flex items-center justify-between mb-3">
                            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                {t('search.filterByCategory') || 'Filter by Category'}
                            </span>
                            {(selectedCategory || searchQuery) && (
                                <button
                                    onClick={clearFilters}
                                    className="text-sm text-emerald-600 hover:text-emerald-700"
                                >
                                    {t('search.clearFilters') || 'Clear Filters'}
                                </button>
                            )}
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <button
                                onClick={() => handleCategoryChange('')}
                                className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
                                    !selectedCategory
                                        ? 'bg-emerald-600 text-white'
                                        : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600'
                                }`}
                            >
                                {t('search.allCategories') || 'All Categories'}
                            </button>
                            {categories?.items?.map((category: any) => (
                                <button
                                    key={category.id}
                                    onClick={() => handleCategoryChange(String(category.id))}
                                    className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
                                        selectedCategory === String(category.id)
                                            ? 'bg-emerald-600 text-white'
                                            : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600'
                                    }`}
                                >
                                    {category.name}
                                </button>
                            ))}
                        </div>
                    </div>
                )}
            </div>

            {/* Results Header */}
            {searchQuery && (
                <div className="flex items-center justify-between">
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
                        {t('search.resultsFor', {query: searchQuery})}
                    </h1>
                    <span className="text-sm text-gray-500">
                        {totalResults} {t('search.results') || 'results'}
                    </span>
                </div>
            )}

            {/* Loading State */}
            {isLoading && (
                <div className="flex items-center justify-center min-h-[300px]">
                    <div
                        className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
                </div>
            )}

            {/* Error State */}
            {!isLoading && error && (
                <div className="py-20 text-center space-y-4">
                    <div className="text-gray-500 text-lg">{t('common.error') || 'Error loading results'}</div>
                    <p className="text-sm text-gray-400">{(error as Error).message}</p>
                    <Link to="/">
                        <button
                            className="flex items-center space-x-2 px-6 py-2.5 bg-slate-900 dark:bg-gray-800 text-white rounded-2xl text-xs font-black hover:bg-emerald-600 transition-all mx-auto">
                            <span>{t('common.backToHome')}</span>
                        </button>
                    </Link>
                </div>
            )}

            {/* Empty State */}
            {!isLoading && !error && searchResults.length === 0 && (
                <div className="py-20 text-center space-y-4">
                    <div className="text-gray-500 text-lg">
                        {searchQuery
                            ? t('search.noResults', {query: searchQuery})
                            : t('search.enterQuery') || 'Enter a search term to find videos'}
                    </div>
                    {searchQuery && (
                        <Link to="/">
                            <button
                                className="flex items-center space-x-2 px-6 py-2.5 bg-slate-900 dark:bg-gray-800 text-white rounded-2xl text-xs font-black hover:bg-emerald-600 transition-all mx-auto">
                                <span>{t('common.backToHome')}</span>
                            </button>
                        </Link>
                    )}
                </div>
            )}

            {/* Results List */}
            {!isLoading && !error && searchResults.length > 0 && (
                <>
                    <div className="space-y-6">
                        {searchResults.map(item => (
                            <Link key={item.id} to="/watch" search={{v: item.short_token}}
                                  className="flex flex-col md:flex-row gap-6 group p-4 rounded-2xl hover:bg-slate-50 dark:hover:bg-gray-800 transition-all">
                                <div
                                    className="relative w-full md:w-72 aspect-video bg-slate-200 rounded-xl overflow-hidden shrink-0 border border-slate-100 dark:border-gray-700 shadow-sm">
                                    <img src={item.thumbnail ? getFullUrl(item.thumbnail) : undefined}
                                         className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                                         alt={item.title}/>
                                    <div
                                        className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">
                                        {formatDuration(item.duration)}
                                    </div>
                                    <div
                                        className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                        <div
                                            className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                                            <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex-1 space-y-2 min-w-0">
                                    <h3 className="text-lg font-bold text-slate-900 dark:text-white group-hover:text-emerald-600 transition-colors line-clamp-2 leading-tight">
                                        {item.title}
                                    </h3>
                                    <div
                                        className="flex items-center space-x-3 text-xs font-medium text-slate-500 dark:text-gray-400">
                                        <span>{item.edges?.user?.[0]?.username || 'Unknown'}</span>
                                        <span>·</span>
                                        <span className="flex items-center gap-1">
                                            <Eye size={12}/>
                                            {formatViews(item.view_count)} {t('common.views')}
                                        </span>
                                        <span>·</span>
                                        <span>{formatDate(item.created_at)}</span>
                                    </div>
                                    <p className="text-sm text-slate-500 dark:text-gray-400 line-clamp-2 leading-relaxed">
                                        {item.description || t('watch.noDescription')}
                                    </p>
                                    {item.tags && item.tags.length > 0 && (
                                        <div className="flex flex-wrap gap-1 pt-1">
                                            {item.tags.slice(0, 3).map(tag => (
                                                <span key={tag}
                                                      className="text-xs text-emerald-600 dark:text-emerald-400">
                                                    #{tag}
                                                </span>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </Link>
                        ))}
                    </div>

                    {/* Pagination */}
                    {totalResults > pageSize && (
                        <div className="flex justify-center pt-8">
                            <div className="flex gap-2">
                                <button
                                    onClick={() => setPage(p => Math.max(1, p - 1))}
                                    disabled={page === 1}
                                    className="px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {t('common.previous') || 'Previous'}
                                </button>
                                {Array.from({length: Math.min(5, Math.ceil(totalResults / pageSize))}, (_, i) => {
                                    const pageNum = i + 1;
                                    return (
                                        <button
                                            key={pageNum}
                                            onClick={() => setPage(pageNum)}
                                            className={`px-4 py-2 rounded-lg ${
                                                page === pageNum
                                                    ? 'bg-emerald-600 text-white'
                                                    : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
                                            }`}
                                        >
                                            {pageNum}
                                        </button>
                                    );
                                })}
                                <button
                                    onClick={() => setPage(p => p + 1)}
                                    disabled={page >= Math.ceil(totalResults / pageSize)}
                                    className="px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {t('common.next') || 'Next'}
                                </button>
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};
export default SearchPage;
