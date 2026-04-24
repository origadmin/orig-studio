import React, {useState, useEffect} from 'react';
import {useParams} from '@tanstack/react-router';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Calendar, Settings, Bell, Heart, UserPlus, MessageSquare} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {userApi} from '@/lib/api/user';
import {mediaApi} from '@/lib/api/media';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import ErrorPage from '@/components/common/ErrorPage';
import SubscribeButton from '@/components/common/SubscribeButton';

const ProfilePage = () => {
    const params = useParams();
    const {t} = useTranslation();
    const [user, setUser] = useState<any>(null);
    const [videos, setVideos] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchUserAndVideos = async () => {
            try {
                setLoading(true);
                setError(null);

                let userResponse;
                if (params.username) {
                    // 通过 username 获取用户
                    userResponse = await userApi.getByUsername(params.username);
                } else if (params.id) {
                    // 通过 id 获取用户
                    userResponse = await userApi.get(params.id);
                } else {
                    throw new Error('No user identifier provided');
                }

                setUser(userResponse);

                // Fetch user videos
                const videosResponse = await mediaApi.list({user_id: userResponse.id});
                setVideos(videosResponse.items || []);
            } catch (err: any) {
                // Check if the error is due to user not found
                if (err.response && err.response.status === 404) {
                    // User not found, set user to null and don't set error
                    setUser(null);
                } else {
                    // Other error
                    setError('Failed to fetch user data');
                }
                console.error('Failed to fetch user data:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchUserAndVideos();
    }, [params.id, params.username]);

    return (
        <div className="space-y-8">
            {loading ? (
                <div className="flex justify-center items-center py-20">
                    <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-emerald-600"></div>
                </div>
            ) : error ? (
                <ErrorPage message={error}/>
            ) : user ? (
                <>
                    <div className="relative">
                        <div className="h-48 md:h-64 rounded-2xl bg-cover bg-center"
                             style={{backgroundImage: `url(${getImageUrl(user.cover, 'cover')})`}}/>
                        <div className="absolute -bottom-16 left-6 flex items-end gap-6">
                            <Avatar className="w-32 h-32 border-4 border-white dark:border-gray-900 shadow-lg">
                                <AvatarImage
                                    src={getImageUrl(user.avatar, 'avatar')}
                                    loading="lazy"
                                    onError={(e) => handleImageError(e, 'avatar')}/>
                                <AvatarFallback
                                    className="text-3xl">{user.username ? user.username.charAt(0) : 'U'}</AvatarFallback>
                            </Avatar>
                        </div>
                        <div className="absolute top-4 right-4">
                            {user.is_me ? (
                                <Button variant="outline" className="bg-white dark:bg-gray-800"><Settings
                                    className="w-4 h-4 mr-2"/>{t('common.editProfile')}</Button>
                            ) : (
                                <SubscribeButton
                                    channelId={user.channel_id || user.id || ''}
                                    initialSubscriberCount={user.subscriber_count || 0}
                                    className="bg-white dark:bg-gray-800"
                                />
                            )}
                        </div>
                    </div>

                    <div className="pt-20 px-6 space-y-4">
                        <div className="flex items-center gap-3">
                            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">{user.username}</h1>
                            {user.is_verified &&
                                <Badge variant="default" className="bg-emerald-500">{t('common.verified')}</Badge>}
                        </div>
                        <p className="text-slate-500 dark:text-gray-400">@{user.username}</p>
                        <p className="text-slate-600 dark:text-gray-300 max-w-2xl">{user.bio || t('profile.noBio')}</p>
                        <div className="flex flex-wrap gap-6 text-sm">
                            <div className="flex items-center gap-2"><UserPlus className="w-4 h-4 text-slate-400"/><span
                                className="font-semibold text-slate-900 dark:text-white">{formatViews(user.subscriber_count || 0)}</span><span
                                className="text-slate-500 dark:text-gray-400">{t('common.subscribers')}</span></div>
                            <div className="flex items-center gap-2"><Play className="w-4 h-4 text-slate-400"/><span
                                className="font-semibold text-slate-900 dark:text-white">{user.video_count || 0}</span><span
                                className="text-slate-500 dark:text-gray-400">{t('common.videos_count')}</span></div>
                            <div className="flex items-center gap-2"><Eye className="w-4 h-4 text-slate-400"/><span
                                className="font-semibold text-slate-900 dark:text-white">{formatViews(user.total_views || 0)}</span><span
                                className="text-slate-500 dark:text-gray-400">{t('common.views')}</span></div>
                            <div className="flex items-center gap-2"><Calendar className="w-4 h-4 text-slate-400"/><span
                                className="text-slate-500 dark:text-gray-400">{t('common.joinedAt', {date: formatDate(user.created_at || new Date().toISOString())})}</span>
                            </div>
                        </div>
                    </div>

                    <Tabs defaultValue="videos" className="w-full">
                        <TabsList
                            className="w-full justify-start border-b dark:border-gray-700 bg-transparent h-auto p-0">
                            {
                                [
                                    {v: 'videos', icon: <Play className="w-4 h-4 mr-2"/>, l: t('channel.tabVideos')},
                                    {v: 'playlists', l: t('profile.tabPlaylists')},
                                    {
                                        v: 'favorites',
                                        icon: <Heart className="w-4 h-4 mr-2"/>,
                                        l: t('profile.tabFavorites')
                                    },
                                    {
                                        v: 'community',
                                        icon: <MessageSquare className="w-4 h-4 mr-2"/>,
                                        l: t('profile.tabCommunity')
                                    },
                                ].map(t => (
                                    <TabsTrigger key={t.v} value={t.v}
                                                 className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-emerald-600 rounded-none px-4 py-3">
                                        {t.icon}{t.l}
                                    </TabsTrigger>
                                ))
                            }
                        </TabsList>
                        <TabsContent value="videos" className="mt-6">
                            {videos.length > 0 ? (
                                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                                    {videos.map(video => (
                                        <Link key={video.id} to="/watch" search={{v: video.short_token}}
                                              className="group">
                                            <div
                                                className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                                <div className="relative aspect-video">
                                                    <img src={getImageUrl(video.thumbnail, 'thumbnail')}
                                                         alt={video.title} loading="lazy"
                                                         onError={(e) => handleImageError(e, 'thumbnail')}
                                                         className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                                    <div
                                                        className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration || 0)}</div>
                                                </div>
                                                <div className="p-3">
                                                    <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                                    <p className="text-xs text-slate-500 dark:text-gray-400 mt-2">{formatViews(video.view_count || 0)} {t('common.views')}
                                                        · {formatDate(video.created_at || video.create_time || new Date().toISOString())}</p>
                                                </div>
                                            </div>
                                        </Link>
                                    ))}
                                </div>
                            ) : (
                                <div className="text-center py-12 text-slate-500 dark:text-gray-400">
                                    {t('profile.noVideos')}
                                </div>
                            )}
                        </TabsContent>
                        <TabsContent value="playlists" className="mt-6">
                            <div
                                className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noPlaylists')}</div>
                        </TabsContent>
                        <TabsContent value="favorites" className="mt-6">
                            <div
                                className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noFavorites')}</div>
                        </TabsContent>
                        <TabsContent value="community" className="mt-6">
                            <div
                                className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noCommunity')}</div>
                        </TabsContent>
                    </Tabs>
                </>
            ) : (
                <ErrorPage
                    statusCode={404}
                    title={t('profile.userNotFound')}
                    message={t('error.404Message')}
                />
            )}
        </div>
    );
};
export default ProfilePage;
