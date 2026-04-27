import React, {useState, useEffect} from 'react';
import {Info, Clock, Monitor, File, BarChart2, Text, Image, PlayCircle} from 'lucide-react';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Skeleton} from '@/components/ui/skeleton';
import {formatDate, formatDuration, formatFileSize} from '@/lib/format';
import {metadataApi, type MediaMetadata} from '@/lib/api/metadata';
import ErrorPage from '@/components/common/ErrorPage';

interface MediaMetadataProps {
    mediaId: string;
}

const MediaMetadata: React.FC<MediaMetadataProps> = ({mediaId}) => {
    const [metadata, setMetadata] = useState<MediaMetadata | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [miningStatus, setMiningStatus] = useState<{
        status: string;
        progress: number;
        message: string
    } | null>(null);
    const [isMining, setIsMining] = useState(false);

    useEffect(() => {
        fetchMetadata();
    }, [mediaId]);

    const fetchMetadata = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await metadataApi.getByMediaId(mediaId);
            setMetadata(response);
        } catch (err) {
            setError('Failed to fetch metadata');
            console.error('Failed to fetch metadata:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleTriggerMining = async () => {
        try {
            setIsMining(true);
            await metadataApi.triggerMining(mediaId);
            // 轮询挖掘状态
            const interval = setInterval(async () => {
                const status = await metadataApi.getMiningStatus(mediaId);
                setMiningStatus(status);
                if (status.status === 'completed' || status.status === 'failed') {
                    clearInterval(interval);
                    setIsMining(false);
                    // 重新获取元数据
                    await fetchMetadata();
                }
            }, 2000);
        } catch (err) {
            setError('Failed to trigger metadata mining');
            console.error('Failed to trigger metadata mining:', err);
            setIsMining(false);
        }
    };

    if (loading) {
        return (
            <div className="space-y-4">
                <Card>
                    <CardHeader>
                        <Skeleton className="h-8 w-1/3"/>
                    </CardHeader>
                    <CardContent>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {Array.from({length: 4}).map((_, i) => (
                                <div key={i} className="space-y-2">
                                    <Skeleton className="h-4 w-1/4"/>
                                    <Skeleton className="h-4 w-3/4"/>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    if (!metadata) {
        return (
            <div className="space-y-4">
                <Card>
                    <CardHeader>
                        <CardTitle>Media Metadata</CardTitle>
                        <CardDescription>No metadata available</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button onClick={handleTriggerMining} disabled={isMining}>
                            {isMining ? 'Mining...' : 'Trigger Metadata Mining'}
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Basic Metadata */}
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <CardTitle className="flex items-center gap-2">
                            <Info className="w-5 h-5"/>
                            Basic Metadata
                        </CardTitle>
                        <Button onClick={handleTriggerMining} disabled={isMining}>
                            {isMining ? 'Mining...' : 'Update Metadata'}
                        </Button>
                    </div>
                    <CardDescription>Basic information about the media file</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Title</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.title}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Description</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.description}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Duration</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{formatDuration(metadata.duration)}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Resolution</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.resolution} ({metadata.width}x{metadata.height})</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Format</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.format}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Codec</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.codec}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Bitrate</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.bitrate} kbps</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">File Size</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{formatFileSize(metadata.file_size)}</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Frame Rate</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.frame_rate} fps</p>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground">Aspect Ratio</h4>
                            <p className="text-sm text-gray-900 dark:text-white">{metadata.aspect_ratio}</p>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Key Frames */}
            {metadata.key_frames && metadata.key_frames.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Image className="w-5 h-5"/>
                            Key Frames
                        </CardTitle>
                        <CardDescription>Key frames extracted from the video</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                            {metadata.key_frames.map((frame) => (
                                <div key={frame.id} className="space-y-2">
                                    <div className="aspect-video rounded-lg overflow-hidden">
                                        <img src={frame.thumbnail_url}
                                             alt={`Key frame at ${formatDuration(frame.time)}`}
                                             className="w-full h-full object-cover"/>
                                    </div>
                                    <p className="text-xs text-gray-500 dark:text-muted-foreground">{formatDuration(frame.time)}</p>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Text Content */}
            {metadata.text_content && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Text className="w-5 h-5"/>
                            Text Content
                        </CardTitle>
                        <CardDescription>Transcript and keywords extracted from the video</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {metadata.text_content.transcript && (
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground mb-2">Transcript</h4>
                                    <p className="text-sm text-gray-900 dark:text-white whitespace-pre-wrap">{metadata.text_content.transcript}</p>
                                </div>
                            )}
                            {metadata.text_content.keywords && metadata.text_content.keywords.length > 0 && (
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground mb-2">Keywords</h4>
                                    <div className="flex flex-wrap gap-2">
                                        {metadata.text_content.keywords.map((keyword, index) => (
                                            <span key={index}
                                                  className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded-full text-xs text-gray-700 dark:text-gray-300">
                        {keyword}
                      </span>
                                        ))}
                                    </div>
                                </div>
                            )}
                            {metadata.text_content.entities && metadata.text_content.entities.length > 0 && (
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 dark:text-muted-foreground mb-2">Entities</h4>
                                    <div className="space-y-2">
                                        {metadata.text_content.entities.map((entity, index) => (
                                            <div key={index} className="flex items-center gap-2">
                        <span
                            className="px-2 py-1 bg-blue-100 dark:bg-blue-900 rounded text-xs text-blue-700 dark:text-blue-300">
                          {entity.type}
                        </span>
                                                <span
                                                    className="text-sm text-gray-900 dark:text-white">{entity.text}</span>
                                                <span className="text-xs text-gray-500 dark:text-muted-foreground">
                          {Math.round(entity.confidence * 100)}%
                        </span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Scene Changes */}
            {metadata.scene_changes && metadata.scene_changes.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <PlayCircle className="w-5 h-5"/>
                            Scene Changes
                        </CardTitle>
                        <CardDescription>Scene changes detected in the video</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {metadata.scene_changes.map((scene) => (
                                <div key={scene.id} className="border-l-4 border-blue-500 pl-4 py-2">
                                    <div className="flex items-center gap-2">
                                        <span
                                            className="text-sm font-medium text-gray-900 dark:text-white">{formatDuration(scene.time)}</span>
                                    </div>
                                    <p className="text-sm text-gray-600 dark:text-muted-foreground mt-1">{scene.description}</p>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
};

export default MediaMetadata;
