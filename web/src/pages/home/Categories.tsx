import {Spinner} from "@/components/ui/spinner"

import React, {useState, useEffect, useMemo} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, ChevronDown, ChevronRight, Folder} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {categoryApi, type Category} from '@/lib/api/category';
import {useMediaList} from '@/hooks/queries';
import {getFullUrl} from '@/lib/utils';
import {buildCategoryTree, type CategoryTreeNode} from '@/lib/utils/categoryTree';

const VideoCard: React.FC<{media: any}> = ({media}) => (
    <Link to="/watch" search={{v: media.short_token}} className="group">
        <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
            <div className="relative aspect-video overflow-hidden">
                <img
                    src={media.thumbnail ? getFullUrl(media.thumbnail) : undefined}
                    alt={media.title}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                />
                <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                    {formatDuration(media.duration)}
                </div>
                <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <div className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                        <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                    </div>
                </div>
            </div>
            <div className="p-3">
                <h3 className="font-medium text-gray-900 dark:text-white text-sm line-clamp-2 mb-1.5 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                    {media.title}
                </h3>
                <div className="flex items-center gap-2 mb-1">
                    <img
                        src={media.edges?.user?.[0]?.avatar ? getFullUrl(media.edges.user[0].avatar) : undefined}
                        alt={media.edges?.user?.[0]?.username}
                        className="w-5 h-5 rounded-full object-cover"
                    />
                    <span className="text-xs text-gray-500 dark:text-muted-foreground">
                        {media.edges?.user?.[0]?.username || 'Unknown'}
                    </span>
                </div>
                <div className="flex items-center gap-3 text-xs text-muted-foreground dark:text-gray-500">
                    <span className="flex items-center gap-1">
                        <Eye size={12}/>
                        {formatViews(media.view_count)}
                    </span>
                </div>
            </div>
        </div>
    </Link>
);

