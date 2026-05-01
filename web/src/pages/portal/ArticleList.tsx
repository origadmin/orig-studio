/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Portal - Article List Page (Video Website Style)
 */

import {useState, useEffect, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {articleApi, type Article} from '@/lib/api/article';
import {API_BASE_URL} from '@/lib/request';
import {Spinner} from '@/components/ui/spinner';
import {Badge} from '@/components/ui/badge';
import {Input} from '@/components/ui/input';
import {Button} from '@/components/ui/button';
import {
    FileText,
    Search,
    Play,
    Eye,
    Clock,
    User,
    Star,
    Film,
} from 'lucide-react';
import {formatRelativeTime} from '@/lib/format';

/**
 * Resolve a potentially relative URL to a full URL.
 */
function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

const PAGE_SIZE = 12;

export default function ArticleListPage() {
    const [articles, setArticles] = useState<Article[]>([]);
    const [featuredArticles, setFeaturedArticles] = useState<Article[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);

    // Load articles
    const loadArticles = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await articleApi.list({
                page,
                page_size: PAGE_SIZE,
                keyword: search || undefined,
            });
            setArticles(response.items || []);
            setTotal(response.total || 0);
        } catch (err) {
            setError('Failed to load articles');
            console.error('Error loading articles:', err);
        } finally {
            setLoading(false);
        }
    }, [page, search]);

    // Load featured articles
    useEffect(() => {
        articleApi.featured(4)
            .then(data => setFeaturedArticles(Array.isArray(data) ? data : []))
            .catch(() => setFeaturedArticles([]));
    }, []);

    // Load articles on page/search change
    useEffect(() => {
        loadArticles();
    }, [loadArticles]);

    const totalPages = Math.ceil(total / PAGE_SIZE);

    return (
        <div className="min-h-screen bg-background">
            <div className="max-w-7xl mx-auto px-4 py-6 space-y-8">
                {/* Page Header */}
                <div className="space-y-4">
                    <div className="flex items-center gap-3">
                        <FileText className="w-8 h-8 text-primary"/>
                        <h1 className="text-3xl font-bold tracking-tight">Articles</h1>
                    </div>
                    <p className="text-muted-foreground">
                        Discover in-depth articles, tutorials, and insights from our community.
                    </p>
                </div>

                {/* Featured Articles Section */}
                {featuredArticles.length > 0 && (
                    <div className="space-y-4">
                        <div className="flex items-center gap-2">
                            <Star className="w-5 h-5 text-amber-500"/>
                            <h2 className="text-xl font-semibold">Featured</h2>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            {featuredArticles.map(article => (
                                <Link
                                    key={article.id}
                                    to="/articles/$slug"
                                    params={{slug: article.slug}}
                                    className="group"
                                >
                                    <div
                                        className="bg-card rounded-lg border overflow-hidden transition-shadow hover:shadow-lg">
                                        {/* Thumbnail */}
                                        <div className="relative aspect-video bg-muted">
                                            {resolveMediaUrl(article.thumbnail || article.media?.thumbnail) ? (
                                                <img
                                                    src={resolveMediaUrl(article.thumbnail || article.media?.thumbnail)}
                                                    alt={article.title}
                                                    className="w-full h-full object-cover"
                                                    loading="lazy"
                                                />
                                            ) : article.media_id ? (
                                                <div className="w-full h-full flex items-center justify-center">
                                                    <Film className="w-8 h-8 text-muted-foreground"/>
                                                </div>
                                            ) : (
                                                <div className="w-full h-full flex items-center justify-center">
                                                    <FileText className="w-8 h-8 text-muted-foreground"/>
                                                </div>
                                            )}
                                            {article.media && article.media.duration > 0 && (
                                                <span className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1.5 py-0.5 rounded">
                                                    {Math.floor(article.media.duration / 60)}:{String(Math.floor(article.media.duration % 60)).padStart(2, '0')}
                                                </span>
                                            )}
                                            {article.featured && (
                                                <span className="absolute top-1 left-1">
                                                    <Badge variant="outline"
                                                           className="text-warning border-amber-300 bg-black/50 text-[10px]">
                                                        Featured
                                                    </Badge>
                                                </span>
                                            )}
                                        </div>
                                        {/* Info */}
                                        <div className="p-3 space-y-1">
                                            <h3 className="font-medium text-sm line-clamp-2 group-hover:text-primary transition-colors">
                                                {article.title}
                                            </h3>
                                            {article.summary && (
                                                <p className="text-xs text-muted-foreground line-clamp-2">
                                                    {article.summary}
                                                </p>
                                            )}
                                            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                <Eye className="w-3 h-3"/>
                                                <span>{article.view_count}</span>
                                                <Clock className="w-3 h-3 ml-1"/>
                                                <span>{formatRelativeTime(article.create_time)}</span>
                                            </div>
                                        </div>
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </div>
                )}

                {/* Search Bar */}
                <div className="relative max-w-md">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                    <Input
                        placeholder="Search articles..."
                        value={search}
                        onChange={e => {
                            setSearch(e.target.value);
                            setPage(1);
                        }}
                        className="pl-10"
                    />
                </div>

                {/* Article Grid */}
                {loading ? (
                    <div className="flex items-center justify-center py-20">
                        <Spinner/>
                    </div>
                ) : error ? (
                    <div className="flex flex-col items-center justify-center py-20 gap-4">
                        <p className="text-muted-foreground">{error}</p>
                        <Button variant="outline" onClick={loadArticles}>Retry</Button>
                    </div>
                ) : articles.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-20 gap-4">
                        <FileText className="w-12 h-12 text-muted-foreground"/>
                        <p className="text-lg text-muted-foreground">
                            {search ? 'No articles match your search' : 'No articles yet'}
                        </p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {articles.map(article => (
                            <Link
                                key={article.id}
                                to="/articles/$slug"
                                params={{slug: article.slug}}
                                className="group"
                            >
                                <div
                                    className="bg-card rounded-lg border overflow-hidden transition-shadow hover:shadow-lg h-full">
                                    {/* Thumbnail */}
                                    <div className="relative aspect-video bg-muted">
                                        {resolveMediaUrl(article.thumbnail || article.media?.thumbnail) ? (
                                            <img
                                                src={resolveMediaUrl(article.thumbnail || article.media?.thumbnail)}
                                                alt={article.title}
                                                className="w-full h-full object-cover"
                                                loading="lazy"
                                            />
                                        ) : article.media_id ? (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Film className="w-8 h-8 text-muted-foreground"/>
                                            </div>
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <FileText className="w-8 h-8 text-muted-foreground"/>
                                            </div>
                                        )}
                                        {article.media && article.media.duration > 0 && (
                                            <span className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1.5 py-0.5 rounded">
                                                {Math.floor(article.media.duration / 60)}:{String(Math.floor(article.media.duration % 60)).padStart(2, '0')}
                                            </span>
                                        )}
                                    </div>
                                    {/* Info */}
                                    <div className="p-4 space-y-2">
                                        <div className="flex items-center gap-2">
                                            <Badge variant={article.state === 'published' ? 'default' : 'secondary'}
                                                   className="text-[10px]">
                                                {article.state}
                                            </Badge>
                                            {article.featured && (
                                                <Badge variant="outline"
                                                       className="text-warning border-amber-300 text-[10px]">
                                                    Featured
                                                </Badge>
                                            )}
                                        </div>
                                        <h3 className="font-semibold line-clamp-2 group-hover:text-primary transition-colors">
                                            {article.title}
                                        </h3>
                                        {article.summary && (
                                            <p className="text-sm text-muted-foreground line-clamp-2">
                                                {article.summary}
                                            </p>
                                        )}
                                        <div className="flex items-center gap-3 text-xs text-muted-foreground">
                                            <div className="flex items-center gap-1">
                                                <User className="w-3 h-3"/>
                                                <span>{article.user_id.substring(0, 8)}</span>
                                            </div>
                                            <div className="flex items-center gap-1">
                                                <Eye className="w-3 h-3"/>
                                                <span>{article.view_count}</span>
                                            </div>
                                            <div className="flex items-center gap-1">
                                                <Clock className="w-3 h-3"/>
                                                <span>{formatRelativeTime(article.create_time)}</span>
                                            </div>
                                        </div>
                                        {article.tags && article.tags.length > 0 && (
                                            <div className="flex flex-wrap gap-1">
                                                {article.tags.slice(0, 3).map((tag, i) => (
                                                    <Badge key={i} variant="secondary"
                                                           className="text-[10px]">{tag}</Badge>
                                                ))}
                                                {article.tags.length > 3 && (
                                                    <Badge variant="secondary"
                                                           className="text-[10px]">+{article.tags.length - 3}</Badge>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </Link>
                        ))}
                    </div>
                )}

                {/* Pagination */}
                {totalPages > 1 && (
                    <div className="flex items-center justify-center gap-2 pt-4">
                        <Button
                            variant="outline"
                            size="sm"
                            disabled={page <= 1}
                            onClick={() => setPage(p => Math.max(1, p - 1))}
                        >
                            Previous
                        </Button>
                        <span className="text-sm text-muted-foreground">
                            Page {page} of {totalPages}
                        </span>
                        <Button
                            variant="outline"
                            size="sm"
                            disabled={page >= totalPages}
                            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                        >
                            Next
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}
