import React, {useState, useEffect} from 'react';
import {
    LineChart,
    Line,
    BarChart,
    Bar,
    PieChart,
    Pie,
    Cell,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer
} from 'recharts';
import {Users, Video, Eye, MessageSquare, Users2, TrendingUp, Calendar, Award, Clock, Activity} from 'lucide-react';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Skeleton} from '@/components/ui/skeleton';
import {useTranslation} from 'react-i18next';
import {statsApi, type DashboardStats} from '@/lib/api/stats';
import ErrorPage from '@/components/common/ErrorPage';
import {formatNumber} from '@/lib/format';

const Dashboard: React.FC = () => {
    const {t} = useTranslation();
    const [stats, setStats] = useState<DashboardStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetchDashboardStats();
    }, []);

    const fetchDashboardStats = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await statsApi.getDashboard();
            setStats(response);
        } catch (err) {
            setError('Failed to fetch dashboard statistics');
            console.error('Failed to fetch dashboard statistics:', err);
        } finally {
            setLoading(false);
        }
    };

    const getColor = (index: number) => {
        const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];
        return colors[index % colors.length];
    };

    if (loading) {
        return (
            <div className="space-y-6">
                {/* Overview Cards */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    {Array.from({length: 4}).map((_, i) => (
                        <Card key={i}>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                                    <Skeleton className="h-4 w-24"/>
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="flex items-center justify-between">
                                    <Skeleton className="h-8 w-32"/>
                                    <Skeleton className="h-10 w-10 rounded-full"/>
                                </div>
                                <Skeleton className="h-4 w-16 mt-2"/>
                            </CardContent>
                        </Card>
                    ))}
                </div>

                {/* Charts */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>
                                <Skeleton className="h-6 w-48"/>
                            </CardTitle>
                            <CardDescription>
                                <Skeleton className="h-4 w-64"/>
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Skeleton className="h-80 w-full"/>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>
                                <Skeleton className="h-6 w-48"/>
                            </CardTitle>
                            <CardDescription>
                                <Skeleton className="h-4 w-64"/>
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Skeleton className="h-80 w-full"/>
                        </CardContent>
                    </Card>
                </div>

                {/* Top Lists */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {Array.from({length: 3}).map((_, i) => (
                        <Card key={i}>
                            <CardHeader>
                                <CardTitle>
                                    <Skeleton className="h-6 w-48"/>
                                </CardTitle>
                                <CardDescription>
                                    <Skeleton className="h-4 w-64"/>
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-4">
                                    {Array.from({length: 5}).map((_, j) => (
                                        <div key={j} className="flex items-center justify-between">
                                            <Skeleton className="h-4 w-32"/>
                                            <Skeleton className="h-4 w-16"/>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    if (!stats) {
        return <ErrorPage message="No statistics available"/>;
    }

    return (
        <div className="space-y-6">
            {/* Overview Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                            {t('admin.totalUsers')}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-3xl font-bold text-gray-900 dark:text-white">
                                    {formatNumber(stats.total_users)}
                                </p>
                                <p className="text-sm text-green-600 dark:text-green-400">
                                    +{stats.new_users_today} {t('admin.today')}
                                </p>
                            </div>
                            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-full">
                                <Users className="h-6 w-6 text-blue-600 dark:text-blue-300"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                            {t('admin.totalMedia')}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-3xl font-bold text-gray-900 dark:text-white">
                                    {formatNumber(stats.total_media)}
                                </p>
                                <p className="text-sm text-green-600 dark:text-green-400">
                                    +{stats.new_media_today} {t('admin.today')}
                                </p>
                            </div>
                            <div className="p-2 bg-green-100 dark:bg-green-900 rounded-full">
                                <Video className="h-6 w-6 text-green-600 dark:text-green-300"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                            {t('admin.totalViews')}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-3xl font-bold text-gray-900 dark:text-white">
                                    {formatNumber(stats.total_views)}
                                </p>
                                <p className="text-sm text-green-600 dark:text-green-400">
                                    +{stats.new_views_today} {t('admin.today')}
                                </p>
                            </div>
                            <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-full">
                                <Eye className="h-6 w-6 text-amber-600 dark:text-amber-300"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                            {t('admin.totalSubscribers')}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-3xl font-bold text-gray-900 dark:text-white">
                                    {formatNumber(stats.total_subscribers)}
                                </p>
                                <p className="text-sm text-green-600 dark:text-green-400">
                                    +{stats.new_subscribers_today} {t('admin.today')}
                                </p>
                            </div>
                            <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-full">
                                <Users2 className="h-6 w-6 text-purple-600 dark:text-purple-300"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle>{t('admin.viewsByDate')}</CardTitle>
                        <CardDescription>
                            {t('admin.last30DaysViews')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="h-80">
                            <ResponsiveContainer width="100%" height="100%">
                                <LineChart data={stats.views_by_date}>
                                    <CartesianGrid strokeDasharray="3 3"/>
                                    <XAxis dataKey="date"/>
                                    <YAxis/>
                                    <Tooltip/>
                                    <Legend/>
                                    <Line type="monotone" dataKey="views" stroke="#3b82f6" strokeWidth={2}
                                          activeDot={{r: 8}}/>
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>{t('admin.mediaByType')}</CardTitle>
                        <CardDescription>
                            {t('admin.mediaTypeDistribution')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="h-80">
                            <ResponsiveContainer width="100%" height="100%">
                                <PieChart>
                                    <Pie
                                        data={[
                                            {name: 'Video', value: stats.media_by_type.video},
                                            {name: 'Image', value: stats.media_by_type.image},
                                            {name: 'Audio', value: stats.media_by_type.audio},
                                            {name: 'Other', value: stats.media_by_type.other},
                                        ]}
                                        cx="50%"
                                        cy="50%"
                                        labelLine={false}
                                        outerRadius={100}
                                        fill="#8884d8"
                                        dataKey="value"
                                        label={({name, percent}) => `${name}: ${(percent * 100).toFixed(0)}%`}
                                    >
                                        {[
                                            {name: 'Video', value: stats.media_by_type.video},
                                            {name: 'Image', value: stats.media_by_type.image},
                                            {name: 'Audio', value: stats.media_by_type.audio},
                                            {name: 'Other', value: stats.media_by_type.other},
                                        ].map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={getColor(index)}/>
                                        ))}
                                    </Pie>
                                    <Tooltip/>
                                    <Legend/>
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Top Lists */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle>{t('admin.topCategories')}</CardTitle>
                        <CardDescription>
                            {t('admin.mostPopularCategories')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {stats.top_categories.slice(0, 5).map((category, index) => (
                                <div key={category.id} className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        <div
                                            className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-800">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {index + 1}
                      </span>
                                        </div>
                                        <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {category.name}
                    </span>
                                    </div>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">
                    {category.count} {t('admin.media')}
                  </span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>{t('admin.topCreators')}</CardTitle>
                        <CardDescription>
                            {t('admin.mostActiveCreators')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {stats.top_creators.slice(0, 5).map((creator, index) => (
                                <div key={creator.id} className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        <div
                                            className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-800">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {index + 1}
                      </span>
                                        </div>
                                        <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {creator.name}
                    </span>
                                    </div>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">
                    {creator.media_count} {t('admin.videos')}
                  </span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>{t('admin.topMedia')}</CardTitle>
                        <CardDescription>
                            {t('admin.mostViewedContent')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {stats.top_media.slice(0, 5).map((media, index) => (
                                <div key={media.id} className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        <div
                                            className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-800">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {index + 1}
                      </span>
                                        </div>
                                        <span
                                            className="text-sm font-medium text-gray-900 dark:text-white truncate max-w-[150px]">
                      {media.title}
                    </span>
                                    </div>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">
                    {formatNumber(media.views)} {t('admin.views')}
                  </span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};

export default Dashboard;