const CategoryFilter: React.FC<{
    tree: CategoryTreeNode[];
    selectedId: number | null;
    onSelect: (id: number | null) => void;
}> = ({tree, selectedId, onSelect}) => {
    const {t} = useTranslation();

    return (
        <div className="flex flex-wrap items-center gap-1.5">
            <button
                onClick={() => onSelect(null)}
                className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                    selectedId === null
                        ? 'bg-emerald-600 text-white shadow-sm'
                        : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-emerald-50 dark:hover:bg-gray-600'
                }`}
            >
                {t('categories.all', 'All')}
            </button>
            {tree.map(parent => (
                <CategoryGroup
                    key={parent.id}
                    node={parent}
                    selectedId={selectedId}
                    onSelect={onSelect}
                />
            ))}
        </div>
    );
};

const CategoryGroup: React.FC<{
    node: CategoryTreeNode;
    selectedId: number | null;
    onSelect: (id: number | null) => void;
}> = ({node, selectedId, onSelect}) => {
    const [open, setOpen] = useState(false);
    const hasChildren = node.children.length > 0;

    if (hasChildren) {
        return (
            <div className="flex items-center gap-1">
                <button
                    onClick={() => onSelect(selectedId === node.id ? null : node.id)}
                    className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                        selectedId === node.id
                            ? 'bg-emerald-600 text-white shadow-sm'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-emerald-50 dark:hover:bg-gray-600'
                    }`}
                >
                    {node.name}
                </button>
                <button
                    onClick={() => setOpen(!open)}
                    className="w-6 h-6 flex items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700 text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                    {open ? <ChevronDown size={12}/> : <ChevronRight size={12}/>}
                </button>
                {open && (
                    <>
                        {node.children.map(child => (
                            <button
                                key={child.id}
                                onClick={() => onSelect(selectedId === child.id ? null : child.id)}
                                className={`px-2.5 py-1 rounded-full text-xs transition-colors ${
                                    selectedId === child.id
                                        ? 'bg-emerald-500 text-white shadow-sm'
                                        : 'bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 hover:bg-emerald-50 dark:hover:bg-gray-700'
                                }`}
                            >
                                {child.name}
                            </button>
                        ))}
                    </>
                )}
            </div>
        );
    }

    return (
        <button
            onClick={() => onSelect(selectedId === node.id ? null : node.id)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                selectedId === node.id
                    ? 'bg-emerald-600 text-white shadow-sm'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-emerald-50 dark:hover:bg-gray-600'
            }`}
        >
            {node.name}
        </button>
    );
};

const CategoriesPage = () => {
    const {t} = useTranslation();
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [page, setPage] = useState(1);
    const pageSize = 12;

    useEffect(() => {
        const fetchCategories = async () => {
            try {
                setLoading(true);
                const response = await categoryApi.getAll();
                setCategories((response as any)?.items || response || []);
            } catch (err) {
                setError(t('common.error', 'Error'));
                console.error('Failed to fetch categories:', err);
            } finally {
                setLoading(false);
            }
        };
        fetchCategories();
    }, [t]);

    const tree = useMemo(() => {
        const enabled = categories.filter(c => c.status === 1);
        return buildCategoryTree(enabled);
    }, [categories]);

    const categoryIdsForFilter = useMemo((): number[] | undefined => {
        if (selectedId === null) return undefined;
        const node = findNodeById(tree, selectedId);
        if (!node) return [selectedId];
        if (node.children.length === 0) return [selectedId];
        return [node.id, ...node.children.map(c => c.id)];
    }, [tree, selectedId]);

    useEffect(() => {
        setPage(1);
    }, [selectedId]);

    const {data: mediaData, isLoading: mediaLoading} = useMediaList({
        page,
        page_size: pageSize,
        status: 'active',
        category_ids: categoryIdsForFilter,
    });

    const items = mediaData?.items || [];
    const total = mediaData?.total || 0;
    const totalPages = Math.ceil(total / pageSize);

    const selectedName = useMemo(() => {
        if (selectedId === null) return null;
        const node = findNodeById(tree, selectedId);
        return node?.name ?? null;
    }, [tree, selectedId]);

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner/>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-16 text-muted-foreground">
                <Folder size={48} className="mx-auto mb-3 opacity-30"/>
                <p>{error}</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Folder size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                    {t('categories.title', 'Categories')}
                </h1>
            </div>

            <CategoryFilter
                tree={tree}
                selectedId={selectedId}
                onSelect={setSelectedId}
            />

            {selectedName && (
                <div className="flex items-center gap-2">
                    <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                        {selectedName}
                    </h2>
                    <span className="text-sm text-gray-400">
                        {total} {t('categories.videos', 'videos')}
                    </span>
                    <button
                        onClick={() => setSelectedId(null)}
                        className="text-xs text-emerald-600 hover:text-emerald-700 ml-2"
                    >
                        {t('categories.clearFilter', 'Clear')}
                    </button>
                </div>
            )}

            {mediaLoading ? (
                <div className="flex items-center justify-center py-16">
                    <Spinner/>
                </div>
            ) : items.length === 0 ? (
                <div className="text-center py-16 text-muted-foreground">
                    <Folder size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>{t('categories.noVideos', 'No videos found')}</p>
                </div>
            ) : (
                <>
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                        {items.map((media: any) => (
                            <VideoCard key={media.id} media={media}/>
                        ))}
                    </div>

                    {totalPages > 1 && (
                        <div className="flex items-center justify-center gap-2 pt-4">
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={page <= 1}
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                            >
                                {t('common.prev', 'Previous')}
                            </Button>
                            <span className="text-sm text-gray-400">
                                {page} / {totalPages}
                            </span>
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={page >= totalPages}
                                onClick={() => setPage(p => p + 1)}
                            >
                                {t('common.next', 'Next')}
                            </Button>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};

function findNodeById(nodes: CategoryTreeNode[], id: number): CategoryTreeNode | null {
    for (const node of nodes) {
        if (node.id === id) return node;
        const found = findNodeById(node.children, id);
        if (found) return found;
    }
    return null;
}

export default CategoriesPage;
