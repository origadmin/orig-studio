import React from 'react';
import {useTranslation} from 'react-i18next';
import {useMediaList, useDeleteMedia} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {Card, CardContent} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {
    Video,
    Clock,
    Eye,
    MoreVertical,
    Trash2,
    Edit,
    Plus,
    ExternalLink
} from 'lucide-react';
import {Link} from '@tanstack/react-router';
import {formatRelativeTime, formatDuration} from '@/lib/format';

const MyVideos = () => {
    const {t} = useTranslation();
    const {user} = useAuth();
    const [page, setPage] = React.useState(1);
    const pageSize = 12;

    const {data, isLoading} = useMediaList({
        page,
        page_size: pageSize,
        user_id: user?.id
    });

    const deleteMutation = useDeleteMedia();

    const handleDelete = async (id: number) => {
        if (window.confirm('确定要删除这个视频吗？')) {
            await deleteMutation.mutateAsync(id.toString());
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    const mediaList = data?.list || [];

    return (
        <div className="space-y-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">我的上传</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400">管理你上传的所有媒体内容</p>
                </div>
                <Button asChild className="bg-emerald-600 hover:bg-emerald-700 text-white">
                    <Link to="/me/upload">
                        <Plus className="w-4 h-4 mr-2"/>
                        上传新内容
                    </Link>
                </Button>
            </div>

            {mediaList.length === 0 ? (
                <Card className="border-dashed border-2">
                    <CardContent className="flex flex-col items-center justify-center py-20 space-y-4">
                        <div
                            className="w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center">
                            <Video className="w-8 h-8 text-gray-400"/>
                        </div>
                        <div className="text-center">
                            <h3 className="text-lg font-medium text-gray-900 dark:text-white">暂无上传内容</h3>
                            <p className="text-sm text-gray-500">你还没有上传过任何视频或图片</p>
                        </div>
                        <Button asChild variant="outline">
                            <Link to="/me/upload">立即上传</Link>
                        </Button>
                    </CardContent>
                </Card>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {mediaList.map((item) => (
                        <Card key={item.id} className="overflow-hidden group hover:shadow-lg transition-shadow">
                            <div className="relative aspect-video bg-gray-100 dark:bg-gray-800">
                                {item.thumbnail ? (
                                    <img
                                        src={item.thumbnail}
                                        alt={item.title}
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <div className="w-full h-full flex items-center justify-center">
                                        <Video className="w-10 h-10 text-gray-300"/>
                                    </div>
                                )}
                                <div
                                    className="absolute bottom-2 right-2 px-1.5 py-0.5 bg-black/60 text-white text-[10px] rounded font-medium">
                                    {formatDuration(item.duration)}
                                </div>
                                <div
                                    className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                                    <Button size="sm" variant="secondary" className="h-8 w-8 p-0 rounded-full" asChild>
                                        <Link to="/watch" search={{v: item.id.toString()}}>
                                            <ExternalLink className="w-4 h-4"/>
                                        </Link>
                                    </Button>
                                </div>
                            </div>
                            <CardContent className="p-4">
                                <div className="flex justify-between items-start gap-2">
                                    <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-1 flex-1">
                                        {item.title}
                                    </h3>
                                    <Badge variant={item.state === 'active' ? 'default' : 'secondary'}
                                           className="text-[10px] px-1.5 py-0 capitalize">
                                        {item.state}
                                    </Badge>
                                </div>

                                <div className="mt-3 flex items-center gap-4 text-xs text-gray-500">
                                    <div className="flex items-center gap-1">
                                        <Eye className="w-3 h-3"/>
                                        {item.view_count}
                                    </div>
                                    <div className="flex items-center gap-1">
                                        <Clock className="w-3 h-3"/>
                                        {formatRelativeTime(item.created_at)}
                                    </div>
                                </div>

                                <div className="mt-4 pt-4 border-t flex justify-end gap-2">
                                    <Button variant="ghost" size="sm"
                                            className="h-8 text-gray-500 hover:text-emerald-600">
                                        <Edit className="w-3.5 h-3.5 mr-1"/>
                                        编辑
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-8 text-gray-500 hover:text-red-600"
                                        onClick={() => handleDelete(item.id)}
                                    >
                                        <Trash2 className="w-3.5 h-3.5 mr-1"/>
                                        删除
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            {data && data.total > pageSize && (
                <div className="flex justify-center pt-8">
                    <div className="flex gap-2">
                        {Array.from({length: Math.ceil(data.total / pageSize)}).map((_, i) => (
                            <Button
                                key={i}
                                variant={page === i + 1 ? 'default' : 'outline'}
                                size="sm"
                                onClick={() => setPage(i + 1)}
                            >
                                {i + 1}
                            </Button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default MyVideos;
