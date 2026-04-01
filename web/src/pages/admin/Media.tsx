/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 媒体管理页面
 */

import {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, MoreVertical, Trash2, Edit, Search, Filter, Upload, Image, Video} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
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

// 模拟媒体数据
const mockMedia = [
    {
        id: "1",
        title: "Building Go Microservices from Scratch",
        type: "video",
        thumbnail: "https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=200",
        size: 524288000,
        duration: 3600,
        view_count: 125400,
        status: "published",
        user: "Gopher Expert",
        created_at: "2024-03-15"
    },
    {
        id: "2",
        title: "React 18 New Features",
        type: "video",
        thumbnail: "https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=200",
        size: 312456789,
        duration: 2400,
        view_count: 89400,
        status: "published",
        user: "React Master",
        created_at: "2024-03-14"
    },
    {
        id: "3",
        title: "Docker Tutorial Banner",
        type: "image",
        thumbnail: "https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=200",
        size: 2048576,
        duration: 0,
        view_count: 12400,
        status: "published",
        user: "DevOps Pro",
        created_at: "2024-03-12"
    },
    {
        id: "4",
        title: "Kubernetes Deep Dive - Draft",
        type: "video",
        thumbnail: "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&q=80&w=200",
        size: 1024000000,
        duration: 5400,
        view_count: 0,
        status: "draft",
        user: "Cloud Expert",
        created_at: "2024-03-10"
    },
    {
        id: "5",
        title: "TypeScript Advanced Types",
        type: "video",
        thumbnail: "https://images.unsplash.com/photo-1516116216624-53e697fedbea?auto=format&fit=crop&q=80&w=200",
        size: 256000000,
        duration: 1800,
        view_count: 67800,
        status: "published",
        user: "TypeScript Guru",
        created_at: "2024-03-08"
    }
];

export default function MediaPage() {
    const [mediaList] = useState(mockMedia);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');

    const filteredMedia = mediaList.filter(item => {
        const matchesSearch = item.title.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || item.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const formatSize = (bytes: number) => {
        if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`;
        if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`;
        if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`;
        return `${bytes} B`;
    };

    const formatDuration = (seconds: number) => {
        if (!seconds) return '-';
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        if (hours > 0) {
            return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
        }
        return `${minutes}:${String(secs).padStart(2, '0')}`;
    };

    const formatViews = (count: number) => {
        if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
        if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
        return count.toString();
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">Media Management</h2>
                    <p className="text-slate-500 text-sm mt-1">Manage all videos and images in your platform</p>
                </div>
                <Button className="bg-blue-600 hover:bg-blue-700">
                    <Upload className="w-4 h-4 mr-2"/>
                    Upload Media
                </Button>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"/>
                    <Input
                        placeholder="Search media..."
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
                    <option value="private">Private</option>
                </select>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">Total Media</p>
                                <p className="text-2xl font-bold">{mediaList.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <Video className="w-6 h-6 text-blue-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">Videos</p>
                                <p className="text-2xl font-bold">{mediaList.filter(m => m.type === 'video').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <Play className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">Images</p>
                                <p className="text-2xl font-bold">{mediaList.filter(m => m.type === 'image').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Image className="w-6 h-6 text-green-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">Total Views</p>
                                <p className="text-2xl font-bold">{formatViews(mediaList.reduce((acc, m) => acc + m.view_count, 0))}</p>
                            </div>
                            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
                                <Eye className="w-6 h-6 text-orange-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Media Table */}
            <Card>
                <CardHeader>
                    <CardTitle>All Media</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Media</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Size</TableHead>
                                <TableHead>Duration</TableHead>
                                <TableHead>Views</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Author</TableHead>
                                <TableHead>Date</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredMedia.map((media) => (
                                <TableRow key={media.id}>
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <div className="w-16 h-10 bg-slate-100 rounded overflow-hidden shrink-0">
                                                <img src={media.thumbnail} alt=""
                                                     className="w-full h-full object-cover"/>
                                            </div>
                                            <span
                                                className="font-medium line-clamp-1 max-w-[200px]">{media.title}</span>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="outline">
                                            {media.type === 'video' ? <Video className="w-3 h-3 mr-1"/> :
                                                <Image className="w-3 h-3 mr-1"/>}
                                            {media.type}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-sm text-slate-500">{formatSize(media.size)}</TableCell>
                                    <TableCell
                                        className="text-sm text-slate-500">{formatDuration(media.duration)}</TableCell>
                                    <TableCell
                                        className="text-sm text-slate-500">{formatViews(media.view_count)}</TableCell>
                                    <TableCell>
                                        <Badge variant={media.status === 'published' ? 'default' : 'secondary'}>
                                            {media.status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-sm text-slate-500">{media.user}</TableCell>
                                    <TableCell className="text-sm text-slate-500">{media.created_at}</TableCell>
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