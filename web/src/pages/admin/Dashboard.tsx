import React from 'react';
import {Film, Users, Eye, Heart, BarChart3, TrendingUp, TrendingDown, Loader2, MessageCircle} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useQuery} from '@tanstack/react-query';
import {statsApi} from '@/lib/api/stats';

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
                <p className="text-sm">{error.message}</p>
            </div>
        );
    }

    const stats = data || {
        total_media: 0,
        total_users: 0,
        total_views: 0,
        total_comments: 0,
        top_media: []
    };

    const formatNumber = (num: number) => {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'k';
        }
        return num.toString();
    };

    return (
        <div className="space-y-8">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900">{t('admin.dashboard')}</h1>
                    <p className="text-slate-500 text-sm mt-1">{t('admin.dashboardDesc')}</p>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        className="px-4 py-2 bg-white border border-gray-200 rounded-lg text-sm font-semibold hover:bg-gray-50 shadow-sm transition-colors">
                        {t('admin.exportReport') || 'Export Report'}
                    </Button>
                    <Button
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-semibold hover:bg-blue-700 shadow-sm transition-colors">
                        {t('admin.manageContent') || 'Manage Content'}
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard icon={<Film className="text-blue-500"/>} label={t('admin.totalMedia') || 'Total Medias'}
                          value={formatNumber(stats.total_media)}
                          trend="+12% from last week"/>
                <StatCard icon={<Users className="text-green-500"/>} label={t('admin.totalUsers') || 'Total Users'}
                          value={formatNumber(stats.total_users)}
                          trend="+5.4% from last month"/>
                <StatCard icon={<Eye className="text-purple-500"/>} label={t('admin.totalViews') || 'Total Views'}
                          value={formatNumber(stats.total_views)}
                          trend="+18.2% from last month"/>
                <StatCard icon={<MessageCircle className="text-rose-500"/>}
                          label={t('admin.totalComments') || 'Total Comments'}
                          value={formatNumber(stats.total_comments)}
                          trend="+3.1% from last month"/>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 bg-white p-6 rounded-2xl border border-gray-100 shadow-sm">
                    <div className="flex items-center justify-between mb-6">
                        <h3 className="font-bold text-slate-900 flex items-center gap-2">
                            <BarChart3 size={20} className="text-slate-400"/>
                            {t('admin.growthAnalytics') || 'Growth Analytics'}
                        </h3>
                        <select className="text-sm border-none bg-gray-50 rounded-lg p-1.5 focus:ring-0 outline-none">
                            <option>{t('admin.last30Days') || 'Last 30 days'}</option>
                            <option>{t('admin.last6Months') || 'Last 6 months'}</option>
                            <option>{t('admin.last12Months') || 'Last 12 months'}</option>
                        </select>
                    </div>
                    <div
                        className="h-64 bg-slate-50 rounded-xl flex items-center justify-center text-slate-400 border border-dashed border-slate-200">
                        [ Chart Placeholder - Content Growth ]
                    </div>
                </div>

                <div className="bg-white p-6 rounded-2xl border border-gray-100 shadow-sm">
                    <h3 className="font-bold text-slate-900 mb-6">{t('admin.trendingContent') || 'Trending Content'}</h3>
                    <div className="space-y-4">
                        {stats.top_media.map((item: any, index: number) => (
                            <TrendingItem key={index} title={item.title} views={formatNumber(item.views)}
                                          trend="up"/>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

const StatCard = ({icon, label, value, trend}: any) => (
    <div className="bg-white p-6 rounded-2xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
        <div className="flex items-center space-x-4">
            <div className="p-3 bg-gray-50 rounded-xl">{icon}</div>
            <div className="flex-1">
                <p className="text-sm font-medium text-slate-500">{label}</p>
                <h3 className="text-2xl font-extrabold text-slate-900 mt-0.5">{value}</h3>
            </div>
        </div>
        <div className="mt-4 pt-4 border-t border-gray-50 flex items-center text-[11px] font-semibold text-green-600">
            <TrendingUp size={14} className="mr-1"/>
            {trend}
        </div>
    </div>
);

const TrendingItem = ({title, views, trend}: any) => (
    <div
        className="flex items-center justify-between group cursor-pointer hover:bg-gray-50 p-2 rounded-lg transition-colors">
        <div className="min-w-0">
            <p className="text-sm font-bold text-slate-800 truncate group-hover:text-blue-600">{title}</p>
            <p className="text-xs text-slate-500 font-medium">{views} views</p>
        </div>
        {trend === 'up' ? <TrendingUp size={16} className="text-green-500 shrink-0"/> :
            <TrendingDown size={16} className="text-rose-500 shrink-0"/>}
    </div>
);

export default Dashboard;
