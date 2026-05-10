import React, {useState, useMemo, useRef, useCallback, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Star, Clock, ChevronRight, LayoutGrid, List, Filter} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Spinner} from '@/components/ui/spinner';
import {Skeleton} from '@/components/ui/skeleton';
import {Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia} from '@/components/ui/empty';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {useTranslation} from 'react-i18next';
import {useMediaList, useInfiniteMediaList} from '@/hooks/queries';
import ErrorPage from '@/components/common/ErrorPage';
import HeroCarousel, {HeroCarouselSkeleton} from '@/components/common/HeroCarousel';
import BannerCarousel, {BannerCarouselSkeleton} from '@/components/common/BannerCarousel';
import CategoryFilter, {CategoryFilterSkeleton} from '@/components/common/CategoryFilter';
import HorizontalScroll from '@/components/common/HorizontalScroll';
import VideoCardSkeleton from '@/components/common/VideoCardSkeleton';

type LayoutMode = 'grid' | 'list';

const FeaturedPage = () => {
    const {t} = useTranslation();
    const [activeCategoryId, setActiveCategoryId] = useState<number | string | null>(null);
    const [layoutMode, setLayoutMode] = useState<LayoutMode>('grid');
    const sectionCRef = useRef<HTMLDivElement>(null);
    const sectionDRef = useRef<HTMLDivElement>(null);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const {data, isLoading, error} = useMediaList({
        featured: 'true',
        page_size: 20,
        status: 'active',
    });

    const {
        data: infiniteData,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteMediaList({
        page_size: 20,
        status: 'active',
    });

    const featuredMedia = data?.items || [];

    const categories = useMemo(() => {
        const catMap = new Map<number, {id: number; name: string}>();
        featuredMedia.forEach((item) => {
            if (item.category_id && item.edges?.category?.name) {
                if (!catMap.has(item.category_id)) {
                    catMap.set(item.category_id, {
                        id: item.category_id,
                        name: item.edges.category.name,
                    });
                }
            }
        });
        return Array.from(catMap.values());
    }, [featuredMedia]);

    const filteredMedia = useMemo(() => {
        if (activeCategoryId === null) return featuredMedia;
        return featuredMedia.filter((item) => item.category_id === activeCategoryId);
    }, [featuredMedia, activeCategoryId]);

    const heroItems = useMemo(() => featuredMedia.slice(0, 5), [featuredMedia]);
    const horizontalItems = useMemo(() => filteredMedia.slice(0, 8), [filteredMedia]);
    const gridItems = useMemo(() => filteredMedia, [filteredMedia]);

    useEffect(() => {
        if (!sentinelRef.current || !hasNextPage) return;
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
                    fetchNextPage();
                }
            },
            {rootMargin: '200px'},
        );
        observer.observe(sentinelRef.current);
        return () => observer.disconnect();
    }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

    const scrollToSectionC = useCallback(() => {
        sectionCRef.current?.scrollIntoView({behavior: 'smooth', block: 'start'});
    }, []);

    if (isLoading) {
        return (
            <div className="space-y-6">
                <HeroCarouselSkeleton/>
                <CategoryFilterSkeleton/>
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <Skeleton className="h-6 w-32"/>
                        <Skeleton className="h-4 w-20"/>
                    </div>
                    <div className="flex gap-4 overflow-hidden">
                        {Array.from({length: 4}).map((_, i) => (
                            <VideoCardSkeleton key={i}/>
                        ))}
                    </div>
                </div>
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <Skeleton className="h-6 w-32"/>
                        <Skeleton className="h-4 w-20"/>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
                        {Array.from({length: 8}).map((_, i) => (
                            <VideoCardSkeleton key={i}/>
                        ))}
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error.message || t('common.noData')}/>;
    }

    if (featuredMedia.length === 0) {
        return (
            <Empty className="py-20">
                <EmptyMedia variant="icon">
                    <Star size={24}/>
                </EmptyMedia>
                <EmptyHeader>
                    <EmptyTitle>{t('featured.emptyTitle')}</EmptyTitle>
                    <EmptyDescription>{t('featured.emptyDesc')}</EmptyDescription>
                </EmptyHeader>
                <Link to="/">
                    <Button variant="outline">{t('error.backToHome')}</Button>
                </Link>
            </Empty>
        );
    }

    return (
        <div className="space-y-6">
            <BannerCarousel className="mb-6" />

            <HeroCarousel
                items={heroItems}
                autoPlayInterval={6000}
                onLearnMore={scrollToSectionC}
            />

            <CategoryFilter
                categories={categories}
                activeId={activeCategoryId}
                onSelect={setActiveCategoryId}
                className="mt-6"
            />

            <div ref={sectionCRef} className="mt-6">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                        <Star size={20} className="text-warning"/>
                        <h2 className="text-xl font-bold text-foreground">{t('featured.editorPick')}</h2>
                    </div>
                    <button
                        onClick={() => sectionDRef.current?.scrollIntoView({behavior: 'smooth', block: 'start'})}
                        className="text-sm font-medium text-primary hover:text-primary/80 transition-colors flex items-center gap-1"
                    >
                        {t('home.viewAll')}
                        <ChevronRight size={16}/>
                    </button>
                </div>

                {horizontalItems.length > 0 ? (
                    <HorizontalScroll itemsPerPage={4}>
                        {horizontalItems.map((item) => (
                            <FeaturedHorizontalCard key={item.id} item={item}/>
                        ))}
                    </HorizontalScroll>
                ) : (
                    <div className="py-8 text-center text-muted-foreground text-sm">
                        {t('common.noData')}
                    </div>
                )}
            </div>

            <div ref={sectionDRef} className="mt-10">
                <div className="flex items-center justify-between mb-5">
                    <div className="flex items-center gap-2">
                        <h2 className="text-xl font-bold text-foreground">{t('featured.allFeatured')}</h2>
                        <span className="text-sm text-muted-foreground">
                            {t('featured.featuredCount', {count: gridItems.length})}
                        </span>
                    </div>
                    <div className="hidden sm:flex items-center gap-1">
                        <Button
                            variant={layoutMode === 'grid' ? 'default' : 'ghost'}
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => setLayoutMode('grid')}
                            aria-label="Grid layout"
                        >
                            <LayoutGrid size={16}/>
                        </Button>
                        <Button
                            variant={layoutMode === 'list' ? 'default' : 'ghost'}
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => setLayoutMode('list')}
                            aria-label="List layout"
                        >
                            <List size={16}/>
                        </Button>
                    </div>
                </div>

                {gridItems.length === 0 ? (
                    <Empty className="py-12">
                        <EmptyMedia variant="icon">
                            <Filter size={24}/>
                        </EmptyMedia>
                        <EmptyHeader>
                            <EmptyTitle>{t('featured.noResultsTitle')}</EmptyTitle>
                            <EmptyDescription>{t('featured.noResultsDesc')}</EmptyDescription>
                        </EmptyHeader>
                        <Button variant="outline" onClick={() => setActiveCategoryId(null)}>
                            {t('featured.clearFilter')}
                        </Button>
                    </Empty>
                ) : (
                    <>
                        {layoutMode === 'grid' ? (
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
                                {gridItems.map((item) => (
                                    <FeaturedGridCard key={item.id} item={item}/>
                                ))}
                            </div>
                        ) : (
                            <div className="flex flex-col gap-3">
                                {gridItems.map((item) => (
                                    <FeaturedListCard key={item.id} item={item}/>
                                ))}
                            </div>
                        )}

                        <div ref={sentinelRef} className="py-4">
                            {isFetchingNextPage && (
                                <div className="flex items-center justify-center gap-2 py-6">
                                    <Spinner size="sm"/>
                                    <span className="text-sm text-muted-foreground">{t('common.loading')}</span>
                                </div>
                            )}
                            {!hasNextPage && gridItems.length > 8 && (
                                <div className="text-center text-sm text-muted-foreground py-4">
                                    — {t('common.allLoaded')} —
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

interface FeaturedCardProps {
    item: {
        id: string;
        short_token?: string;
        title: string;
        description?: string;
        thumbnail?: string;
        duration: number;
        view_count: number;
        create_time?: string;
        edges?: {
            user?: Array<{
                id: string;
                username: string;
                nickname?: string;
                avatar?: string;
            }>;
            category?: {
                id: number;
                name: string;
            };
        };
    };
}

const FeaturedHorizontalCard: React.FC<FeaturedCardProps> = ({item}) => {
    const {t} = useTranslation();
    const user = item.edges?.user?.[0];

    return (
        <Link
            to="/watch"
            search={{v: item.short_token || item.id}}
            className="group block w-56 sm:w-60 md:w-64 lg:w-72 shrink-0"
        >
            <div className="rounded-xl overflow-hidden bg-card border border-border shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-1">
                <div className="relative aspect-video overflow-hidden">
                    <img
                        src={getImageUrl(item.thumbnail, 'thumbnail')}
                        alt={item.title}
                        loading="lazy"
                        onError={(e) => handleImageError(e, 'thumbnail')}
                        className="object-cover w-full h-full group-hover:scale-105 transition-transform duration-500"
                    />
                    <Badge
                        variant="secondary"
                        className="absolute bottom-2 right-2 bg-black/80 text-white text-xs backdrop-blur-sm"
                    >
                        {formatDuration(item.duration)}
                    </Badge>
                    <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <div
                            className="w-12 h-12 rounded-full bg-white/90 dark:bg-gray-800/90 flex items-center justify-center shadow-lg scale-75 group-hover:scale-100 transition-transform duration-300">
                            <Play className="w-5 h-5 text-foreground ml-0.5" fill="currentColor"/>
                        </div>
                    </div>
                </div>
                <div className="p-3">
                    <h3 className="text-sm font-medium text-foreground line-clamp-2 mb-1.5 group-hover:text-primary transition-colors">
                        {item.title}
                    </h3>
                    <div className="flex items-center gap-2 mb-1">
                        <Avatar className="h-5 w-5">
                            <AvatarImage
                                src={user?.avatar ? getImageUrl(user.avatar, 'avatar') : undefined}
                                alt={user?.username}
                            />
                            <AvatarFallback className="text-[10px]">
                                {user?.username?.[0] || 'U'}
                            </AvatarFallback>
                        </Avatar>
                        <span className="text-xs text-muted-foreground truncate">
                            {user?.username || 'Unknown'}
                        </span>
                    </div>
                    <div className="text-xs text-muted-foreground flex items-center gap-1">
                        <Eye size={12}/>
                        {formatViews(item.view_count)}
                    </div>
                </div>
            </div>
        </Link>
    );
};

const FeaturedGridCard: React.FC<FeaturedCardProps> = ({item}) => {
    const {t} = useTranslation();
    const user = item.edges?.user?.[0];

    return (
        <Link to="/watch" search={{v: item.short_token || item.id}} className="group block">
            <div
                className="rounded-xl bg-card overflow-hidden border border-border shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                <div className="relative aspect-video overflow-hidden">
                    <img
                        src={getImageUrl(item.thumbnail, 'thumbnail')}
                        alt={item.title}
                        loading="lazy"
                        onError={(e) => handleImageError(e, 'thumbnail')}
                        className="object-cover w-full h-full group-hover:scale-105 transition-transform duration-500"
                    />
                    <Badge
                        variant="secondary"
                        className="absolute bottom-2 right-2 bg-black/80 text-white text-xs backdrop-blur-sm"
                    >
                        {formatDuration(item.duration)}
                    </Badge>
                    <div
                        className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center">
                        <div
                            className="w-12 h-12 bg-white/90 dark:bg-gray-800/90 rounded-full flex items-center justify-center shadow-lg transform scale-75 group-hover:scale-100 transition-transform duration-300">
                            <Play className="w-5 h-5 text-foreground ml-0.5" fill="currentColor"/>
                        </div>
                    </div>
                </div>
                <div className="p-3">
                    <h3 className="text-sm font-medium text-foreground line-clamp-2 mb-1.5 group-hover:text-primary transition-colors">
                        {item.title}
                    </h3>
                    <div className="flex items-center gap-2 mb-1">
                        <Avatar className="h-5 w-5">
                            <AvatarImage
                                src={user?.avatar ? getImageUrl(user.avatar, 'avatar') : undefined}
                                alt={user?.username}
                            />
                            <AvatarFallback className="text-[10px]">
                                {user?.username?.[0] || 'U'}
                            </AvatarFallback>
                        </Avatar>
                        <span className="text-xs text-muted-foreground truncate">
                            {user?.username || 'Unknown'}
                        </span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                            <Eye size={12}/>
                            {formatViews(item.view_count)}
                        </span>
                        {item.create_time && (
                            <span className="flex items-center gap-1">
                                <Clock size={12}/>
                                {formatDate(item.create_time)}
                            </span>
                        )}
                    </div>
                </div>
            </div>
        </Link>
    );
};

const FeaturedListCard: React.FC<FeaturedCardProps> = ({item}) => {
    const {t} = useTranslation();
    const user = item.edges?.user?.[0];

    return (
        <Link to="/watch" search={{v: item.short_token || item.id}} className="group block">
            <div className="flex gap-3 rounded-xl bg-card overflow-hidden border border-border p-2 hover:shadow-md transition-all duration-200">
                <div className="w-40 shrink-0 aspect-video rounded-lg overflow-hidden relative">
                    <img
                        src={getImageUrl(item.thumbnail, 'thumbnail')}
                        alt={item.title}
                        loading="lazy"
                        onError={(e) => handleImageError(e, 'thumbnail')}
                        className="object-cover w-full h-full group-hover:scale-105 transition-transform duration-500"
                    />
                    <Badge
                        variant="secondary"
                        className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] backdrop-blur-sm"
                    >
                        {formatDuration(item.duration)}
                    </Badge>
                </div>
                <div className="flex flex-col justify-center py-1 min-w-0">
                    <h3 className="text-sm font-medium text-foreground line-clamp-2 mb-1 group-hover:text-primary transition-colors">
                        {item.title}
                    </h3>
                    <span className="text-xs text-muted-foreground mb-0.5">
                        {user?.username || 'Unknown'}
                    </span>
                    <div className="text-xs text-muted-foreground flex items-center gap-2">
                        <span className="flex items-center gap-1">
                            <Eye size={12}/>
                            {formatViews(item.view_count)}
                        </span>
                        {item.create_time && (
                            <span className="flex items-center gap-1">
                                <Clock size={12}/>
                                {formatDate(item.create_time)}
                            </span>
                        )}
                    </div>
                </div>
            </div>
        </Link>
    );
};

export default FeaturedPage;
