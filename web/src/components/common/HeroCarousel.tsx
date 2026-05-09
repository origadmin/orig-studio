import React, {useState, useEffect, useCallback, useRef} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, ChevronLeft, ChevronRight, Eye, Star} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Skeleton} from '@/components/ui/skeleton';
import {formatDuration, formatViews} from '@/lib/format';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import type {Media} from '@/lib/api/media';

interface HeroCarouselItem {
    id: string;
    short_token?: string;
    title: string;
    description?: string;
    thumbnail?: string;
    duration: number;
    view_count: number;
    edges?: {
        user?: Array<{
            id: string;
            username: string;
            nickname?: string;
            avatar?: string;
        }>;
    };
}

interface HeroCarouselProps {
    items: HeroCarouselItem[];
    autoPlayInterval?: number;
    className?: string;
    onLearnMore?: () => void;
}

const HeroCarousel: React.FC<HeroCarouselProps> = ({
    items,
    autoPlayInterval = 6000,
    className = '',
    onLearnMore,
}) => {
    const {t} = useTranslation();
    const [currentIndex, setCurrentIndex] = useState(0);
    const [isPaused, setIsPaused] = useState(false);
    const [touchStart, setTouchStart] = useState<number | null>(null);
    const [touchEnd, setTouchEnd] = useState<number | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const prefersReducedMotion = typeof window !== 'undefined'
        ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
        : false;

    const total = items.length;

    const goTo = useCallback((index: number) => {
        setCurrentIndex(index >= total ? 0 : index < 0 ? total - 1 : index);
    }, [total]);

    const goNext = useCallback(() => {
        goTo(currentIndex + 1);
    }, [currentIndex, goTo]);

    const goPrev = useCallback(() => {
        goTo(currentIndex - 1);
    }, [currentIndex, goTo]);

    const resetTimer = useCallback(() => {
        if (timerRef.current) {
            clearTimeout(timerRef.current);
        }
    }, []);

    useEffect(() => {
        if (isPaused || total <= 1 || prefersReducedMotion) {
            resetTimer();
            return;
        }
        resetTimer();
        timerRef.current = setTimeout(goNext, autoPlayInterval);
        return resetTimer;
    }, [currentIndex, isPaused, goNext, autoPlayInterval, total, prefersReducedMotion, resetTimer]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!containerRef.current?.contains(document.activeElement)) return;
            if (e.key === 'ArrowLeft') {
                goPrev();
                resetTimer();
            } else if (e.key === 'ArrowRight') {
                goNext();
                resetTimer();
            }
        };
        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [goPrev, goNext, resetTimer]);

    const handleTouchStart = (e: React.TouchEvent) => {
        setTouchEnd(null);
        setTouchStart(e.targetTouches[0].clientX);
    };

    const handleTouchMove = (e: React.TouchEvent) => {
        setTouchEnd(e.targetTouches[0].clientX);
    };

    const handleTouchEnd = () => {
        if (!touchStart || !touchEnd) return;
        const distance = touchStart - touchEnd;
        const isLeftSwipe = distance > 50;
        const isRightSwipe = distance < -50;
        if (isLeftSwipe) {
            goNext();
            resetTimer();
        }
        if (isRightSwipe) {
            goPrev();
            resetTimer();
        }
    };

    if (total === 0) return null;

    const currentItem = items[currentIndex];
    const user = currentItem.edges?.user?.[0];
    const thumbnailUrl = getImageUrl(currentItem.thumbnail, 'thumbnail');

    const visibleIndices = total <= 3
        ? Array.from({length: total}, (_, i) => i)
        : [
            (currentIndex - 1 + total) % total,
            currentIndex,
            (currentIndex + 1) % total,
        ];

    return (
        <div
            ref={containerRef}
            className={`group relative ${className}`}
            role="region"
            aria-label={t('featured.carouselAriaLabel', 'Featured content carousel')}
            aria-roledescription="carousel"
            onMouseEnter={() => setIsPaused(true)}
            onMouseLeave={() => setIsPaused(false)}
            tabIndex={0}
        >
            <div
                className="relative h-[260px] sm:h-[280px] md:h-[320px] lg:h-[400px] rounded-xl md:rounded-2xl overflow-hidden"
                onTouchStart={handleTouchStart}
                onTouchMove={handleTouchMove}
                onTouchEnd={handleTouchEnd}
            >
                {items.map((item, index) => {
                    const isActive = index === currentIndex;
                    if (!visibleIndices.includes(index)) return null;

                    const itemThumbnailUrl = getImageUrl(item.thumbnail, 'thumbnail');

                    return (
                        <div
                            key={item.id}
                            role="group"
                            aria-roledescription="slide"
                            aria-label={t('featured.slideLabel', 'Slide {{current}} of {{total}}', {current: index + 1, total})}
                            className={`absolute inset-0 transition-opacity duration-500 ${isActive ? 'opacity-100 z-10' : 'opacity-0 z-0'}`}
                        >
                            <div
                                className="absolute inset-0 bg-cover bg-center"
                                style={{backgroundImage: `url(${itemThumbnailUrl})`}}
                            />
                            <div className="absolute inset-0 bg-gradient-to-r from-background/95 via-background/70 to-background/30 dark:from-background/95 dark:via-background/80 dark:to-background/40"/>
                            <div className="absolute bottom-0 left-0 right-0 h-1/4 bg-gradient-to-t from-background/60 to-transparent"/>

                            <div className="relative h-full flex items-center px-5 py-6 sm:px-8 sm:py-8 md:px-10 md:py-10">
                                <div className="max-w-full sm:max-w-sm md:max-w-md lg:max-w-lg">
                                    <Badge variant="warning" className="mb-4">
                                        <Star size={12} fill="currentColor" className="mr-1"/>
                                        {t('featured.editorPick')}
                                    </Badge>

                                    <h2 className="text-xl md:text-2xl lg:text-3xl font-bold md:font-bold lg:font-black text-foreground leading-tight line-clamp-2 mb-3">
                                        {item.title}
                                    </h2>

                                    <p className="text-sm md:text-base text-muted-foreground line-clamp-2 mb-5">
                                        {item.description || t('watch.noDescription')}
                                    </p>

                                    <div className="flex items-center gap-3 mb-6">
                                        <Avatar className="h-8 w-8">
                                            <AvatarImage
                                                src={item.edges?.user?.[0]?.avatar ? getImageUrl(item.edges.user[0].avatar, 'avatar') : undefined}
                                                alt={item.edges?.user?.[0]?.username}
                                            />
                                            <AvatarFallback className="text-xs">
                                                {item.edges?.user?.[0]?.username?.[0] || 'U'}
                                            </AvatarFallback>
                                        </Avatar>
                                        <span className="text-sm font-medium text-foreground">
                                            {item.edges?.user?.[0]?.username || 'Unknown'}
                                        </span>
                                        <span className="text-sm text-muted-foreground flex items-center gap-1">
                                            <Eye size={14}/>
                                            {formatViews(item.view_count)}
                                        </span>
                                        <Badge variant="secondary" className="text-xs">
                                            {formatDuration(item.duration)}
                                        </Badge>
                                    </div>

                                    <div className="flex flex-col sm:flex-row gap-3">
                                        <Link to="/watch" search={{v: item.short_token || item.id}}>
                                            <Button size="lg">
                                                <Play size={18} className="mr-2"/>
                                                {t('featured.watchNow')}
                                            </Button>
                                        </Link>
                                        <Button
                                            size="lg"
                                            variant="outline"
                                            onClick={onLearnMore}
                                            className="hidden sm:inline-flex"
                                        >
                                            {t('featured.learnMore')}
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>

            {total > 1 && (
                <>
                    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 z-20">
                        {items.map((_, index) => (
                            <button
                                key={index}
                                onClick={() => {
                                    goTo(index);
                                    resetTimer();
                                }}
                                aria-label={t('featured.goToSlide', 'Go to slide {{n}}', {n: index + 1})}
                                aria-current={index === currentIndex ? 'true' : undefined}
                                className={`h-2 rounded-full transition-all duration-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 ${
                                    index === currentIndex
                                        ? 'w-6 bg-primary'
                                        : 'w-2 bg-foreground/30 hover:bg-foreground/50'
                                }`}
                            />
                        ))}
                    </div>

                    <Button
                        variant="ghost"
                        size="icon"
                        className="absolute left-4 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-background/60 backdrop-blur-sm hover:bg-background/80 text-foreground shadow-md opacity-0 group-hover:opacity-100 transition-opacity hidden md:flex focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        onClick={() => {
                            goPrev();
                            resetTimer();
                        }}
                        aria-label={t('featured.prevSlide', 'Previous slide')}
                    >
                        <ChevronLeft size={20}/>
                    </Button>

                    <Button
                        variant="ghost"
                        size="icon"
                        className="absolute right-4 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-background/60 backdrop-blur-sm hover:bg-background/80 text-foreground shadow-md opacity-0 group-hover:opacity-100 transition-opacity hidden md:flex focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        onClick={() => {
                            goNext();
                            resetTimer();
                        }}
                        aria-label={t('featured.nextSlide', 'Next slide')}
                    >
                        <ChevronRight size={20}/>
                    </Button>
                </>
            )}
        </div>
    );
};

export const HeroCarouselSkeleton: React.FC = () => (
    <div className="relative h-[260px] sm:h-[280px] md:h-[320px] lg:h-[400px] rounded-xl md:rounded-2xl overflow-hidden bg-card">
        <div className="relative h-full flex items-center px-5 py-6 sm:px-8 sm:py-8 md:px-10 md:py-10">
            <div className="max-w-lg space-y-4">
                <Skeleton className="h-6 w-24 rounded-full"/>
                <Skeleton className="h-8 w-3/4"/>
                <Skeleton className="h-4 w-1/2"/>
                <div className="flex items-center gap-3">
                    <Skeleton className="h-8 w-8 rounded-full"/>
                    <Skeleton className="h-4 w-24"/>
                    <Skeleton className="h-4 w-16"/>
                </div>
                <div className="flex gap-3">
                    <Skeleton className="h-10 w-28 rounded-btn"/>
                    <Skeleton className="h-10 w-28 rounded-btn"/>
                </div>
            </div>
        </div>
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2">
            <Skeleton className="h-2 w-6 rounded-full"/>
            <Skeleton className="h-2 w-2 rounded-full"/>
            <Skeleton className="h-2 w-2 rounded-full"/>
        </div>
    </div>
);

export default HeroCarousel;
