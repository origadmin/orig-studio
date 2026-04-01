/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 标签页 - 展示所有标签及对应媒体数量
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Tag, Hash, Search} from 'lucide-react';

interface TagInfo {
    name: string;
    count: number;
}

const mockTags: TagInfo[] = [
    {name: 'Go', count: 128}, {name: 'React', count: 256}, {name: 'Docker', count: 89},
    {name: 'Kubernetes', count: 67}, {name: 'TypeScript', count: 198}, {name: 'Python', count: 312},
    {name: 'AWS', count: 145}, {name: 'Vue', count: 134}, {name: 'Redis', count: 56},
    {name: 'GraphQL', count: 78}, {name: 'MongoDB', count: 45}, {name: 'Linux', count: 203},
    {name: 'Node.js', count: 167}, {name: 'Rust', count: 34}, {name: 'C++', count: 92},
    {name: 'Java', count: 187}, {name: 'Swift', count: 41}, {name: 'Flutter', count: 63},
    {name: 'DevOps', count: 112}, {name: '机器学习', count: 89}, {name: '深度学习', count: 56},
    {name: '数据科学', count: 73}, {name: '区块链', count: 28}, {name: '网络安全', count: 47},
    {name: '算法', count: 156}, {name: '系统设计', count: 98}, {name: '前端', count: 221},
    {name: '后端', count: 178}, {name: '全栈', count: 134}, {name: '微服务', count: 76},
];

const TagsPage = () => {
    const [filter, setFilter] = useState('');
    const filteredTags = mockTags.filter(t =>
        t.name.toLowerCase().includes(filter.toLowerCase())
    );

    // 按数量降序排列，搜索时按名称排序
    const sortedTags = filter
        ? [...filteredTags].sort((a, b) => a.name.localeCompare(b.name))
        : [...filteredTags].sort((a, b) => b.count - a.count);

    return (
        <div className="space-y-6">
            {/* 标题 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Tag size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">标签</h1>
                </div>
                <span className="text-sm text-gray-500">{mockTags.length} 个标签</span>
            </div>

            {/* 搜索 */}
            <div className="relative max-w-md">
                <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"/>
                <input
                    type="text"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder="搜索标签..."
                    className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg pl-9 pr-4 py-2 text-sm focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none"
                />
            </div>

            {/* 标签网格 */}
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
                {sortedTags.map((tag) => (
                    <Link
                        key={tag.name}
                        to="/search"
                        search={{q: tag.name}}
                        className="group flex items-center gap-2 p-3 bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 rounded-xl hover:border-emerald-300 dark:hover:border-emerald-700 hover:shadow-md transition-all"
                    >
                        <Hash size={16} className="text-emerald-500 shrink-0"/>
                        <div className="min-w-0">
                            <p className="text-sm font-medium text-gray-800 dark:text-gray-200 truncate group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                {tag.name}
                            </p>
                            <p className="text-xs text-gray-400">{tag.count} 个视频</p>
                        </div>
                    </Link>
                ))}
            </div>

            {filteredTags.length === 0 && (
                <div className="text-center py-16 text-gray-400">
                    <Tag size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>没有找到匹配的标签</p>
                </div>
            )}
        </div>
    );
};

export default TagsPage;
