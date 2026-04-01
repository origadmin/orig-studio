/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 内容管理页面
 */

import {useState} from 'react';
import {Search, Plus, FileText, MoreVertical, Trash2, Edit, Eye} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

// 模拟内容数据
const mockContent = [
    {
        id: "1",
        title: "Getting Started with OrigCMS",
        type: "article",
        status: "published",
        author: "Admin",
        views: 12500,
        created_at: "2024-03-15"
    },
    {
        id: "2",
        title: "Media Upload Best Practices",
        type: "article",
        status: "published",
        author: "Editor",
        views: 8900,
        created_at: "2024-03-14"
    },
    {
        id: "3",
        title: "Community Guidelines",
        type: "page",
        status: "published",
        author: "Admin",
        views: 45000,
        created_at: "2024-03-10"
    },
    {
        id: "4",
        title: "New Features Coming Soon",
        type: "article",
        status: "draft",
        author: "Editor",
        views: 0,
        created_at: "2024-03-08"
    },
    {
        id: "5",
        title: "Privacy Policy",
        type: "page",
        status: "published",
        author: "Admin",
        views: 23000,
        created_at: "2024-03-01"
    }
];

export default function ContentPage() {
    const [contents] = useState(mockContent);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');

    const filteredContent = contents.filter(item => {
        const matchesSearch = item.title.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || item.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const formatViews = (count: number) => {
        if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
        return count.toString();
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">Content Management</h2>
                    <p className="text-slate-500 text-sm mt-1">Manage articles, pages, and static content</p>
                </div>
                <Button className="bg-blue-600 hover:bg-blue-700">
                    <Plus className="w-4 h-4 mr-2"/>
                    Create Content
                </Button>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"/>
                    <Input
                        placeholder="Search content..."
                        className="pl-10"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <select
                    className="h-10 px-3 rounded-md border border-input bg-background text-sm"
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                >
                    <option value="all">All Status</option>
                    <option value="published">Published</option>
                    <option value="draft">Draft</option>
                    <option value="archived">Archived</option>
                </select>
            </div>

            {/* Content Table */}
            <Card>
                <CardHeader>
                    <CardTitle>All Content</CardTitle>
                    <CardDescription>Manage your articles and pages</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Title</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Author</TableHead>
                                <TableHead>Views</TableHead>
                                <TableHead>Date</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredContent.map((item) => (
                                <TableRow key={item.id}>
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <div
                                                className="w-10 h-10 bg-slate-100 rounded-lg flex items-center justify-center">
                                                <FileText className="w-5 h-5 text-slate-500"/>
                                            </div>
                                            <span className="font-medium">{item.title}</span>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="outline">{item.type}</Badge>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={item.status === 'published' ? 'default' : 'secondary'}>
                                            {item.status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-sm text-slate-500">{item.author}</TableCell>
                                    <TableCell className="text-sm text-slate-500">{formatViews(item.views)}</TableCell>
                                    <TableCell className="text-sm text-slate-500">{item.created_at}</TableCell>
                                    <TableCell className="text-right">
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" size="sm">
                                                    <MoreVertical className="w-4 h-4"/>
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end">
                                                <DropdownMenuItem>
                                                    <Eye className="w-4 h-4 mr-2"/>
                                                    View
                                                </DropdownMenuItem>
                                                <DropdownMenuItem>
                                                    <Edit className="w-4 h-4 mr-2"/>
                                                    Edit
                                                </DropdownMenuItem>
                                                <DropdownMenuItem className="text-red-600">
                                                    <Trash2 className="w-4 h-4 mr-2"/>
                                                    Delete
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}