import React, {useState} from 'react';
import {useQuery} from '@tanstack/react-query';
import {Link, useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {articleApi} from '@/lib/api/article';
import {Avatar, AvatarFallback} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Spinner} from '@/components/ui/spinner';
import {FileText, Clock, User, Plus} from 'lucide-react';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {formatDate} from '@/lib/format';
import {CreateChannelDialog} from '@/components/channel/CreateChannelDialog';
import {useAuth} from '@/hooks/useAuth';

const ArticleHomeLayout: React.FC = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const {isAuthenticated} = useAuth();
    const [createDialogOpen, setCreateDialogOpen] = useState(false);

    const {data, isLoading, error} = useQuery({
        queryKey: ['articles', 'latest'],
        queryFn: () => articleApi.list({page: 1, page_size: 20, state: 'published'}),
        staleTime: 30_000,
    });

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <Spinner className="h-8 w-8"/>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-12 text-muted-foreground">
                Failed to load articles. Please try again later.
            </div>
        );
    }

    const articles = data?.items ?? [];

    if (articles.length === 0) {
        return (
            <div className="text-center py-12">
                <FileText className="h-12 w-12 mx-auto mb-4 text-muted-foreground"/>
                <h2 className="text-xl font-semibold mb-2">{t('channel.noArticles') || 'No Articles Yet'}</h2>
                <p className="text-muted-foreground mb-6">{t('channel.noArticlesSelf') || 'Check back later for new content.'}</p>
                {isAuthenticated && (
                    <>
                        <Button onClick={() => setCreateDialogOpen(true)} size="lg" className="mb-2">
                            <Plus size={16} className="mr-2"/>
                            {t('channel.create.title') || 'Create Channel'}
                        </Button>
                        <CreateChannelDialog
                            open={createDialogOpen}
                            onOpenChange={setCreateDialogOpen}
                            onSuccess={({short_token: newToken}) => {
                                if (newToken) {
                                    navigate({to: '/c/$id', params: {id: newToken}});
                                }
                            }}
                        />
                    </>
                )}
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold">{t('home.latestArticles')}</h1>
                <p className="text-muted-foreground text-sm">{t('home.latestArticlesDesc')}</p>
            </div>

            <div className="divide-y divide-border">
                {articles.map((article) => (
                    <Link
                        key={article.id}
                        to="/articles/$slug"
                        params={{slug: article.slug}}
                        className="block py-4 hover:bg-muted/50 -mx-2 px-2 rounded-lg transition-colors"
                    >
                        <div className="flex gap-3">
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                    <Avatar className="h-5 w-5">
                                        <AvatarFallback className="text-[10px]">
                                            <User className="h-3 w-3"/>
                                        </AvatarFallback>
                                    </Avatar>
                                    <span className="text-sm font-medium truncate">
                                        {article.user_id ? `User` : 'Anonymous'}
                                    </span>
                                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                                        <Clock className="h-3 w-3"/>
                                        {formatDate(article.create_time)}
                                    </span>
                                </div>
                                <h3 className="font-semibold text-base leading-snug mb-1 line-clamp-2">
                                    {article.title}
                                </h3>
                                {article.summary && (
                                    <p className="text-sm text-muted-foreground line-clamp-2">
                                        {article.summary}
                                    </p>
                                )}
                                <div className="flex items-center gap-2 mt-2">
                                    {article.tags?.slice(0, 3).map((tag, idx) => (
                                        <Badge key={typeof tag === 'string' ? tag : idx} variant="outline" className="text-xs">
                                            #{typeof tag === 'string' ? tag : tag}
                                        </Badge>
                                    ))}
                                </div>
                            </div>
                            {article.thumbnail && (
                                <div className="flex-shrink-0">
                                    <img
                                        src={getImageUrl(article.thumbnail, 'cover')}
                                        alt=""
                                        onError={(e) => handleImageError(e, 'cover')}
                                        className="w-24 h-24 object-cover rounded-lg"
                                    />
                                </div>
                            )}
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
};

export default ArticleHomeLayout;
