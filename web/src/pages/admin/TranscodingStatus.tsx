/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Transcoding Status Page — Flat task list
 *
 * Each row = one encoding task (one profile for one media).
 * Supports filtering by status and media_id (from URL param).
 */

import {useEffect, useState, useCallback} from "react";
import {useLocation, useNavigate} from "@tanstack/react-router";
import {mediaApi} from "../../lib/api/media";
import {getAccessToken} from "../../lib/request";
import {useTranscoding} from "../../hooks/useTranscoding";
import {Badge} from "../../components/ui/badge";
import {Button} from "../../components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "../../components/ui/card";
import {Progress} from "../../components/ui/progress";
import {Skeleton} from "../../components/ui/skeleton";
import {
    Table, TableBody, TableCell, TableRow,
    TableHead, TableHeader
} from "../../components/ui/table";
import {Tabs, TabsList, TabsTrigger} from "../../components/ui/tabs";
import {
    Film, ExternalLink, RotateCcw,
    Loader2, Radio, CheckCircle2,
    Clock, XCircle, Video
} from "lucide-react";

// ─── Types ─────────────────────────────────────────────

type StatusFilter = "all" | "pending" | "success" | "skipped";

interface EncodingTask {
    id: number;
    media_id: number;
    profile_id: number;
    profile_name?: string;
    status: string;
    progress: number;
    output_path: string;
    error_message: string;
    created_at?: any;
    update_time?: any;
}

interface EncodingTaskListResponse {
    processing_count: number;
    pending_count: number;
    partial_count: number;
    failed_count: number;
    success_count: number;
    total_filtered: number;
    page: number;
    page_size: number;
    items: EncodingTask[];
}

// ─── Helpers ──────────────────────────────────────────

const getApiBaseUrl = () => (import.meta as any).env?.VITE_API_BASE_URL || "http://localhost:9090";

/** Build link to Media page that searches for this media ID */
function mediaLink(mediaId: number): string {
    return `/admin/media?q=%23${mediaId}`;
}

const formatTime = (ts?: any): string => {
    if (!ts) return "--";
    if (typeof ts === "object" && ts !== null && "seconds" in ts) {
        const sec = typeof ts.seconds === "string" ? parseInt(ts.seconds, 10) : ts.seconds;
        const d = new Date((sec || 0) * 1000);
        if (!isNaN(d.getTime())) return d.toLocaleString("zh-CN", {
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit"
        });
        return "--";
    }
    if (typeof ts === "string" || typeof ts === "number") {
        const d = new Date(ts);
        if (!isNaN(d.getTime())) return d.toLocaleString("zh-CN", {
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit"
        });
    }
    return String(ts);
};

// ─── Status helpers ───────────────────────────────────

const statusMap: Record<string, { color: string; icon: typeof CheckCircle2; label: string }> = {
    processing: {color: "blue", icon: Loader2, label: "转码中"},
    pending: {color: "yellow", icon: Clock, label: "排队中"},
    success: {color: "green", icon: CheckCircle2, label: "完成"},
    skipped: {color: "red", icon: XCircle, label: "失败"},
};

function getStatus(s: string) {
    return statusMap[s] || {color: "slate", icon: XCircle, label: s};
}

const knownProfiles: Record<number, string> = {
    1: "h265-240", 2: "vp9-240", 3: "h264-240",
    4: "h265-360", 5: "vp9-360", 6: "h264-360",
    7: "h265-480", 8: "vp9-480", 9: "h264-480",
    10: "h265-720", 11: "vp9-720", 12: "h264-720",
    13: "h265-1080", 14: "vp9-1080", 15: "h264-1080",
    22: "preview",
};

function profileName(task: EncodingTask): string {
    return task.profile_name ?? knownProfiles[task.profile_id] ?? `profile-${task.profile_id}`;
}

// ─── Color map for status badges ──────────────────────

