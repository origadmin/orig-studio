import React from 'react';
import {
    Film,
    Users,
    Eye,
    Heart,
    BarChart3,
    TrendingUp,
    TrendingDown,
    Loader2,
    MessageCircle,
    DollarSign
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useQuery} from '@tanstack/react-query';
import {statsApi, DashboardStats} from '@/lib/api/stats';
import {Link} from '@tanstack/react-router';

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
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-20 text-gray-400">
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
        <div className="space-y-8">
            {/* Header */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white">{t('admin.dashboard')}</h1>
                    <p className="text-slate-500 text-sm mt-1">{t('admin.dashboardDesc') || 'Overview of your platform performance'}</p>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-semibold hover:bg-gray-50 dark:hover:bg-gray-700 shadow-sm transition-colors">
                        {t('admin.exportReport') || 'Export Report'}
                    </Button>
                    <Link to="/admin/content">
                        <Button
                            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-semibold shadow-sm transition-colors">
                            {t('admin.manageContent') || 'Manage Content'}
                        </Button>
                    </Link>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    icon={<Film className="text-blue-500"/>}
                    label={t('admin.totalMedia') || 'Total Medias'}
                    value={formatNumber(stats.total_media)}
                    trend={`+${stats.new_media_today} today`}
                    trendUp={stats.new_media_today > 0}
                />
                <StatCard
                    icon={<Users className="text-green-500"/>}
                    label={t('admin.totalUsers') || 'Total Users'}
                    value={formatNumber(stats.total_users)}
                    trend={`+${stats.new_users_today} today`}
                    trendUp={stats.new_users_today > 0}
                />
                <StatCard
                    icon={<Eye className="text-purple-500"/>}
                    label={t('admin.totalViews') || 'Total Views'}
                    value={formatNumber(stats.total_views)}
                    trend={`+${formatNumber(stats.new_views_today)} today`}
                    trendUp={stats.new_views_today > 0}
                />
                <StatCard
                    icon={<MessageCircle className="text-rose-500"/>}
                    label={t('admin.totalComments') || 'Total Comments'}
                    value={formatNumber(stats.total_comments)}
                    trend={`+${stats.new_comments_today} today`}
                    trendUp={stats.new_comments_today > 0}
                />
            </div>

            {/* Secondary Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                <StatCard
                    icon={<Heart className="text-pink-500"/>}
                    label={t('admin.totalSubscribers') || 'Total Subscribers'}
                    value={formatNumber(stats.total_subscribers)}
                    trend={`+${stats.new_subscribers_today} today`}
                    trendUp={stats.new_subscribers_today > 0}
                    small
                />
                <StatCard
                    icon={<DollarSign className="text-amber-500"/>}
                    label={t('admin.totalRevenue') || 'Total Revenue'}
                    value={`$${formatNumber(stats.total_revenue)}`}
                    trend="+12% this month"
                    trendUp={true}
                    small
                />
                <StatCard
                    icon={<Users className="text-cyan-500"/>}
                    label={t('admin.activeUsers') || 'Active Users'}
                    value={formatNumber(stats.active_users)}
                    trend="Currently online"
                    trendUp={true}
                    small
                />
            </div>

            {/* Content Breakdown */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Media by Type */}
                <div
                    className="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm">
                    <h3 className="font-bold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
                        <BarChart3 size={20} className="text-slate-400"/>
                        {t('admin.mediaByType') || 'Media by Type'}
                    </h3>
                    <div className="space-y-4">
                        <TypeBar label="Videos" count={stats.media_by_type?.video || 0} total={stats.total_media}
                                 color="bg-blue-500"/>
                        <TypeBar label="Images" count={stats.media_by_type?.image || 0} total={stats.total_media}
                                 color="bg-green-500"/>
                        <TypeBar label="Audio" count={stats.media_by_type?.audio || 0} total={stats.total_media}
                                 color="bg-purple-500"/>
                        <TypeBar label="Other" count={stats.media_by_type?.other || 0} total={stats.total_media}
                                 color="bg-gray-500"/>
                    </div>
                </div>

                {/* Users by Role */}
                <div
                    className="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm">
                    <h3 className="font-bold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
                        <Users size={20} className="text-slate-400"/>
                        {t('admin.usersByRole') || 'Users by Role'}
                    </h3>
                    <div className="space-y-4">
                        <TypeBar label="Admins" count={stats.users_by_role?.admin || 0} total={stats.total_users}
                                 color="bg-red-500"/>
                        <TypeBar label="Editors" count={stats.users_by_role?.editor || 0} total={stats.total_users}
                                 color="bg-amber-500"/>
                        <TypeBar label="Users" count={stats.users_by_role?.user || 0} total={stats.total_users}
                                 color="bg-emerald-500"/>
                    </div>
                </div>

                {/* Trending Content */}
                <div
                    className="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm">
                    <h3 className="font-bold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
                        <TrendingUp size={20} className="text-slate-400"/>
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
                </div>
            </div>

            {/* Top Categories & Creators */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Top Categories */}
                <div
                    className="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm">
                    <h3 className="font-bold text-slate-900 dark:text-white mb-6">
                        {t('admin.topCategories') || 'Top Categories'}
                    </h3>
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
                </div>

                {/* Top Creators */}
                <div
                    className="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm">
                    <h3 className="font-bold text-slate-900 dark:text-white mb-6">
                        {t('admin.topCreators') || 'Top Creators'}
                    </h3>
                    <div className="space-y-3">
                        {stats.top_creators?.map((creator: any, index: number) => (
                            <div key={creator.id}
                                 className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                                <div className="flex items-center gap-3">
                                    <span
                                        className="w-6 h-6 flex items-center justify-center bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400 text-xs font-bold rounded">
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
                </div>
            </div>
        </div>
    );
};

const StatCard = ({icon, label, value, trend, trendUp = true, small = false}: any) => (
    <div
        className={`bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow ${small ? 'p-4' : ''}`}>
        <div className="flex items-center space-x-4">
            <div className={`p-3 bg-gray-50 dark:bg-gray-700 rounded-xl ${small ? 'p-2' : ''}`}>{icon}</div>
            <div className="flex-1">
                <p className="text-sm font-medium text-slate-500 dark:text-gray-400">{label}</p>
                <h3 className={`font-extrabold text-slate-900 dark:text-white mt-0.5 ${small ? 'text-xl' : 'text-2xl'}`}>{value}</h3>
            </div>
        </div>
        <div
            className={`mt-4 pt-4 border-t border-gray-50 dark:border-gray-700 flex items-center text-[11px] font-semibold ${trendUp ? 'text-green-600' : 'text-gray-500'}`}>
            {trendUp ? <TrendingUp size={14} className="mr-1"/> : <TrendingDown size={14} className="mr-1"/>}
            {trend}
        </div>
    </div>
);

const TypeBar = ({label, count, total, color}: { label: string, count: number, total: number, color: string }) => {
    const percentage = total > 0 ? (count / total) * 100 : 0;
    return (
        <div className="space-y-1">
            <div className="flex justify-between text-sm">
                <span className="text-slate-700 dark:text-gray-300">{label}</span>
                <span className="font-medium text-slate-900 dark:text-white">{count}</span>
            </div>
            <div className="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div className={`h-full ${color} rounded-full transition-all duration-500`}
                     style={{width: `${percentage}%`}}/>
            </div>
        </div>
    );
};

const TrendingItem = ({title, views, index}: { title: string, views: string, index: number }) => (
    <div className="flex items-center gap-3 p-2 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">
        <span
            className="w-6 h-6 flex items-center justify-center bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs font-bold rounded shrink-0">
            {index}
        </span>
        <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-slate-800 dark:text-gray-200 truncate">{title}</p>
            <p className="text-xs text-slate-500">{views} views</p>
        </div>
        <TrendingUp size={16} className="text-green-500 shrink-0"/>
    </div>
);

export default Dashboard;
