/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState} from "react";
import {mediaApi, type TranscodingStatus, type TranscodingMediaItem} from "../../lib/api/media";
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from "../../components/ui/card";
import {Badge} from "../../components/ui/badge";
import {Activity, Clock, AlertTriangle, CheckCircle, Film, Settings, Download} from "lucide-react";
import {Button} from "../../components/ui/button";
import {ScrollArea} from "../../components/ui/scroll-area";
import {Progress} from "../../components/ui/progress";

export default function TranscodingStatus() {
    const [status, setStatus] = useState<TranscodingStatus | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchStatus = async () => {
        try {
            // In our request.ts, api.get returns response.data directly
            const response = await mediaApi.getTranscodingStatus();
            setStatus(response);
        } catch (error) {
            console.error("Failed to fetch status:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, 5000);
        return () => clearInterval(interval);
    }, []);

    if (loading) return <div className="p-6 text-center">Loading status...</div>;

    const cards = [
        {
            title: "Processing",
            count: status?.processing_count || 0,
            icon: Activity,
            color: "text-blue-500",
            bgColor: "bg-blue-50"
        },
        {
            title: "Pending",
            count: status?.pending_count || 0,
            icon: Clock,
            color: "text-yellow-500",
            bgColor: "bg-yellow-50"
        },
        {
            title: "Success",
            count: status?.success_count || 0,
            icon: CheckCircle,
            color: "text-green-500",
            bgColor: "bg-green-50"
        },
        {
            title: "Failed",
            count: status?.failed_count || 0,
            icon: AlertTriangle,
            color: "text-red-500",
            bgColor: "bg-red-50"
        },
    ];

    const getStatusBadge = (taskStatus: string) => {
        switch (taskStatus) {
            case "processing":
                return <Badge variant="default" className="bg-blue-500">Processing</Badge>;
            case "pending":
                return <Badge variant="secondary" className="bg-yellow-500">Pending</Badge>;
            case "success":
                return <Badge variant="default" className="bg-green-500">Success</Badge>;
            case "failed":
                return <Badge variant="destructive">Failed</Badge>;
            default:
                return <Badge variant="outline">Unknown</Badge>;
        }
    };

    const renderTaskDetails = (item: TranscodingMediaItem) => {
        return (
            <Card key={item.media.id} className="mb-4">
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <Film className="h-5 w-5 text-primary"/>
                            <div>
                                <CardTitle className="text-lg">{item.media.title}</CardTitle>
                                <CardDescription className="flex items-center gap-2">
                                    <span>ID: {item.media.id}</span>
                                    <span>•</span>
                                    <span>Type: {item.media.type}</span>
                                    {item.media.encoding_status && (
                                        <>
                                            <span>•</span>
                                            <Badge variant="outline">{item.media.encoding_status}</Badge>
                                        </>
                                    )}
                                </CardDescription>
                            </div>
                        </div>
                        {item.media.hls_file && (
                            <Button variant="outline" size="sm"
                                    onClick={() => window.open(item.media.hls_file, '_blank')}>
                                <Download className="h-4 w-4 mr-2"/>
                                View HLS
                            </Button>
                        )}
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="space-y-3">
                        {item.tasks.map((task) => (
                            <div key={task.id} className="border rounded-lg p-3 bg-muted/20">
                                <div className="flex items-center justify-between mb-2">
                                    <div className="flex items-center gap-2">
                                        <Settings className="h-4 w-4 text-muted-foreground"/>
                                        <span className="font-medium">Profile #{task.profile_id}</span>
                                        {getStatusBadge(task.status)}
                                    </div>
                                    <span className="text-sm text-muted-foreground">
                                        Progress: {task.progress}%
                                    </span>
                                </div>
                                <Progress value={task.progress} className="h-2"/>
                                {task.output_path && (
                                    <div className="mt-2 text-xs text-muted-foreground flex items-center gap-1">
                                        <Download className="h-3 w-3"/>
                                        Output: {task.output_path}
                                    </div>
                                )}
                                {task.error_message && (
                                    <div className="mt-2 text-xs text-destructive">
                                        Error: {task.error_message}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        );
    };

    return (
        <div className="space-y-6 p-6">
            <h2 className="text-3xl font-bold tracking-tight">Transcoding Dashboard</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {cards.map((card) => (
                    <Card key={card.title}>
                        <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                            <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
                            <card.icon className={`h-4 w-4 ${card.color}`}/>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{card.count}</div>
                            <CardDescription>Total {card.title.toLowerCase()} tasks</CardDescription>
                        </CardContent>
                    </Card>
                ))}
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Media Transcoding Details</CardTitle>
                    <CardDescription>Detailed view of all media items and their encoding tasks (auto-refreshes every 5
                        seconds)</CardDescription>
                </CardHeader>
                <CardContent>
                    {(!status?.items || status.items.length === 0) ? (
                        <div className="text-center py-8 text-muted-foreground">
                            No transcoding tasks found. Upload a media file to start transcoding.
                        </div>
                    ) : (
                        <ScrollArea className="h-[600px] pr-4">
                            <div className="space-y-4">
                                {status.items.map((item: TranscodingMediaItem) => renderTaskDetails(item))}
                            </div>
                        </ScrollArea>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
