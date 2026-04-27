import {Spinner} from "@/components/ui/spinner"
import React from 'react';
import {
    Film,
    Users,
    Eye,
    Heart,
    BarChart3,
    TrendingUp,
    TrendingDown,
    Minus,
    Loader2,
    MessageCircle,
    DollarSign
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useQuery} from '@tanstack/react-query';
import {statsApi, DashboardStats} from '@/lib/api/stats';
import {Link} from '@tanstack/react-router';
import {Card, CardContent} from '@/components/ui/card';

const Dashboard = () => {
    const {t} = useTranslation();

    const {data, isLoading, error} = useQuery({
        queryKey: ['admin', 'stats'],
        queryFn: async () => {
            return await statsApi.getDashboard();
        }
    });

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-20 text-muted-foreground">
                <p className="text-lg mb-1">{t('common.loading')}</p>
                <p className="text-sm">{(error as Error).message}</p>
            </div>
        );
    }

    const stats: DashboardStats = data || {
        total_media: 0,
        total_users: 0,
        total_views: 0,
        total_comments: 0,
        total_subscribers: 0,
        total_revenue: 0,
        active_users: 0,
        new_users_today: 0,
        new_media_today: 0,
        new_views_today: 0,
        new_comments_today: 0,
        new_subscribers_today: 0,
        media_by_type: {video: 0, image: 0, audio: 0, other: 0},
        users_by_role: {admin: 0, editor: 0, user: 0},
        views_by_date: [],
        media_by_date: [],
        top_categories: [],
        top_creators: [],
        top_media: []
    };

    const formatNumber = (num: number | undefined | null) => {
        if (num === undefined || num === null) return '0';
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'k';
        }
        return num.toString();
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* Header Card */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h1 className="text-3xl font-extrabold text-slate-900 dark:text-slate-50">{t('admin.dashboard')}</h1>
                                <p className="text-slate-500 text-sm mt-1">{t('admin.dashboardDesc') || 'Overview of your platform performance'}</p>
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                >
                                    {t('admin.exportReport') || 'Export Report'}
                                </Button>
                                <Link to="/admin/content">
                                    <Button
                                        variant="default"
                                        size="sm"
                                    >
                                        {t('admin.manageContent') || 'Manage Content'}
                                    </Button>
                                </Link>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatCard
                    icon={<Film className="h-6 w-6"/>}
                    label={t('admin.totalMedia') || 'Total Medias'}
                    value={formatNumber(stats.total_media)}
                    trend={`+${stats.new_media_today} today`}
                    trendUp={stats.new_media_today > 0}
                    color="sky"
                />
                <StatCard
                    icon={<Users className="h-6 w-6"/>}
                    label={t('admin.totalUsers') || 'Total Users'}
                    value={formatNumber(stats.total_users)}
                    trend={`+${stats.new_users_today} today`}
                    trendUp={stats.new_users_today > 0}
                    color="emerald"
                />
                <StatCard
                    icon={<Eye className="h-6 w-6"/>}
                    label={t('admin.totalViews') || 'Total Views'}
                    value={formatNumber(stats.total_views)}
                    trend={`+${formatNumber(stats.new_views_today)} today`}
                    trendUp={stats.new_views_today > 0}
                    color="purple"
                />
                <StatCard
                    icon={<MessageCircle className="h-6 w-6"/>}
                    label={t('admin.totalComments') || 'Total Comments'}
                    value={formatNumber(stats.total_comments)}
                    trend={`+${stats.new_comments_today} today`}
                    trendUp={stats.new_comments_today > 0}
                    color="red"
                />
            </div>

            {/* Secondary Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <StatCard
                    icon={<Heart className="h-6 w-6"/>}
                    label={t('admin.totalSubscribers') || 'Total Subscribers'}
                    value={formatNumber(stats.total_subscribers)}
                    trend={`+${stats.new_subscribers_today} today`}
                    trendUp={stats.new_subscribers_today > 0}
                    color="pink"
                    small
                />
                <StatCard
                    icon={<DollarSign className="h-6 w-6"/>}
                    label={t('admin.totalRevenue') || 'Total Revenue'}
                    value={`$${formatNumber(stats.total_revenue)}`}
                    trend="+12% this month"
                    trendUp={true}
                    color="amber"
                    small
                />
                <StatCard
                    icon={<Users className="h-6 w-6"/>}
                    label={t('admin.activeUsers') || 'Active Users'}
                    value={formatNumber(stats.active_users)}
                    trend="Currently online"
                    trendUp={true}
                    color="cyan"
                    small
                />
            </div>

            {/* Content Breakdown */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Media by Type */}
                <Card className="shadow-sm relative overflow-hidden">
                    <CardContent className="p-6">
                        <h3 className="font-bold text-slate-900 dark:text-slate-50 mb-6 flex items-center gap-2">
                            <BarChart3 size={20} className="text-info"/>
                            {t('admin.mediaByType') || 'Media by Type'}
                        </h3>
                        <div className="space-y-4">
                            <TypeBar label="Videos" count={stats.media_by_type?.video || 0} total={stats.total_media}
                                     color="bg-info"/>
                            <TypeBar label="Images" count={stats.media_by_type?.image || 0} total={stats.total_media}
                                     color="bg-success"/>
                            <TypeBar label="Audio" count={stats.media_by_type?.audio || 0} total={stats.total_media}
                                     color="bg-purple-500"/>
                            <TypeBar label="Other" count={stats.media_by_type?.other || 0} total={stats.total_media}
                                     color="bg-gray-500"/>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                    </CardContent>
                </Card>

                {/* Users by Role */}
                <Card className="shadow-sm relative overflow-hidden">
                    <CardContent className="p-6">
                        <h3 className="font-bold text-slate-900 dark:text-slate-50 mb-6 flex items-center gap-2">
                            <Users size={20} className="text-destructive"/>
                            {t('admin.usersByRole') || 'Users by Role'}
                        </h3>
                        <div className="space-y-4">
                            <TypeBar label="Admins" count={stats.users_by_role?.admin || 0} total={stats.total_users}
                                     color="bg-destructive"/>
                            <TypeBar label="Editors" count={stats.users_by_role?.editor || 0} total={stats.total_users}
                                     color="bg-amber-500"/>
                            <TypeBar label="Users" count={stats.users_by_role?.user || 0} total={stats.total_users}
                                     color="bg-emerald-500"/>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-destructive w-full opacity-10"/>
                    </CardContent>
                </Card>

                {/* Trending Content */}
                <Card className="shadow-sm relative overflow-hidden">
                    <CardContent className="p-6">
                        <h3 className="font-bold text-slate-900 dark:text-slate-50 mb-6 flex items-center gap-2">
                            <TrendingUp size={20} className="text-success"/>
                            {t('admin.trendingContent') || 'Trending Content'}
                        </h3>
                        <div className="space-y-4">
                            {stats.top_media?.slice(0, 5).map((item: any, index: number) => (
                                <TrendingItem
                                    key={index}
                                    title={item.title}
                                    views={formatNumber(item.views)}
                                    index={index + 1}
                                />
                            ))}
                            {(!stats.top_media || stats.top_media.length === 0) && (
                                <p className="text-sm text-gray-500 text-center py-4">No trending content yet</p>
                            )}
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                    </CardContent>
                </Card>
            </div>

            {/* Top Categories & Creators */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Top Categories */}
                <Card className="shadow-sm relative overflow-hidden">
                    <CardContent className="p-6">
                        <h3 className="font-bold text-slate-900 dark:text-slate-50 mb-6 flex items-center gap-2">
                            <BarChart3 size={20} className="text-emerald-500"/>
                            {t('admin.topCategories') || 'Top Categories'}
                        </h3>
                        <div className="absolute bottom-0 left-0 h-1 bg-emerald-500 w-full opacity-10"/>
                        <div className="space-y-3">
                            {stats.top_categories?.map((category: any, index: number) => (
                                <div key={category.id}
                                     className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                                    <div className="flex items-center gap-3">
                                        <span
                                            className="w-6 h-6 flex items-center justify-center bg-emerald-100 dark:bg-emerald-900 text-emerald-600 dark:text-emerald-400 text-xs font-bold rounded">
                                            {index + 1}
                                        </span>
                                        <span
                                            className="font-medium text-slate-800 dark:text-gray-200">{category.name}</span>
                                    </div>
                                    <span className="text-sm text-gray-500">{category.count} items</span>
                                </div>
                            ))}
                            {(!stats.top_categories || stats.top_categories.length === 0) && (
                                <p className="text-sm text-gray-500 text-center py-4">No categories yet</p>
                            )}
                        </div>
                    </CardContent>
                </Card>

                {/* Top Creators */}
                <Card className="shadow-sm relative overflow-hidden">
                    <CardContent className="p-6">
                        <h3 className="font-bold text-slate-900 dark:text-slate-50 mb-6 flex items-center gap-2">
                            <Users size={20} className="text-info"/>
                            {t('admin.topCreators') || 'Top Creators'}
                        </h3>
                        <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                        <div className="space-y-3">
                            {stats.top_creators?.map((creator: any, index: number) => (
                                <div key={creator.id}
                                     className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                                    <div className="flex items-center gap-3">
                                        <span
                                            className="w-6 h-6 flex items-center justify-center bg-blue-100 dark:bg-blue-900 text-info dark:text-blue-400 text-xs font-bold rounded">
                                            {index + 1}
                                        </span>
                                        <div>
                                            <span
                                                className="font-medium text-slate-800 dark:text-gray-200 block">{creator.name}</span>
                                            <span className="text-xs text-gray-500">{creator.media_count} videos</span>
                                        </div>
                                    </div>
                                    <span className="text-sm text-gray-500">{formatNumber(creator.views)} views</span>
                                </div>
                            ))}
                            {(!stats.top_creators || stats.top_creators.length === 0) && (
                                <p className="text-sm text-gray-500 text-center py-4">No creators yet</p>
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};

const StatCard = ({icon, label, value, trend, trendUp, small = false, color = "primary"}: any) => {
    // 定义颜色映射
    const colorMap = {
        primary: {
            bg: 'bg-blue-50 dark:bg-blue-950/30',
            text: 'text-info dark:text-blue-400',
            bar: 'bg-info'
        },
        pink: {
            bg: 'bg-pink-50 dark:bg-pink-950/30',
            text: 'text-pink-500 dark:text-pink-400',
            bar: 'bg-pink-500'
        },
        cyan: {
            bg: 'bg-cyan-50 dark:bg-cyan-950/30',
            text: 'text-cyan-500 dark:text-cyan-400',
            bar: 'bg-cyan-500'
        },
        amber: {
            bg: 'bg-amber-50 dark:bg-amber-950/30',
            text: 'text-amber-500 dark:text-amber-400',
            bar: 'bg-amber-500'
        },
        green: {
            bg: 'bg-green-50 dark:bg-green-950/30',
            text: 'text-success dark:text-green-400',
            bar: 'bg-success'
        },
        sky: {
            bg: 'bg-sky-50 dark:bg-sky-950/30',
            text: 'text-sky-500 dark:text-sky-400',
            bar: 'bg-sky-500'
        },
        emerald: {
            bg: 'bg-emerald-50 dark:bg-emerald-950/30',
            text: 'text-emerald-500 dark:text-emerald-400',
            bar: 'bg-emerald-500'
        },
        purple: {
            bg: 'bg-purple-50 dark:bg-purple-950/30',
            text: 'text-purple-500 dark:text-purple-400',
            bar: 'bg-purple-500'
        },
        red: {
            bg: 'bg-red-50 dark:bg-red-950/30',
            text: 'text-destructive dark:text-red-400',
            bar: 'bg-destructive'
        }
    };
    
    // 确保 color 是一个有效的颜色值
    const validColor = colorMap[color] || colorMap.primary;
    
    return (
        <Card className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
            <CardContent className={`p-5 ${small ? 'p-4' : ''}`}>
                <div className="flex items-start justify-between">
                    <div className="space-y-1">
                        <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground dark:text-slate-500">{label}</p>
                        <h3 className={`text-3xl font-bold tabular-nums ${validColor.text} ${small ? 'text-xl' : ''}`}>{value}</h3>
                    </div>
                    <div
                        className={`p-2.5 rounded-xl ${validColor.bg}`}>
                        {React.cloneElement(icon, {
                            className: `h-6 w-6 ${validColor.text}`
                        })}
                    </div>
                </div>
                <div
                    className={`mt-4 pt-4 border-t border-gray-50 dark:border-gray-700 flex items-center text-[11px] font-semibold ${trendUp === true ? 'text-success' : trendUp === false ? 'text-destructive' : 'text-gray-500'}`}>
                    {trendUp === true ? <TrendingUp size={14} className="mr-1"/> : trendUp === false ? <TrendingDown size={14} className="mr-1"/> : <Minus size={14} className="mr-1"/>}
                    {trend}
                </div>
                {/* 确保底部色条正确显示 */}
                <div className={`absolute bottom-0 left-0 h-1 ${validColor.bar} w-full opacity-10 z-10`}/>
            </CardContent>
        </Card>
    );
};

const TypeBar = ({label, count, total, color}: { label: string, count: number, total: number, color: string }) => {
    const percentage = total > 0 ? (count / total) * 100 : 0;
    return (
        <div className="space-y-1">
            <div className="flex justify-between text-sm">
                <span className="text-slate-700 dark:text-gray-300">{label}</span>
                <span className="font-medium text-slate-900 dark:text-white">{count}</span>
            </div>
            <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                <div className={`h-full ${color} rounded-full transition-all duration-500`}
                     style={{width: `${percentage}%`}}/>
            </div>
        </div>
    );
};

const TrendingItem = ({title, views, index}: { title: string, views: string, index: number }) => (
    <div className="flex items-center gap-3 p-2 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">
        <span
            className="w-6 h-6 flex items-center justify-center bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-muted-foreground text-xs font-bold rounded shrink-0">
            {index}
        </span>
        <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-slate-800 dark:text-gray-200 truncate">{title}</p>
            <p className="text-xs text-slate-500">{views} views</p>
        </div>
        <TrendingUp size={16} className="text-success shrink-0"/>
    </div>
);

export default Dashboard;
