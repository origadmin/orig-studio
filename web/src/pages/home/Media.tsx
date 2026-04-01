// 用户端 - 媒体浏览页面
import {useState, useEffect} from "react";
import {Link} from "@tanstack/react-router";
import {mediaApi, categoryApi, type Media, type Category} from "@/lib/api";
import {useTranslation} from 'react-i18next';

export default function MediaPage() {
    const {t} = useTranslation();
    const [mediaList, setMediaList] = useState<Media[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [activeCategory, setActiveCategory] = useState<string>("all");

    useEffect(() => {
        loadData();
    }, [activeCategory]);

    const loadData = async () => {
        setLoading(true);
        setError("");
        try {
            const params = activeCategory === "all" ? {} : {category_id: activeCategory};
            const [catRes] = await Promise.all([
                categoryApi.getAll()
            ]);
            const mediaRes = await mediaApi.list(params);
            setMediaList(mediaRes.list);
            setCategories(catRes);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const formatSize = (bytes: number) => {
        if (bytes > 1024 * 1024 * 1024) return (bytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
        if (bytes > 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + " MB";
        if (bytes > 1024) return (bytes / 1024).toFixed(1) + " KB";
        return bytes + " B";
    };

    const formatDuration = (seconds: number) => {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${s.toString().padStart(2, "0")}`;
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-50">
                <div className="max-w-7xl mx-auto px-4 py-8">
                    <div className="animate-pulse space-y-4">
                        <div className="h-8 bg-gray-200 rounded w-48"></div>
                        <div className="grid grid-cols-4 gap-6">
                            {[1, 2, 3, 4].map(i => (
                                <div key={i} className="bg-white rounded-lg overflow-hidden">
                                    <div className="h-48 bg-gray-200"></div>
                                    <div className="p-4 space-y-2">
                                        <div className="h-4 bg-gray-200 rounded"></div>
                                        <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-red-500 mb-2">{t('media.loadFailed', {error})}</p>
                    <button onClick={loadData} className="text-indigo-600 hover:underline">
                        {t('media.retry')}
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="max-w-7xl mx-auto px-4 py-8">
                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-2xl font-bold text-gray-900">{t('media.title')}</h1>
                    <Link to="/" className="text-indigo-600 hover:underline">
                        {t('media.back')}
                    </Link>
                </div>

                {/* 分类筛选 */}
                <div className="flex gap-2 mb-6 overflow-x-auto pb-2">
                    <button
                        onClick={() => setActiveCategory("all")}
                        className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap transition ${
                            activeCategory === "all"
                                ? "bg-indigo-600 text-white"
                                : "bg-white text-gray-700 hover:bg-gray-100"
                        }`}
                    >
                        {t('media.all')}
                    </button>
                    {categories.map((cat) => (
                        <button
                            key={cat.id}
                            onClick={() => setActiveCategory(cat.id)}
                            className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap transition ${
                                activeCategory === cat.id
                                    ? "bg-indigo-600 text-white"
                                    : "bg-white text-gray-700 hover:bg-gray-100"
                            }`}
                        >
                            {cat.name}
                        </button>
                    ))}
                </div>

                {/* 媒体列表 */}
                {mediaList.length === 0 ? (
                    <div className="text-center py-12 text-gray-500">
                        {t('media.noMedia')}
                    </div>
                ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                        {mediaList.map((media) => (
                            <Link
                                key={media.id}
                                to="/v/$id"
                                params={{id: media.id}}
                                className="bg-white rounded-lg overflow-hidden shadow-sm hover:shadow-md transition group"
                            >
                                <div className="relative aspect-video bg-gray-100">
                                    {media.thumbnail ? (
                                        <img
                                            src={media.thumbnail}
                                            alt={media.title}
                                            className="w-full h-full object-cover"
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center text-4xl">
                                            {media.type === "video" ? "🎬" : media.type === "audio" ? "🎵" : "🖼️"}
                                        </div>
                                    )}
                                    {media.duration > 0 && (
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/70 text-white text-xs px-2 py-1 rounded">
                                            {formatDuration(media.duration)}
                                        </div>
                                    )}
                                </div>
                                <div className="p-4">
                                    <h3 className="font-semibold text-gray-900 group-hover:text-indigo-600 transition line-clamp-2">
                                        {media.title}
                                    </h3>
                                    <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                                        {media.description}
                                    </p>
                                    <div className="flex items-center justify-between mt-3 text-xs text-gray-400">
                                        <span>👁 {media.views}</span>
                                        <span>❤️ {media.likes}</span>
                                        <span>{formatSize(media.size)}</span>
                                    </div>
                                </div>
                            </Link>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
