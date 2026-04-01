/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Featured Page (TODO: i18n)
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Star} from 'lucide-react';
import {formatDuration, formatViews} from '@/lib/format';

// TODO: replace with API call
const mockFeatured = [
    {
        id: 101, title: 'Kubernetes 生产环境最佳实践',
        description: '从零搭建生产级 K8s 集群，涵盖监控、日志、自动扩缩容等核心话题。',
        thumbnail: 'https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=800',
        duration: 5400, view_count: 234500, author_name: 'DevOps Pro',
        author_avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&q=80&w=100',
    },
    {
        id: 102, title: '系统设计面试完全指南',
        description: '全面覆盖 Top 公司系统设计面试的套路和答题框架。',
        thumbnail: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&q=80&w=800',
        duration: 7200, view_count: 456000, author_name: 'Interview Coach',
        author_avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&q=80&w=100',
    },
    {
        id: 103, title: 'Go 语言并发编程深入',
        description: '掌握 goroutine、channel、context 的底层原理和高级用法。',
        thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=800',
        duration: 4800, view_count: 178000, author_name: 'Gopher Expert',
        author_avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=100',
    },
    {
        id: 104, title: 'React Server Components 实战',
        description: '深入理解 RSC 架构，从理论到生产实践。',
        thumbnail: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=800',
        duration: 3600, view_count: 142000, author_name: 'React Master',
        author_avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=100',
    },
];

const FeaturedPage = () => (
    <div className="space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
                <Star size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{/* TODO: i18n */}精选内容</h1>
            </div>
            <span className="text-sm text-gray-500 dark:text-gray-400">
                {/* TODO: i18n */}{mockFeatured.length} 部精选视频
            </span>
        </div>

        {/* Hero cards */}
        {mockFeatured.slice(0, 2).map((item) => (
            <Link key={item.id} to="/v/$id" params={{id: String(item.id)}} className="group block">
                <div
                    className="relative rounded-2xl overflow-hidden bg-gradient-to-r from-gray-900 to-gray-800 dark:from-gray-800 dark:to-gray-900">
                    <div className="flex flex-col md:flex-row">
                        <div className="relative aspect-video md:w-[480px] shrink-0">
                            <img src={item.thumbnail} alt={item.title} className="w-full h-full object-cover"/>
                            <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent"/>
                            <div className="absolute bottom-3 right-3 bg-black/80 text-white text-xs px-2 py-1 rounded">
                                {formatDuration(item.duration)}
                            </div>
                            <div
                                className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                <div
                                    className="w-16 h-16 bg-white/90 rounded-full flex items-center justify-center shadow-xl">
                                    <Play className="w-7 h-7 text-gray-900 ml-1" fill="currentColor"/>
                                </div>
                            </div>
                        </div>
                        <div className="p-6 md:p-8 flex flex-col justify-center">
                            <span className="inline-flex items-center gap-1 text-emerald-400 text-xs font-medium mb-3">
                                <Star size={12} fill="currentColor"/>{' '}
                                {/* TODO: i18n */}编辑精选
                            </span>
                            <h2 className="text-2xl font-bold text-white mb-3 group-hover:text-emerald-300 transition-colors">
                                {item.title}
                            </h2>
                            <p className="text-gray-300 text-sm mb-4 line-clamp-2">{item.description}</p>
                            <div className="flex items-center gap-3">
                                <img src={item.author_avatar} alt={item.author_name} className="w-8 h-8 rounded-full"/>
                                <span className="text-gray-400 text-sm">{item.author_name}</span>
                                <span className="text-gray-500 text-sm flex items-center gap-1">
                                    <Eye size={14}/>{formatViews(item.view_count)} {/* TODO: i18n */}次观看
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </Link>
        ))}

        {/* Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {mockFeatured.slice(2).map((item) => (
                <Link key={item.id} to="/v/$id" params={{id: String(item.id)}} className="group">
                    <div
                        className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                        <div className="relative aspect-video overflow-hidden">
                            <img src={item.thumbnail} alt={item.title}
                                 className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                            <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">
                                {formatDuration(item.duration)}
                            </div>
                        </div>
                        <div className="p-4">
                            <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-2 mb-2 group-hover:text-emerald-600 transition-colors">
                                {item.title}
                            </h3>
                            <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{item.description}</p>
                        </div>
                    </div>
                </Link>
            ))}
        </div>
    </div>
);

export default FeaturedPage;
