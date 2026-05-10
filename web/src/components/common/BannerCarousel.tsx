import React, {useState, useEffect, useCallback, useRef} from 'react';
import {Link} from '@tanstack/react-router';
import {ChevronLeft, ChevronRight} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import type {Banner} from '@/lib/api/portal';
import {usePortalConfig} from '@/hooks/queries';

interface BannerCarouselProps {
    className?: string;
}

const BannerCarousel: React.FC<BannerCarouselProps> = ({className = ''}) => {
    const {t, i18n} = useTranslation();
    const [currentIndex, setCurrentIndex] = useState(0);
    const [isPaused, setIsPaused] = useState(false);
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const {data: portalConfig} = usePortalConfig();
    const banners = portalConfig?.banners || [];
    const activeBanners = banners.filter(b => b.is_active);

    const getLocalizedText = useCallback((text?: string, i18nMap?: Record<string, string>) => {
        if (!text && !i18nMap) return '';
        const lang = i18n.language;
        if (i18nMap && i18nMap[lang]) return i18nMap[lang];
        return text || '';
    }, [i18n.language]);

    useEffect(() => {
        if (activeBanners.length <= 1 || isPaused) {
            if (timerRef.current) clearTimeout(timerRef.current);
            return;
        }

        const currentBanner = activeBanners[currentIndex];
        const interval = currentBanner?.auto_slide_interval || 5000;

        timerRef.current = setTimeout(() => {
            setCurrentIndex(prev => (prev + 1) % activeBanners.length);
        }, interval);

        return () => {
            if (timerRef.current) clearTimeout(timerRef.current);
        };
    }, [currentIndex, isPaused, activeBanners]);

    const goNext = useCallback(() => {
        if (activeBanners.length === 0) return;
        setCurrentIndex(prev => (prev + 1) % activeBanners.length);
    }, [activeBanners.length]);

    const goPrev = useCallback(() => {
        if (activeBanners.length === 0) return;
        setCurrentIndex(prev => (prev - 1 + activeBanners.length) % activeBanners.length);
    }, [activeBanners.length]);

    if (activeBanners.length === 0) {
        return null;
    }

    const currentBanner = activeBanners[currentIndex];

    return (
        <div
            className={`relative w-full overflow-hidden rounded-lg ${className}`}
            onMouseEnter={() => setIsPaused(true)}
            onMouseLeave={() => setIsPaused(false)}
        >
            <div className="relative aspect-[21/9] md:aspect-[21/9] w-full">
                {currentBanner.image_url ? (
                    <img
                        src={currentBanner.image_url}
                        alt={getLocalizedText(currentBanner.title, currentBanner.title_i18n)}
                        className="w-full h-full object-cover"
                    />
                ) : (
                    <div
                        className="w-full h-full flex items-center justify-center text-white text-xl font-bold"
                        style={{
                            background: `linear-gradient(135deg, ${currentBanner.bg_color_start || '#667eea'}, ${currentBanner.bg_color_end || '#764ba2'})`,
                        }}
                    >
                        {getLocalizedText(currentBanner.title, currentBanner.title_i18n)}
                    </div>
                )}

                <div
                    className="absolute inset-0 bg-gradient-to-r from-black/60 to-transparent"
                    style={{
                        background: currentBanner.bg_color_start
                            ? `linear-gradient(90deg, ${currentBanner.bg_color_start}cc 0%, transparent 60%)`
                            : undefined,
                    }}
                />

                <div className="absolute inset-0 flex items-center">
                    <div className="px-6 md:px-12 max-w-xl">
                        {currentBanner.badge_text && (
                            <span className="inline-block px-3 py-1 mb-3 text-xs font-semibold text-white bg-white/20 backdrop-blur-sm rounded-full">
                                {currentBanner.badge_text}
                            </span>
                        )}

                        <h2 className="text-2xl md:text-4xl font-bold text-white mb-2 drop-shadow-lg">
                            {getLocalizedText(currentBanner.title, currentBanner.title_i18n)}
                        </h2>

                        {currentBanner.subtitle && (
                            <p className="text-white/90 text-sm md:text-base mb-4 line-clamp-2">
                                {getLocalizedText(currentBanner.subtitle, currentBanner.subtitle_i18n)}
                            </p>
                        )}

                        <div className="flex gap-3">
                            {currentBanner.primary_btn_text && currentBanner.primary_btn_url && (
                                <Link to={currentBanner.primary_btn_url.startsWith('/') ? currentBanner.primary_btn_url as '/' | `/search` | `/watch/$id` | `/channel/$token` | `/playlist/$token` | `/user/$username` | `/${string}` : '/'}>
                                    <Button className="bg-white text-gray-900 hover:bg-white/90">
                                        {currentBanner.primary_btn_text}
                                    </Button>
                                </Link>
                            )}
                            {currentBanner.secondary_btn_text && currentBanner.secondary_btn_url && (
                                <Link to={currentBanner.secondary_btn_url.startsWith('/') ? currentBanner.secondary_btn_url as '/' | `/search` | `/watch/$id` | `/channel/$token` | `/playlist/$token` | `/user/$username` | `/${string}` : '/'}>
                                    <Button variant="outline" className="border-white text-white hover:bg-white/20">
                                        {currentBanner.secondary_btn_text}
                                    </Button>
                                </Link>
                            )}
                        </div>
                    </div>
                </div>

                {activeBanners.length > 1 && (
                    <>
                        <button
                            onClick={goPrev}
                            className="absolute left-2 md:left-4 top-1/2 -translate-y-1/2 w-8 h-8 md:w-10 md:h-10 rounded-full bg-black/40 hover:bg-black/60 text-white flex items-center justify-center transition-colors"
                            aria-label="Previous banner"
                        >
                            <ChevronLeft className="w-5 h-5"/>
                        </button>
                        <button
                            onClick={goNext}
                            className="absolute right-2 md:right-4 top-1/2 -translate-y-1/2 w-8 h-8 md:w-10 md:h-10 rounded-full bg-black/40 hover:bg-black/60 text-white flex items-center justify-center transition-colors"
                            aria-label="Next banner"
                        >
                            <ChevronRight className="w-5 h-5"/>
                        </button>

                        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2">
                            {activeBanners.map((_, idx) => (
                                <button
                                    key={idx}
                                    onClick={() => setCurrentIndex(idx)}
                                    className={`w-2 h-2 rounded-full transition-all ${
                                        idx === currentIndex ? 'bg-white w-6' : 'bg-white/50 hover:bg-white/70'
                                    }`}
                                    aria-label={`Go to banner ${idx + 1}`}
                                />
                            ))}
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

export default BannerCarousel;

export const BannerCarouselSkeleton: React.FC<{className?: string}> = ({className = ''}) => (
    <div className={`relative w-full overflow-hidden rounded-lg bg-gradient-to-br from-primary/20 to-primary/5 animate-pulse ${className}`}>
        <div className="aspect-[21/9] w-full"/>
    </div>
);
