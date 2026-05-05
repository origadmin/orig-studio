/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Portal - Article View Page (X/Twitter Style)
 */

import {useState, useEffect, useMemo} from 'react';
import {useParams} from '@tanstack/react-router';
import {articleApi, type Article} from '@/lib/api/article';
import {publicMediaApi, type Media} from '@/lib/api/media';
import {userApi, type User as AuthorUser} from '@/lib/api/user';
import {API_BASE_URL} from '@/lib/request';
import {Spinner} from '@/components/ui/spinner';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarImage, AvatarFallback} from '@/components/ui/avatar';
import {AlertTriangle, Eye, Clock, ArrowLeft} from 'lucide-react';
import {formatDate, formatViews} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import VideoPlayer from '@/components/common/VideoPlayer';

function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

function renderMarkdown(content: string): string {
    let html = content
        .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h2>$1</h2>')
        .replace(/^# (.+)$/gm, '<h1>$1</h1>')
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
        .replace(/^- (.+)$/gm, '<li>$1</li>')
        .replace(/^(?!<[huplo]|<li|<pre|<code)(.+)$/gm, '<p>$1</p>')
        .replace(/\n\n/g, '');

    html = html.replace(/(<li>[\s\S]*?<\/li>)+/g, '<ul>$&</ul>');

    return html;
}

export default function ArticleViewPage() {
    const {t} = useTranslation();
    const {slug} = useParams({strict: false}) as {slug?: string};
    const [article, setArticle] = useState<Article | null>(null);
    const [media, setMedia] = useState<Media | null>(null);
    const [author, setAuthor] = useState<AuthorUser | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!slug) return;
        setLoading(true);
        setError(null);
        articleApi.getBySlug(slug)
            .then(data => {
                const articleData = (data as any)?.article ?? (data as any)?.data?.article ?? (data as any)?.data ?? data;
                setArticle(articleData);
                if (articleData.media_id && articleData.media?.short_token) {
                    publicMediaApi.get(articleData.media.short_token)
                        .then(mediaRes => {
                            const mediaData = (mediaRes as any)?.media ?? (mediaRes as any)?.data?.media ?? (mediaRes as any)?.data ?? mediaRes;
                            setMedia(mediaData);
                        })
                        .catch(err => {
                            console.error('Error loading media:', err);
                        });
                }
                if (articleData.user_id) {
                    userApi.get(String(articleData.user_id))
                        .then(authorData => {
                            const authorRes = (authorData as any)?.user ?? (authorData as any)?.data?.user ?? (authorData as any)?.data ?? authorData;
                            setAuthor(authorRes);
                        })
                        .catch(err => {
                            console.error('Error loading author:', err);
                        });
                }
            })
            .catch(err => {
                setError('Article not found');
                console.error('Error loading article:', err);
            })
            .finally(() => setLoading(false));
    }, [slug]);

    const isProcessing = media ? media.encoding_status !== 'success' : false;

    const renderedContent = useMemo(() => {
        if (!article?.content) return '';
        return renderMarkdown(article.content);
    }, [article?.content]);

    const displayThumbnail = resolveMediaUrl(article?.thumbnail || article?.media?.thumbnail);

    const authorName = author?.nickname || author?.username || (article?.user_id ? `User ${String(article.user_id).substring(0, 8)}` : 'Unknown');
    const authorAvatar = author?.avatar ? resolveMediaUrl(author.avatar) : undefined;
    const authorBio = author?.bio || '';

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
                <h2 className="text-xl font-semibold">{t('article.notFound')}</h2>
                <p className="text-muted-foreground">
                    {t('article.notFoundDesc')}
                </p>
                <Button variant="outline" onClick={() => window.history.back()}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>{t('article.goBack')}
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            <div className="max-w-3xl mx-auto px-6 py-12">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => window.history.back()}
                    className="mb-8 -ml-2 text-muted-foreground hover:text-foreground"
                >
                    <ArrowLeft className="w-4 h-4 mr-1"/>
                    {t('article.back')}
                </Button>

                <article>
                    <div className="space-y-4 mb-8">
                        <div className="flex items-center gap-2">
                            {article.state === 'published' && (
                                <Badge variant="default" className="text-xs">{t('article.published')}</Badge>
                            )}
                            {article.featured && (
                                <Badge variant="outline" className="text-warning border-amber-300 text-xs">
                                    {t('article.featured')}
                                </Badge>
                            )}
                        </div>

                        <h1 className="text-4xl font-bold tracking-tight leading-tight">
                            {article.title}
                        </h1>

                        <div className="flex items-center gap-3">
                            <Avatar className="w-10 h-10">
                                {authorAvatar ? (
                                    <AvatarImage src={authorAvatar} alt={authorName}/>
                                ) : null}
                                <AvatarFallback>{authorName[0]}</AvatarFallback>
                            </Avatar>
                            <div>
                                <p className="font-medium text-foreground">{authorName}</p>
                                <p className="text-sm text-muted-foreground flex items-center gap-2">
                                    <span className="flex items-center gap-1">
                                        <Clock className="w-3.5 h-3.5"/>
                                        {formatDate(article.published_at || article.create_time)}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Eye className="w-3.5 h-3.5"/>
                                        {formatViews(article.view_count)} {t('article.views')}
                                    </span>
                                </p>
                            </div>
                        </div>

                        {article.summary && (
                            <p className="text-lg text-muted-foreground leading-relaxed">
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

                    <div className="border-t mb-8"/>

                    {article.media_id && media && (
                        <div className="my-8 rounded-xl overflow-hidden shadow-sm">
                            <VideoPlayer
                                src={media.url || ''}
                                hlsSrc={media.hls_file}
                                poster={resolveMediaUrl(media.poster || media.thumbnail)}
                                isProcessing={isProcessing}
                            />
                        </div>
                    )}

                    {!article.media_id && displayThumbnail && (
                        <div className="my-8 rounded-xl overflow-hidden shadow-sm">
                            <img
                                src={displayThumbnail}
                                alt={article.title}
                                className="w-full h-auto object-cover"
                            />
                        </div>
                    )}

                    <div
                        className="prose prose-slate dark:prose-invert max-w-none leading-relaxed"
                        dangerouslySetInnerHTML={{__html: renderedContent}}
                    />

                    <div className="border-t mt-12 mb-8"/>

                    <div className="p-6 bg-muted/50 rounded-xl border">
                        <div className="flex items-center gap-4">
                            <Avatar className="w-16 h-16">
                                {authorAvatar ? (
                                    <AvatarImage src={authorAvatar} alt={authorName}/>
                                ) : null}
                                <AvatarFallback className="text-xl">{authorName[0]}</AvatarFallback>
                            </Avatar>
                            <div>
                                <p className="font-semibold text-lg">{authorName}</p>
                                {authorBio && (
                                    <p className="text-muted-foreground text-sm mt-1">{authorBio}</p>
                                )}
                            </div>
                        </div>
                    </div>
                </article>
            </div>
        </div>
    );
}