const badgeStyle: Record<string, string> = {
    blue: "bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/30 dark:text-blue-400 dark:border-blue-800",
    yellow: "bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-950/30 dark:text-yellow-400 dark:border-yellow-800",
    green: "bg-green-50 text-green-700 border-green-200 dark:bg-green-950/30 dark:text-green-400 dark:border-green-800",
    red: "bg-red-50 text-red-700 border-red-200 dark:bg-red-950/30 dark:text-red-400 dark:border-red-800",
    slate: "bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700",
};

// ─── Task Row Component ──────────────────────────────

function TaskRow({
                     task,
                     onRetry,
                     isRetrying,
                 }: {
    task: EncodingTask;
    onRetry: () => void;
    isRetrying: boolean;
}) {
    const st = getStatus(task.status);
    const StIcon = st.icon;

    const isProcessing = task.status === "processing";
    const isSkipped = task.status === "skipped";
    const isSuccess = task.status === "success";
    const canRetry = isSkipped;

    return (
        <TableRow className="group hover:bg-slate-50/80 dark:hover:bg-slate-800/50 transition-colors">
            {/* Media */}
            <TableCell className="w-[220px]">
                <a href={mediaLink(task.media_id)}
                   className="group/media inline-flex items-center gap-2.5 text-sm font-medium text-slate-900 dark:text-slate-100 hover:text-blue-600 dark:hover:text-blue-400">
                    <span
                        className="w-8 h-8 rounded-lg bg-slate-100 dark:bg-slate-800 flex items-center justify-center shrink-0 group-hover/media:bg-blue-50 dark:group-hover/media:bg-blue-950/30 transition-colors">
                        <Video
                            className="w-3.5 h-3.5 text-slate-400 group-hover/media:text-blue-500 transition-colors"/>
                    </span>
                    <div className="min-w-0">
                        <span className="block truncate max-w-[140px]">#{task.media_id}</span>
                    </div>
                    <ExternalLink
                        className="w-3 h-3 opacity-0 group-hover/media:opacity-100 text-blue-500 shrink-0 transition-opacity"/>
                </a>
            </TableCell>

            {/* Profile */}
            <TableCell className="w-[120px]">
                <Badge variant="outline"
                       className="font-mono text-[11px] px-2 py-0 h-6 border-dashed border-slate-300 dark:border-slate-600">
                    {profileName(task)}
                </Badge>
            </TableCell>

            {/* Status */}
            <TableCell className="w-[100px]">
                <span
                    className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border whitespace-nowrap ${badgeStyle[st.color] ?? badgeStyle.slate}`}>
                    <StIcon className="w-3 h-3 shrink-0"/>
                    {st.label}
                </span>
            </TableCell>

            {/* Detail: Progress / Error / Output */}
            <TableCell className="min-w-[200px] max-w-[280px]">
                {isProcessing && (
                    <div className="flex items-center gap-3">
                        <Progress value={task.progress} className="h-1.5 flex-1"/>
                        <span
                            className="text-[11px] tabular-nums text-muted-foreground w-8 text-right">{task.progress}%</span>
                    </div>
                )}
                {isSuccess && task.output_path && (
                    <span
                        className="text-[11px] text-emerald-600 dark:text-emerald-400 bg-emerald-50/80 dark:bg-emerald-950/20 px-2 py-1 rounded font-mono block truncate"
                        title={task.output_path}>
                        {task.output_path.split("/").pop() || task.output_path}
                    </span>
                )}
                {isSkipped && task.error_message && (
                    <span className="text-xs text-red-500 block truncate max-w-[260px]" title={task.error_message}>
                        {task.error_message}
                    </span>
                )}
                {isSuccess && !task.output_path && (
                    <span className="text-xs text-emerald-600">✓ Completed</span>
                )}
            </TableCell>

            {/* Time + Actions */}
            <TableCell className="w-[180px]">
                <div className="flex items-center justify-between gap-2">
                    <span className="text-[11px] text-muted-foreground tabular-nums whitespace-nowrap">
                        {formatTime(task.update_time || task.created_at)}
                    </span>
                    {canRetry && (
                        <Button variant="ghost" size="sm"
                                className="h-7 text-[11px] px-2 text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:text-orange-400 dark:hover:bg-orange-950/20 opacity-0 group-hover:opacity-100 transition-opacity"
                                disabled={isRetrying}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onRetry();
                                }}>
                            {isRetrying
                                ? <Loader2 className="w-3 h-3 animate-spin mr-1"/>
                                : <RotateCcw className="w-3 h-3 mr-1"/>}
                            重试
                        </Button>
                    )}
                </div>
            </TableCell>
        </TableRow>
    );
}

// ─── Main Page ───────────────────────────────────────

export default function TranscodingStatus() {
    const location = useLocation();
    const navigate = useNavigate();
    const urlMediaId = new URLSearchParams(location.search).get("media_id");

    const [data, setData] = useState<EncodingTaskListResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<StatusFilter>("all");
    const [page, setPage] = useState(1);
    const [retryingTaskId, setRetryingTaskId] = useState<number | null>(null);

    const {lastEvent, sseStatus} = useTranscoding(undefined);

    const fetchTasks = useCallback(async (filter: StatusFilter, pageNum: number, mediaId?: string | null) => {
        try {
            const response = await mediaApi.getEncodingTasks({
                status: filter,
                page: pageNum,
                page_size: 25,
                media_id: mediaId ? parseInt(mediaId, 10) : undefined,
            });
            setData(response as unknown as EncodingTaskListResponse);
        } catch (error) {
            console.error("Failed to fetch encoding tasks:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        setLoading(true);
        fetchTasks(activeTab, page, urlMediaId);
    }, [activeTab, page, urlMediaId, fetchTasks]);

    // SSE-driven refresh
    useEffect(() => {
        if (!lastEvent) return;
        if (urlMediaId && lastEvent.media_id !== parseInt(urlMediaId, 10)) return;
        if (activeTab === "all" || activeTab === "pending") {
            const debounce = setTimeout(() => fetchTasks(activeTab, page, urlMediaId), 500);
            return () => clearTimeout(debounce);
        }
    }, [lastEvent, activeTab, page, urlMediaId, fetchTasks]);

    // Retry handler
    const handleRetryTask = async (taskId: number) => {
        setRetryingTaskId(taskId);
        try {
            const base = getApiBaseUrl();
            const token = getAccessToken();
            const resp = await fetch(`${base}/api/v1/media/retry?task_id=${taskId}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    ...(token ? {Authorization: `Bearer ${token}`} : {}),
                },
            });
            if (!resp.ok) {
                const err = await resp.json();
                throw new Error(err.error || "Retry failed");
            }
            setTimeout(() => fetchTasks(activeTab, page, urlMediaId), 1000);
        } catch (err: any) {
            console.error("Retry task failed:", err.message);
        } finally {
            setRetryingTaskId(null);
        }
    };

    const handleTabChange = (value: string) => {
        setActiveTab(value as StatusFilter);
        setPage(1);
    };

    const clearMediaFilter = () => {
        const params = new URLSearchParams(location.search);
        params.delete("media_id");
        navigate({search: params.toString()});
    };

    // ─── Tab definitions ──
    const tabs: { value: StatusFilter; label: string; count: number }[] = [
        {value: "all", label: "全部", count: data?.total_filtered ?? 0},
        {value: "pending", label: "排队", count: data?.pending_count ?? 0},
        {value: "success", label: "完成", count: data?.success_count ?? 0},
        {value: "skipped", label: "失败", count: data?.failed_count ?? 0},
    ];

    // ─── Loading skeleton ─────────────────────────────
    if (loading && !data) {
        return (
            <div className="space-y-4 p-4 md:p-6 max-w-6xl mx-auto">
                <Skeleton className="h-7 w-48"/>
                <Skeleton className="h-24 w-full rounded-xl"/>
                <Card><CardContent className="pt-4"><Skeleton className="h-48 w-full"/></CardContent></Card>
            </div>
        );
    }

    const totalPages = Math.ceil((data?.total_filtered ?? 0) / (data?.page_size ?? 25));

    return (
        <div className="space-y-4 p-4 md:p-6 max-w-6xl mx-auto">

            {/* ═══ Header ════════════════════════════════ */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                <div>
                    <h2 className="text-xl font-bold tracking-tight">Transcoding Tasks</h2>
                    <p className="text-sm text-muted-foreground mt-0.5">
                        Monitor and manage all encoding tasks across media
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Badge variant="outline" className="gap-1.5 text-xs">
                        <Radio
                            className={`h-3 w-3 ${sseStatus.connected ? "text-green-500" : "text-muted-foreground"}`}/>
                        {sseStatus.connected ? "Live" : "Offline"}
                    </Badge>
                    {urlMediaId && (
                        <Badge variant="secondary" className="gap-1 text-xs">
                            <Film className="w-3 h-3"/>
                            #{urlMediaId}
                            <button onClick={clearMediaFilter} className="ml-1 hover:text-destructive leading-none"
                                    title="Clear filter">×
                            </button>
                        </Badge>
                    )}
                </div>
            </div>

            {/* ═══ Stats Cards ════════════════════════════ */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
                {[
                    {label: "Total", value: data?.total_filtered ?? 0, color: "text-slate-700 dark:text-slate-300"},
                    {label: "Pending", value: data?.pending_count ?? 0, color: "text-amber-600 dark:text-amber-400"},
                    {label: "Done", value: data?.success_count ?? 0, color: "text-emerald-600 dark:text-emerald-400"},
                    {label: "Failed", value: data?.failed_count ?? 0, color: "text-red-600 dark:text-red-400"},
                ].map((card) => (
                    <div key={card.label} className="rounded-lg border bg-card p-3 flex items-center justify-between">
                        <div>
                            <p className="text-[11px] uppercase tracking-wide text-muted-foreground font-medium">{card.label}</p>
                            <p className={`text-lg font-bold tabular-nums ${card.color}`}>{card.value}</p>
                        </div>
                    </div>
                ))}
            </div>

            {/* ═══ Tabs ═════════════════════════════════ */}
            <Tabs value={activeTab} onValueChange={handleTabChange}>
                <TabsList className="h-8 w-full justify-start bg-muted/50">
                    {tabs.map((t) => (
                        <TabsTrigger key={t.value} value={t.value}
                                     className="text-xs px-3 h-7 data-[state=active]:bg-background data-[state=active]:shadow-sm">
                            {t.label}
                            <span className="ml-1.5 tabular-nums text-[11px] opacity-60">({t.count})</span>
                        </TabsTrigger>
                    ))}
                </TabsList>
            </Tabs>

            {/* ═══ Task Table ═════════════════════════════ */}
            <Card>
                <CardContent className="pt-4 pb-0">
                    {!data?.items?.length ? (
                        <div className="py-12 text-center">
                            <Film className="w-10 h-10 mx-auto mb-2 opacity-15"/>
                            <p className="text-sm text-muted-foreground">
                                No tasks found{urlMediaId ? ` for Media #${urlMediaId}` : ""}
                            </p>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow className="border-b hover:bg-transparent">
                                    <TableHead
                                        className="w-[220px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Media</TableHead>
                                    <TableHead
                                        className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Profile</TableHead>
                                    <TableHead
                                        className="w-[100px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Status</TableHead>
                                    <TableHead
                                        className="py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Detail</TableHead>
                                    <TableHead
                                        className="w-[180px] py-2.5 text-right text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Time
                                        / Action</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {data.items.map((task) => (
                                    <TaskRow key={task.id}
                                             task={task}
                                             onRetry={() => handleRetryTask(task.id)}
                                             isRetrying={retryingTaskId === task.id}/>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            {/* ═══ Pagination ══════════════════════════════ */}
            {data && data.total_filtered > data.page_size && (
                <div className="flex items-center justify-between pt-1 text-xs text-muted-foreground">
                    <span className="tabular-nums">
                        Page {data.page} of {totalPages} · {data.total_filtered} total
                    </span>
                    <div className="flex gap-1.5">
                        <Button variant="outline" size="sm" className="h-7 text-xs px-2"
                                disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
                            ← Prev
                        </Button>
                        <Button variant="outline" size="sm" className="h-7 text-xs px-2"
                                disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                            Next →
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}
