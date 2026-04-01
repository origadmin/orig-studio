import React from 'react';
import {Film, Users, Eye, Heart, BarChart3, TrendingUp, TrendingDown} from 'lucide-react';

const Dashboard = () => {
    return (
        <div className="space-y-8">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900">Platform Overview</h1>
                    <p className="text-slate-500 text-sm mt-1">Real-time statistics for your media platform.</p>
                </div>
                <div className="flex gap-2">
                    <button
                        className="px-4 py-2 bg-white border border-gray-200 rounded-lg text-sm font-semibold hover:bg-gray-50 shadow-sm transition-colors">Export
                        Report
                    </button>
                    <button
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-semibold hover:bg-blue-700 shadow-sm transition-colors">Manage
                        Content
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard icon={<Film className="text-blue-500"/>} label="Total Medias" value="1,248"
                          trend="+12% from last week"/>
                <StatCard icon={<Users className="text-green-500"/>} label="Total Users" value="48.2k"
                          trend="+5.4% from last month"/>
                <StatCard icon={<Eye className="text-purple-500"/>} label="Total Views" value="2.4M"
                          trend="+18.2% from last month"/>
                <StatCard icon={<Heart className="text-rose-500"/>} label="Likes Given" value="184k"
                          trend="+3.1% from last month"/>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 bg-white p-6 rounded-2xl border border-gray-100 shadow-sm">
                    <div className="flex items-center justify-between mb-6">
                        <h3 className="font-bold text-slate-900 flex items-center gap-2">
                            <BarChart3 size={20} className="text-slate-400"/>
                            Growth Analytics
                        </h3>
                        <select className="text-sm border-none bg-gray-50 rounded-lg p-1.5 focus:ring-0 outline-none">
                            <option>Last 30 days</option>
                            <option>Last 6 months</option>
                            <option>Last 12 months</option>
                        </select>
                    </div>
                    <div
                        className="h-64 bg-slate-50 rounded-xl flex items-center justify-center text-slate-400 border border-dashed border-slate-200">
                        [ Chart Placeholder - Content Growth ]
                    </div>
                </div>

                <div className="bg-white p-6 rounded-2xl border border-gray-100 shadow-sm">
                    <h3 className="font-bold text-slate-900 mb-6">Trending Content</h3>
                    <div className="space-y-4">
                        <TrendingItem title="Top 10 AI Tools 2024" views="84k" trend="up"/>
                        <TrendingItem title="React Framework Comparison" views="62k" trend="up"/>
                        <TrendingItem title="Docker Tutorial for Beginners" views="48k" trend="down"/>
                        <TrendingItem title="Mastering Go Concurrency" views="36k" trend="up"/>
                        <TrendingItem title="Tailwind CSS Best Practices" views="22k" trend="up"/>
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
