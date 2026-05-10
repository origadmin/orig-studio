import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {createFileRoute} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Skeleton} from '@/components/ui/skeleton';
import {ArrowLeft} from 'lucide-react';
import {Link} from '@tanstack/react-router';
import {api} from '@/lib/request';
import type {CustomPage} from '@/lib/api/portal';
import {useModuleState} from '@/contexts/ModuleConfigContext';

export default function CustomPage() {
    const route = createFileRoute('/_portal/p/$slug')();
    const {slug} = route.useParams();
    const {t} = useTranslation();
    const [page, setPage] = useState<CustomPage | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchPage = async () => {
            setLoading(true);
            setError(null);
            try {
                const res = await api.get<CustomPage>(`/p/${slug}`);
                setPage(res);
            } catch (err) {
                console.error('Failed to fetch custom page:', err);
                setError('Page not found');
            } finally {
                setLoading(false);
            }
        };

        if (slug) {
            fetchPage();
        }
    }, [slug]);

    if (loading) {
        return (
            <div className="container mx-auto px-4 py-8 max-w-4xl">
                <Skeleton className="h-10 w-64 mb-4"/>
                <Skeleton className="h-4 w-full mb-2"/>
                <Skeleton className="h-4 w-full mb-2"/>
                <Skeleton className="h-4 w-3/4"/>
                <div className="mt-8">
                    <Skeleton className="h-64 w-full"/>
                </div>
            </div>
        );
    }

    if (error || !page) {
        return (
            <div className="container mx-auto px-4 py-16 text-center">
                <h1 className="text-4xl font-bold mb-4">{t('error.pageNotFound')}</h1>
                <p className="text-muted-foreground mb-8">{error || 'The page you are looking for does not exist.'}</p>
                <Link to="/">
                    <Button>
                        <ArrowLeft className="mr-2 h-4 w-4"/>
                        {t('common.backToHome')}
                    </Button>
                </Link>
            </div>
        );
    }

    const renderContent = () => {
        switch (page.content_format) {
            case 'markdown':
                return (
                    <div
                        className="prose prose-lg dark:prose-invert max-w-none"
                        dangerouslySetInnerHTML={{__html: parseMarkdown(page.content)}}
                    />
                );
            case 'html':
                return (
                    <div
                        className="custom-html-content"
                        dangerouslySetInnerHTML={{__html: page.content}}
                    />
                );
            default:
                return (
                    <p className="whitespace-pre-wrap">{page.content}</p>
                );
        }
    };

    return (
        <div className="container mx-auto px-4 py-8">
            <article className={`max-w-4xl mx-auto ${page.layout === 'full_width' ? 'max-w-none' : ''}`}>
                <header className="mb-8">
                    <Link
                        to="/"
                        className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4 transition-colors"
                    >
                        <ArrowLeft className="mr-1 h-4 w-4"/>
                        {t('common.backToHome')}
                    </Link>

                    <div className="flex items-center gap-2 mb-4">
                        <Badge variant="secondary">{page.type}</Badge>
                        {page.is_published ? (
                            <Badge variant="default">{t('common.published')}</Badge>
                        ) : (
                            <Badge variant="outline">{t('common.draft')}</Badge>
                        )}
                    </div>

                    <h1 className="text-4xl font-bold mb-4">{page.title}</h1>

                    {page.seo_description && (
                        <p className="text-lg text-muted-foreground">{page.seo_description}</p>
                    )}

                    <div className="flex items-center gap-4 text-sm text-muted-foreground mt-4">
                        {page.published_at && (
                            <span>
                                {t('common.publishedOn')}: {new Date(page.published_at).toLocaleDateString()}
                            </span>
                        )}
                        <span>
                            {page.view_count} {t('common.views')}
                        </span>
                    </div>
                </header>

                {page.featured_image && (
                    <div className="mb-8">
                        <img
                            src={page.featured_image}
                            alt={page.title}
                            className="w-full h-auto rounded-lg object-cover max-h-96"
                        />
                    </div>
                )}

                <div className={`${page.layout === 'sidebar' ? 'grid grid-cols-1 md:grid-cols-4 gap-8' : ''}`}>
                    <div className={page.layout === 'sidebar' ? 'md:col-span-3' : ''}>
                        <div className="bg-card rounded-lg p-6 md:p-8">
                            {renderContent()}
                        </div>
                    </div>

                    {page.layout === 'sidebar' && (
                        <aside className="md:col-span-1">
                            <div className="sticky top-24 bg-muted/50 rounded-lg p-4">
                                <h3 className="font-semibold mb-2">{page.title}</h3>
                                <p className="text-sm text-muted-foreground">{page.seo_description}</p>
                            </div>
                        </aside>
                    )}
                </div>
            </article>
        </div>
    );
}

function parseMarkdown(content: string): string {
    let html = content;

    html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
    html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
    html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');

    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-primary hover:underline">$1</a>');

    html = html.replace(/\n\n/g, '</p><p>');
    html = '<p>' + html + '</p>';

    html = html.replace(/<p><\/p>/g, '');

    return html;
}
