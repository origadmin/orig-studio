/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Portal - Article View Page (Video Website Style)
 */

import {useState, useEffect, useMemo} from 'react';
import {useParams} from '@tanstack/react-router';
import {articleApi, type Article} from '@/lib/api/article';
import {API_BASE_URL} from '@/lib/request';
import {Spinner} from '@/components/ui/spinner';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {AlertTriangle, Play, Eye, MessageSquare, Clock, User, ArrowLeft} from 'lucide-react';
import {formatDateTime} from '@/lib/format';
import VideoPlayer from '@/components/common/VideoPlayer';

/**
 * Resolve a potentially relative URL to a full URL.
 */
function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

/**
 * Render markdown content to HTML (basic implementation).
 * For production, use a proper markdown library like marked or remark.
 */
function renderMarkdown(content: string): string {
    // Basic markdown rendering - headers, bold, italic, links, code, lists
    let html = content
        // Code blocks
        .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
        // Inline code
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        // Headers
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h2>$1</h2>')
        .replace(/^# (.+)$/gm, '<h1>$1</h1>')
        // Bold
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        // Italic
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        // Links
        .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
        // Unordered lists
        .replace(/^- (.+)$/gm, '<li>$1</li>')
        // Paragraphs (lines not already wrapped)
        .replace(/^(?!<[huplo]|<li|<pre|<code)(.+)$/gm, '<p>$1</p>')
        // Line breaks
        .replace(/\n\n/g, '');

    // Wrap consecutive <li> in <ul>
    html = html.replace(/(<li>[\s\S]*?<\/li>)+/g, '<ul>$&</ul>');

    return html;
}

export default function ArticleViewPage() {
    const {slug} = useParams({strict: false}) as {slug?: string};
    const [article, setArticle] = useState<Article | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!slug) return;
        setLoading(true);
        setError(null);
        articleApi.getBySlug(slug)
            .then(data => {
                setArticle(data);
            })
            .catch(err => {
                setError('Article not found');
                console.error('Error loading article:', err);
            })
            .finally(() => setLoading(false));
    }, [slug]);

    // Resolve video source URL
    const videoSrc = useMemo(() => {
        if (!article?.media?.short_token) return undefined;
        return `${API_BASE_URL}/stream/${article.media.short_token}/index.m3u8`;
    }, [article?.media?.short_token]);

    // Rendered markdown content
    const renderedContent = useMemo(() => {
        if (!article?.content) return '';
        return renderMarkdown(article.content);
    }, [article?.content]);

    // Display thumbnail
    const displayThumbnail = resolveMediaUrl(article?.thumbnail || article?.media?.thumbnail);

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[80vh]">
                <Spinner/>
            </div>
        );
    }

    if (error || !article) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[80vh] gap-4">
                <AlertTriangle className="w-12 h-12 text-muted-foreground"/>
                <h2 className="text-xl font-semibold">Article Not Found</h2>
                <p className="text-muted-foreground">
                    The article you are looking for does not exist or is not publicly available.
                </p>
                <Button variant="outline" onClick={() => window.history.back()}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>Go Back
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            {/* Video Player (full width, if media associated) */}
            {article.media_id && article.media && videoSrc && (
                <div className="w-full bg-black">
                    <div className="max-w-5xl mx-auto">
                        <VideoPlayer
                            src={videoSrc}
                            poster={displayThumbnail}
                        />
                    </div>
                </div>
            )}

            {/* Cover image (if no video but has thumbnail) */}
            {!article.media_id && displayThumbnail && (
                <div className="w-full max-h-[400px] overflow-hidden">
                    <img src={displayThumbnail} alt={article.title}
                         className="w-full h-full object-cover"/>
                </div>
            )}

            <div className="max-w-5xl mx-auto px-4 py-8">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Main content area */}
                    <div className="lg:col-span-2 space-y-6">
                        {/* Article header */}
                        <div className="space-y-4">
                            <div className="flex items-center gap-2">
                                <Badge variant={article.state === 'published' ? 'default' : 'secondary'}>
                                    {article.state}
                                </Badge>
                                {article.featured && (
                                    <Badge variant="outline" className="text-warning border-amber-300">
                                        Featured
                                    </Badge>
                                )}
                            </div>
                            <h1 className="text-3xl font-bold tracking-tight">{article.title}</h1>
                            <div className="flex items-center gap-4 text-sm text-muted-foreground">
                                <div className="flex items-center gap-1">
                                    <User className="w-4 h-4"/>
                                    <span>{article.user_id}</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Clock className="w-4 h-4"/>
                                    <span>{formatDateTime(article.create_time)}</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Eye className="w-4 h-4"/>
                                    <span>{article.view_count} views</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <MessageSquare className="w-4 h-4"/>
                                    <span>{article.comment_count} comments</span>
                                </div>
                            </div>
                            {article.summary && (
                                <p className="text-muted-foreground text-lg leading-relaxed">
                                    {article.summary}
                                </p>
                            )}
                            {article.tags && article.tags.length > 0 && (
                                <div className="flex flex-wrap gap-1.5">
                                    {article.tags.map((tag, i) => (
                                        <Badge key={i} variant="secondary" className="text-xs">{tag}</Badge>
                                    ))}
                                </div>
                            )}
                        </div>

                        <div className="border-t"/>

                        {/* Article content */}
                        <div
                            className="prose prose-slate dark:prose-invert max-w-none"
                            dangerouslySetInnerHTML={{__html: renderedContent}}
                        />
                    </div>

                    {/* Right sidebar */}
                    <div className="space-y-6">
                        {/* Author card */}
                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">Author</h3>
                            <div className="flex items-center gap-3">
                                <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
                                    <User className="w-5 h-5 text-muted-foreground"/>
                                </div>
                                <div>
                                    <p className="text-sm font-medium">User {article.user_id.substring(0, 8)}</p>
                                    <p className="text-xs text-muted-foreground">Content Creator</p>
                                </div>
                            </div>
                        </div>

                        {/* Article info */}
                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">Article Info</h3>
                            <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
                                <span className="text-muted-foreground">Published</span>
                                <span className="text-xs text-right whitespace-nowrap">
                                    {article.published_at ? formatDateTime(article.published_at) : 'Not published'}
                                </span>
                                <span className="text-muted-foreground">Updated</span>
                                <span className="text-xs text-right whitespace-nowrap">{formatDateTime(article.update_time)}</span>
                                <span className="text-muted-foreground">Views</span>
                                <span className="text-xs text-right">{article.view_count}</span>
                            </div>
                        </div>

                        {/* Related video */}
                        {article.media && (
                            <div className="bg-card rounded-lg border p-4 space-y-3">
                                <h3 className="font-medium">Related Video</h3>
                                <div className="relative aspect-video bg-muted rounded-md overflow-hidden">
                                    {resolveMediaUrl(article.media.thumbnail) ? (
                                        <img src={resolveMediaUrl(article.media.thumbnail)} alt={article.media.title}
                                             className="w-full h-full object-cover" loading="lazy"/>
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center">
                                            <Play className="w-8 h-8 text-muted-foreground"/>
                                        </div>
                                    )}
                                    {article.media.duration > 0 && (
                                        <span className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1.5 py-0.5 rounded">
                                            {Math.floor(article.media.duration / 60)}:{String(Math.floor(article.media.duration % 60)).padStart(2, '0')}
                                        </span>
                                    )}
                                </div>
                                <p className="text-sm font-medium">{article.media.title}</p>
                                {article.media.short_token && (
                                    <Button variant="outline" size="sm" className="w-full"
                                            onClick={() => window.open(`/watch?v=${article.media!.short_token}`, '_blank')}>
                                        <Play className="w-3 h-3 mr-1"/>Watch Video
                                    </Button>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
