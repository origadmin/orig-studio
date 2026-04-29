/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Transcoding Status Page — Enhanced with better UI/UX
 *
 * Each row = one encoding task (one profile for one media).
 * Supports filtering by status and media_id (from URL param).
 */

import {useEffect, useState, useCallback, useMemo, useRef} from "react";
import {useLocation, useNavigate} from "@tanstack/react-router";
import {mediaApi, encodingApi, type EncodeProfile} from "../../lib/api/media";
import {API_BASE_URL} from "../../lib/request";
import {useAuth} from "../../hooks/useAuth";
import {useTranscoding} from "../../hooks/useTranscoding";
import {Badge} from "../../components/ui/badge";
import {Button} from "../../components/ui/button";
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from "../../components/ui/card";
import {Progress} from "../../components/ui/progress";
import {Skeleton} from "../../components/ui/skeleton";
import {
    Table, TableBody, TableCell, TableRow,
    TableHead, TableHeader
} from "../../components/ui/table";
import {Tabs, TabsList, TabsTrigger} from "../../components/ui/tabs";
import {Input} from "../../components/ui/input";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "../../components/ui/select";
import {Checkbox} from "../../components/ui/checkbox";
import {Separator} from "../../components/ui/separator";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "../../components/ui/dropdown-menu";
import {
    Film, ExternalLink, RotateCcw,
    Loader2, Radio, CheckCircle2,
    Clock, XCircle, Video, Search,
    Filter, MoreVertical, Trash2, Play,
    Pause, Settings, ArrowUpDown, Download, AlertCircle
} from "lucide-react";
import {formatDate, formatRelativeTime} from "../../lib/format";
import {TablePagination} from '@/components/common/TablePagination';

// ─── Types ─────────────────────────────────────────────

type StatusFilter = "all" | "pending" | "processing" | "success" | "failed" | "skipped";

interface EncodingTask {
    id: number | string;
    media_id: string;
    media_title?: string;
    thumbnail?: string;
    profile_id: number;
    profile_name?: string;
    status: string;
    output_path: string;
    error_message: string;
    created_at?: any;
    update_time?: any;
    progress?: number;
    speed?: string;
    fps?: number;
    time?: number;
}

interface EncodingTaskListResponse {
    processing_count: number;
    pending_count: number;
    partial_count: number;
    failed_count: number;
    success_count: number;
    total: number;
    page: number;
    page_size: number;
    items: EncodingTask[];
}

// ─── Helpers ──────────────────────────────────────────

/** Build link to Media page that searches for this media ID */
function mediaLink(mediaId: string): string {
    return `/admin/media?q=%23${mediaId}`;
}

const formatTime = (ts?: any): string => {
    if (!ts) return "--";
    return formatDate(ts);
};

// ─── Status helpers ───────────────────────────────────

const statusMap: Record<string, { color: string; icon: typeof CheckCircle2; label: string }> = {
    processing: {color: "sky", icon: Loader2, label: "转码中"},
    pending: {color: "amber", icon: Clock, label: "排队中"},
    success: {color: "emerald", icon: CheckCircle2, label: "完成"},
    skipped: {color: "rose", icon: AlertCircle, label: "跳过"},
    failed: {color: "rose", icon: XCircle, label: "失败"},
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
    sky: "bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950/30 dark:text-sky-400 dark:border-sky-800",
    amber: "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/30 dark:text-amber-400 dark:border-amber-800",
    emerald: "bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/30 dark:text-emerald-400 dark:border-emerald-800",
    rose: "bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/30 dark:text-rose-400 dark:border-rose-800",
    slate: "bg-slate-50 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-muted-foreground dark:border-slate-700",
};

// ─── Task Row Component ──────────────────────────────

function TaskRow({
                     task,
                     onRetry,
                     isRetrying,
                     isSelected,
                     onToggleSelect,
                 }: {
    task: EncodingTask;
    onRetry: () => void;
    isRetrying: boolean;
    isSelected: boolean;
    onToggleSelect: () => void;
}) {
    const st = getStatus(task.status);
    const StIcon = st.icon;
    const [showError, setShowError] = useState(false);
    const [thumbError, setThumbError] = useState(false);

    const isProcessing = task.status === "processing";
    const isSkipped = task.status === "skipped";
    const isFailed = task.status === "failed";
    const isSuccess = task.status === "success";
    const canRetry = isSkipped || isFailed;

    const thumbSrc = task.thumbnail && !thumbError
        ? (task.thumbnail.startsWith('http') ? task.thumbnail : `${API_BASE_URL}${task.thumbnail.startsWith('/') ? '' : '/'}${task.thumbnail}`)
        : '';

    return (
        <>
            <TableRow
                key={`task-${task.id}`}
                className={`group hover:bg-slate-50/80 dark:hover:bg-slate-800/50 transition-colors ${isSelected ? 'bg-muted/30' : ''}`}
                style={{height: '64px'}}>
                <TableCell className="w-[50px]">
                    <Checkbox
                        checked={isSelected}
                        onCheckedChange={onToggleSelect}
                        aria-label={`Select task ${task.id}`}
                    />
                </TableCell>
                {/* ENCODING */}
                <TableCell className="max-w-[320px]">
                    <a href={mediaLink(task.media_id)}
                       className="group/media flex items-center gap-3 text-sm font-medium text-slate-900 dark:text-slate-100 hover:text-info dark:hover:text-blue-400 transition-all duration-200">
                        <span
                            className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-50 to-slate-100 dark:from-blue-950/30 dark:to-slate-800 flex items-center justify-center shrink-0 group-hover/media:from-blue-100 group-hover/media:to-blue-50 dark:group-hover/media:from-blue-900/30 dark:group-hover/media:to-blue-950/50 transition-all duration-200 shadow-sm overflow-hidden">
                            {thumbSrc ? (
                                <img src={thumbSrc}
                                     alt="" className="w-full h-full object-cover"
                                     onError={() => setThumbError(true)}/>
                            ) : (
                                <Video className="w-4 h-4 text-info group-hover/media:text-info dark:group-hover/media:text-blue-400 transition-colors"/>
                            )}
                        </span>
                        <div className="flex-1 min-w-0">
                            {task.media_title && <span className="block truncate font-medium text-slate-900 dark:text-slate-100" title={task.media_title}>{task.media_title}</span>}
                            <span className="block font-mono text-[11px] leading-tight text-muted-foreground" title={task.media_id}>{task.media_id}</span>
                        </div>
                        <ExternalLink
                            className="w-3.5 h-3.5 opacity-0 group-hover/media:opacity-100 text-info shrink-0 transition-opacity"/>
                    </a>
                </TableCell>

                {/* PROFILE */}
                <TableCell className="w-[120px] whitespace-nowrap">
                    <Badge variant="outline"
                           className="font-mono text-[11px] px-2 py-0 h-6 border-dashed border-slate-300 dark:border-slate-600 whitespace-nowrap">
                        {profileName(task)}
                    </Badge>
                </TableCell>

                {/* PROGRESS */}
                <TableCell className="w-[150px]">
                    {isProcessing && (
                        <div className="flex flex-col gap-1.5">
                            <div className="flex items-center justify-between">
                                <span
                                    className="text-[10px] font-bold text-sky-600 dark:text-sky-400 uppercase tracking-tight">Processing</span>
                                <span
                                    className="text-[10px] tabular-nums font-bold text-slate-600 dark:text-muted-foreground">{task.progress ?? 0}%</span>
                            </div>
                            <div className="h-2.5 w-full bg-sky-100 dark:bg-sky-950/30 rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-sky-600 transition-all duration-500 ease-out animate-pulse"
                                    style={{width: `${Math.min(Math.max(task.progress ?? 0, 0), 100)}%`}}
                                />
                            </div>
                        </div>
                    )}
                    {isSuccess && (
                        <div className="flex flex-col gap-1.5">
                            <div className="flex items-center justify-between">
                                <span
                                    className="text-[10px] font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-tight">Completed</span>
                                <span
                                    className="text-[10px] tabular-nums font-bold text-emerald-600 dark:text-emerald-400">100%</span>
                            </div>
                            <div
                                className="h-2.5 w-full bg-emerald-100 dark:bg-emerald-950/30 rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-emerald-600"
                                    style={{width: '100%'}}
                                />
                            </div>
                        </div>
                    )}
                    {!isProcessing && !isSuccess && (
                        <div className="flex flex-col gap-1.5 opacity-40">
                            <div className="flex items-center justify-between">
                                <span
                                    className="text-[10px] font-bold uppercase tracking-tight">{task.status === 'pending' ? 'Queued' : 'Waiting'}</span>
                                <span className="text-[10px] tabular-nums font-bold text-slate-600 dark:text-muted-foreground">0%</span>
                            </div>
                            <div className="h-2.5 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden"/>
                        </div>
                    )}
                </TableCell>

                {/* STATUS */}
                <TableCell className="w-[120px]">
                    <span
                        className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium border whitespace-nowrap ${badgeStyle[st.color] ?? badgeStyle.slate}`}>
                        <StIcon className="w-3 h-3 shrink-0"/>
                        {st.label}
                    </span>
                </TableCell>

                {/* HAS FILE */}
                <TableCell className="w-[80px]">
                    {task.output_path ? (
                        <Badge className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300 border-2 border-emerald-200 dark:border-emerald-800">
                            Yes
                        </Badge>
                    ) : (
                        <Badge variant="outline" className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-slate-100 text-slate-800 dark:bg-slate-900/30 dark:text-slate-300 border-slate-200">
                            No
                        </Badge>
                    )}
                </TableCell>

                {/* Time */}
                <TableCell className="w-[120px] text-right">
                    <span className="text-[11px] text-muted-foreground tabular-nums whitespace-nowrap">
                        {formatRelativeTime(task.update_time || task.create_time)}
                    </span>
                </TableCell>

                {/* Action */}
                <TableCell className="w-[180px] text-right">
                    <div className="flex items-center justify-end gap-1">
                        {canRetry && (
                            <Button variant="ghost" size="icon-sm"
                                    title={showError ? 'Hide Details' : 'Show Details'}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowError(!showError);
                                    }}>
                                <AlertCircle className="h-3 w-3"/>
                            </Button>
                        )}
                        {canRetry && (
                            <Button variant="ghost" size="icon-sm" title="Retry"
                                    disabled={isRetrying}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onRetry();
                                    }}>
                                {isRetrying
                                    ? <Loader2 className="w-3 h-3 animate-spin"/>
                                    : <RotateCcw className="w-3 h-3"/>}
                            </Button>
                        )}
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon-sm" title="More Actions">
                                    <MoreVertical className="h-3 w-3"/>
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                <DropdownMenuSeparator/>
                                {task.output_path && (
                                    <DropdownMenuItem>
                                        <Download className="h-4 w-4 mr-2"/>
                                        Download Output
                                    </DropdownMenuItem>
                                )}
                                <DropdownMenuItem className="text-destructive focus:text-destructive">
                                    <Trash2 className="h-4 w-4 mr-2"/>
                                    Delete Task
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </TableCell>
            </TableRow>
            {/* Error message row */}
            {showError && (isSkipped || isFailed) && task.error_message && (
                <TableRow key={`error-${task.id}`}>
                    <TableCell colSpan={9}
                               className="bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-800">
                        <div className="p-4">
                            <div className="flex items-start gap-3">
                                <div className="flex-shrink-0">
                                    <div
                                        className="w-8 h-8 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
                                        <svg className="w-4 h-4 text-destructive dark:text-red-400" fill="none"
                                             viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                        </svg>
                                    </div>
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-start gap-2 mb-2">
                                        <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mt-0.5">
                                            {isFailed ? 'Encoding Failed' : 'Task Skipped'}
                                        </h4>
                                        <span
                                            className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                                                isFailed
                                                    ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300'
                                                    : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300'
                                            }`}>
                                            {task.status}
                                        </span>
                                    </div>
                                    <div className="bg-gray-900 dark:bg-black rounded-lg p-3 overflow-auto max-h-40">
                                        <pre
                                            className="text-xs font-mono text-gray-300 whitespace-pre-wrap break-words">
                                            {task.error_message}
                                        </pre>
                                    </div>
                                    {profileName(task) === 'preview' && (
                                        <div
                                            className="mt-3 flex items-start gap-2 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
                                            <svg
                                                className="w-4 h-4 text-info dark:text-blue-400 flex-shrink-0 mt-0.5"
                                                fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                            </svg>
                                            <p className="text-xs text-blue-700 dark:text-blue-300">
                                                Preview generation failed. This may be due to invalid input file or
                                                missing dependencies.
                                            </p>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    </TableCell>
                </TableRow>
            )}
        </>
    );
}

// ─── Main Page ───────────────────────────────────────

export default function TranscodingStatus() {
    const location = useLocation();
    const navigate = useNavigate();
    const urlMediaId = new URLSearchParams(location.search).get("media_id");
    const {token, isAuthenticated} = useAuth();

    // State for filtered tasks and stats (used for filter buttons and task list)
    const [filteredData, setFilteredData] = useState<EncodingTaskListResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<StatusFilter>("all");
    const [page, setPage] = useState(1);
    const [retryingTaskId, setRetryingTaskId] = useState<number | string | null>(null);
    const [encodeProfiles, setEncodeProfiles] = useState<EncodeProfile[]>([]);

    // Applied filters (used for fetching data)
    const [searchQuery, setSearchQuery] = useState('');
    const [profileFilter, setProfileFilter] = useState<string>('');
    const [statusFilter, setStatusFilter] = useState<StatusFilter | ''>('');
    const [chunkFilter, setChunkFilter] = useState<string>('');

    // Pending filters (used in UI inputs)
    const [pendingSearchQuery, setPendingSearchQuery] = useState('');
    const [pendingProfileFilter, setPendingProfileFilter] = useState<string>('');
    const [pendingStatusFilter, setPendingStatusFilter] = useState<StatusFilter | ''>('');
    const [pendingChunkFilter, setPendingChunkFilter] = useState<string>('');

    const [selectedRows, setSelectedRows] = useState<(string | number)[]>([]);
    const [sortConfig, setSortConfig] = useState<{ key: keyof EncodingTask, direction: 'asc' | 'desc' } | null>(null);
    
    // Total stats for card area (always complete, never filtered)
    const [totalStats, setTotalStats] = useState<{
        total: number;
        pending: number;
        success: number;
        failed: number;
    }>({
        total: 0,
        pending: 0,
        success: 0,
        failed: 0,
    });

    const {lastEvent, sseStatus} = useTranscoding(urlMediaId ?? undefined);

    // Fetch total stats (card area) - always complete data
    const fetchTotalStats = useCallback(async () => {
        try {
            const response = await encodingApi.getTasks({
                only_stats: true
            });
            setTotalStats({
                total: response.processing_count + response.pending_count + response.success_count + response.failed_count,
                pending: response.pending_count,
                success: response.success_count,
                failed: response.failed_count
            });
        } catch (error) {
            console.error("Failed to fetch total stats:", error);
        }
    }, []);

    // Fetch encoding profiles for filter dropdown
    const fetchEncodeProfiles = useCallback(async () => {
        try {
            const response = await encodingApi.profiles.list();
            setEncodeProfiles(response.profiles || []);
        } catch (error) {
            console.error("Failed to fetch encode profiles:", error);
        }
    }, []);

    // Fetch filtered tasks and stats (filter buttons and task list)
    const fetchFilteredTasks = useCallback(async (
        status: StatusFilter | '', 
        pageNum: number, 
        mediaId?: string | null,
        profile?: string,
        chunk?: string,
        search?: string
    ) => {
        try {
            const params: any = {
                page: pageNum,
                page_size: 25,
            };
            
            if (status !== '') {
                params.status = status;
            }
            
            if (mediaId) {
                params.media_id = mediaId;
            }
            if (profile && profile !== '') {
                params.profile = profile;
            }
            if (chunk && chunk !== '') {
                params.chunk = chunk;
            }
            if (search && search !== '') {
                params.search = search;
            }
            
            const response = await encodingApi.getTasks(params);
            setFilteredData(response as unknown as EncodingTaskListResponse);
        } catch (error) {
            console.error("Failed to fetch filtered tasks:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    // Keep a stable reference to fetchFilteredTasks for setTimeout
    const fetchFilteredTasksRef = useRef(fetchFilteredTasks);
    useEffect(() => {
        fetchFilteredTasksRef.current = fetchFilteredTasks;
    }, [fetchFilteredTasks]);

    // Search and Reset handlers
    const handleSearch = () => {
        setSearchQuery(pendingSearchQuery);
        setProfileFilter(pendingProfileFilter);
        setStatusFilter(pendingStatusFilter);
        setChunkFilter(pendingChunkFilter);
        if (pendingStatusFilter !== '') {
            setActiveTab(pendingStatusFilter as StatusFilter);
        } else {
            setActiveTab('all');
        }
        setPage(1);
        setLoading(true);
        fetchFilteredTasks(
            pendingStatusFilter, 
            1, 
            urlMediaId, 
            pendingProfileFilter, 
            pendingChunkFilter, 
            pendingSearchQuery
        );
    };

    const handleReset = () => {
        setPendingSearchQuery('');
        setPendingProfileFilter('');
        setPendingStatusFilter('');
        setPendingChunkFilter('');

        setSearchQuery('');
        setProfileFilter('');
        setStatusFilter('');
        setChunkFilter('');
        setActiveTab('all');
        setPage(1);
        setLoading(true);
        fetchFilteredTasks('', 1, urlMediaId, '', '', '');
    };

    // Handle tab change (status filter)
    const handleTabChange = (filterVal: StatusFilter) => {
        setActiveTab(filterVal);
        const status = filterVal === 'all' ? '' : filterVal;
        setStatusFilter(status);
        setPendingStatusFilter(status);
        setPage(1);
        setSelectedRows([]);
        setLoading(true);
        fetchFilteredTasks(status, 1, urlMediaId, profileFilter, chunkFilter, searchQuery);
    };

    // Fetch data on component mount
    useEffect(() => {
        setLoading(true);
        // Fetch total stats (card area) - only once on mount
        fetchTotalStats();
        // Fetch filtered tasks (initial state)
        fetchFilteredTasks(statusFilter, page, urlMediaId, profileFilter, chunkFilter, searchQuery);
        // Fetch encoding profiles
        fetchEncodeProfiles();
    }, []);

    // Refresh total stats periodically (every 60 seconds)
    useEffect(() => {
        const interval = setInterval(fetchTotalStats, 60000);
        return () => clearInterval(interval);
    }, [fetchTotalStats]);

    // Smart polling: only poll when SSE is disconnected
    useEffect(() => {
        let interval: NodeJS.Timeout | null = null;

        if (!sseStatus.connected) {
            interval = setInterval(() => {
                fetchFilteredTasks(statusFilter, page, urlMediaId, profileFilter, chunkFilter, searchQuery);
            }, 30000); // Poll every 30 seconds when SSE is disconnected
        }

        return () => {
            if (interval) clearInterval(interval);
        };
    }, [sseStatus.connected, statusFilter, page, urlMediaId, profileFilter, chunkFilter, searchQuery, fetchFilteredTasks]);

    // Fetch data when page changes
    useEffect(() => {
        fetchFilteredTasks(statusFilter, page, urlMediaId, profileFilter, chunkFilter, searchQuery);
    }, [page, statusFilter, urlMediaId, profileFilter, chunkFilter, searchQuery, fetchFilteredTasks]);
    // SSE event handling - only update changed data
    useEffect(() => {
        if (!lastEvent) return;
        if (urlMediaId && lastEvent.media_id !== urlMediaId) return;

        // Optimization: only update changed tasks instead of entire list
        setFilteredData(prev => {
            if (!prev || !prev.items) return prev;

            const updatedItems = prev.items.map(task => {
                // Compare task.id (string) with lastEvent.task_id (string)
                if (String(task.id) === String(lastEvent.task_id)) {
                    const updatedTask: EncodingTask = {
                        ...task,
                        status: lastEvent.status,
                        update_time: new Date().toISOString(),
                    };
                    
                    // Update progress fields if available
                    if (lastEvent.progress !== undefined) {
                        updatedTask.progress = lastEvent.progress;
                    }
                    if (lastEvent.speed !== undefined) {
                        updatedTask.speed = lastEvent.speed;
                    }
                    if (lastEvent.fps !== undefined) {
                        updatedTask.fps = lastEvent.fps;
                    }
                    if (lastEvent.time !== undefined) {
                        updatedTask.time = lastEvent.time;
                    }
                    
                    // When task is success, refetch to get output_path
                    if (lastEvent.status === 'success') {
                        // Refetch data in 1 second to get updated output_path
                        setTimeout(() => {
                            fetchFilteredTasksRef.current(statusFilter, page, urlMediaId, profileFilter, chunkFilter, searchQuery);
                        }, 1000);
                    }
                    
                    return updatedTask;
                }
                return task;
            });

            // Update statistics
            const pendingCount = updatedItems.filter(t => t.status === 'pending').length;
            const successCount = updatedItems.filter(t => t.status === 'success').length;
            const failedCount = updatedItems.filter(t => t.status === 'failed' || t.status === 'skipped').length;
            const processingCount = updatedItems.filter(t => t.status === 'processing').length;

            return {
                ...prev,
                items: updatedItems,
                pending_count: pendingCount,
                success_count: successCount,
                failed_count: failedCount,
                processing_count: processingCount
            };
        });
    }, [lastEvent, urlMediaId, statusFilter, page, profileFilter, chunkFilter, searchQuery]);

    // Sort logic (filtering is handled by backend)
const filteredTasks = useMemo(() => {
    if (!filteredData?.items) return [];
    
    // Remove duplicates by id
    const uniqueTasks = Array.from(new Map(filteredData.items.map(task => [task.id, task])).values());
    
    let result = [...uniqueTasks];

    // Sorting
    if (sortConfig) {
        result.sort((a, b) => {
            const aVal = a[sortConfig.key];
            const bVal = b[sortConfig.key];

            if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
            if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
            return 0;
        });
    }

    return result;
}, [filteredData?.items, sortConfig]);

    // 获取所有可用的配置
    const availableProfiles = useMemo(() => {
        return encodeProfiles.map(profile => profile.name).sort();
    }, [encodeProfiles]);

    const handleSort = (key: keyof EncodingTask) => {
        setSortConfig(current => {
            if (current?.key === key) {
                return {key, direction: current.direction === 'asc' ? 'desc' : 'asc'};
            }
            return {key, direction: 'asc'};
        });
    };

    // 批量选择
    const toggleSelectAll = () => {
        if (selectedRows.length === filteredTasks.length) {
            setSelectedRows([]);
        } else {
            setSelectedRows(filteredTasks.map(t => t.id));
        }
    };

    const toggleSelectRow = (id: string | number) => {
        setSelectedRows(prev =>
            prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
        );
    };

    // 批量操作
    const handleBatchRetry = async () => {
        // 立即更新本地任务状态，提供更好的用户体验
        setFilteredData(prev => {
            if (!prev) return prev;

            const updatedItems = prev.items.map(task => {
                if (selectedRows.includes(task.id)) {
                    return {
                        ...task,
                        status: "pending",
                        update_time: new Date().toISOString()
                    };
                }
                return task;
            });

            // 更新统计数据
            const pendingCount = updatedItems.filter(t => t.status === 'pending').length;
            const successCount = updatedItems.filter(t => t.status === 'success').length;
            const failedCount = updatedItems.filter(t => t.status === 'failed' || t.status === 'skipped').length;
            const processingCount = updatedItems.filter(t => t.status === 'processing').length;

            return {
                ...prev,
                items: updatedItems,
                pending_count: pendingCount,
                success_count: successCount,
                failed_count: failedCount,
                processing_count: processingCount
            };
        });

        for (const id of selectedRows) {
            setRetryingTaskId(id);
            try {
                await encodingApi.retryTask(String(id));
            } catch (err: any) {
                console.error("Retry task failed:", err.message);
                // 出错时恢复任务状态
                setFilteredData(prev => {
                    if (!prev) return prev;

                    const updatedItems = prev.items.map(task => {
                        if (task.id === id) {
                            return {
                                ...task,
                                status: "failed",
                                update_time: new Date().toISOString()
                            };
                        }
                        return task;
                    });

                    // 更新统计数据
                    const pendingCount = updatedItems.filter(t => t.status === 'pending').length;
                    const successCount = updatedItems.filter(t => t.status === 'success').length;
                    const failedCount = updatedItems.filter(t => t.status === 'failed' || t.status === 'skipped').length;
                    const processingCount = updatedItems.filter(t => t.status === 'processing').length;

                    return {
                        ...prev,
                        items: updatedItems,
                        pending_count: pendingCount,
                        success_count: successCount,
                        failed_count: failedCount,
                        processing_count: processingCount
                    };
                });
            }
        }
        setRetryingTaskId(null);
        setSelectedRows([]);
        // 不再立即刷新，依赖 SSE 事件更新
    };

    // Retry handler
    const handleRetryTask = async (taskId: string) => {
        setRetryingTaskId(taskId);

        // 立即更新本地任务状态，提供更好的用户体验
        setFilteredData(prev => {
            if (!prev) return prev;

            const updatedItems = prev.items.map(task => {
                if (task.id === taskId) {
                    return {
                        ...task,
                        status: "pending",
                        update_time: new Date().toISOString()
                    };
                }
                return task;
            });

            // 更新统计数据
            const pendingCount = updatedItems.filter(t => t.status === 'pending').length;
            const successCount = updatedItems.filter(t => t.status === 'success').length;
            const failedCount = updatedItems.filter(t => t.status === 'failed' || t.status === 'skipped').length;
            const processingCount = updatedItems.filter(t => t.status === 'processing').length;

            return {
                ...prev,
                items: updatedItems,
                pending_count: pendingCount,
                success_count: successCount,
                failed_count: failedCount,
                processing_count: processingCount
            };
        });
        
        try {
            await encodingApi.retryTask(taskId);
            // 不再立即刷新，依赖 SSE 事件更新
        } catch (err: any) {
            console.error("Retry task failed:", err.message);
            // 出错时恢复任务状态
            setFilteredData(prev => {
                if (!prev) return prev;

                const updatedItems = prev.items.map(task => {
                    if (task.id === taskId) {
                        return {
                            ...task,
                            status: "failed",
                            update_time: new Date().toISOString()
                        };
                    }
                    return task;
                });

                // 更新统计数据
                const pendingCount = updatedItems.filter(t => t.status === 'pending').length;
                const successCount = updatedItems.filter(t => t.status === 'success').length;
                const failedCount = updatedItems.filter(t => t.status === 'failed' || t.status === 'skipped').length;
                const processingCount = updatedItems.filter(t => t.status === 'processing').length;

                return {
                    ...prev,
                    items: updatedItems,
                    pending_count: pendingCount,
                    success_count: successCount,
                    failed_count: failedCount,
                    processing_count: processingCount
                };
            });
        } finally {
            setRetryingTaskId(null);
        }
    };

    const clearMediaFilter = () => {
        const params = new URLSearchParams(location.search);
        params.delete("media_id");
        navigate({search: params.toString()});
    };

    // Tab definitions
    const tabs: { value: StatusFilter; label: string; count: number }[] = [
        {value: "all", label: "全部", count: filteredData?.total ?? 0},
        {value: "pending", label: "排队", count: filteredData?.pending_count ?? 0},
        {value: "success", label: "完成", count: filteredData?.success_count ?? 0},
        {value: "failed", label: "失败", count: filteredData?.failed_count ?? 0},
    ];

    // Loading skeleton
    if (loading && !filteredData) {
        return (
            <div className="space-y-4 p-4 md:p-6 max-w-7xl mx-auto">
                <Skeleton className="h-7 w-48"/>
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
                    {[1, 2, 3, 4].map(i => (
                        <Skeleton key={i} className="h-24 w-full rounded-xl"/>
                    ))}
                </div>
                <Card><CardContent className="pt-4"><Skeleton className="h-48 w-full"/></CardContent></Card>
            </div>
        );
    }

    return (
        <div className="space-y-6 p-4 md:p-6">
            {/* ═══ Header ════════════════════════════════ */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题和状态 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">Transcoding
                                    Status</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5 flex items-center gap-2">
                                    <span className="inline-block w-2 h-2 rounded-full bg-sky-500 animate-pulse"/>
                                    Live monitoring of video processing workflows
                                </p>
                            </div>
                            <div className="flex items-center gap-3">
                                <Badge variant="outline"
                                       className={`gap-2 text-[11px] font-bold px-3 py-1 border-2 ${sseStatus.connected ? "text-emerald-500 border-emerald-100 bg-emerald-50/50 dark:bg-emerald-950/20 dark:border-emerald-900" : "text-muted-foreground border-slate-100 bg-slate-50/50 dark:bg-slate-900/20 dark:border-slate-800"}`}>
                                    <div
                                        className={`w-1.5 h-1.5 rounded-full ${sseStatus.connected ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-slate-300"}`}/>
                                    {sseStatus.connected ? "CONNECTED" : "DISCONNECTED"}
                                </Badge>
                                {urlMediaId && (
                                    <Badge variant="secondary"
                                           className="gap-1.5 text-[11px] font-bold px-3 py-1 bg-sky-100 text-sky-700 dark:bg-sky-950/40 dark:text-sky-400 border border-sky-200 dark:border-sky-800">
                                        <Film className="w-3 h-3"/>
                                        MEDIA ID: #{urlMediaId}
                                        <button onClick={clearMediaFilter}
                                                className="ml-1 hover:bg-sky-200 dark:hover:bg-sky-900 rounded-full p-0.5 transition-colors"
                                                title="Clear filter">
                                            <XCircle className="w-3 h-3"/>
                                        </button>
                                    </Badge>
                                )}
                            </div>
                        </div>

                        {/* 分隔线 */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        {/* 搜索和筛选 */}
                        <div className="flex flex-col lg:flex-row gap-4 items-center">
                            <div className="flex-1 min-w-0">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder="Search tasks by media ID, profile, or status..."
                                        value={pendingSearchQuery}
                                        onChange={(e) => setPendingSearchQuery(e.target.value)}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-3">
                                <Select
                                    value={pendingProfileFilter}
                                    onValueChange={(val) => setPendingProfileFilter(val === 'all' ? '' : val)}
                                >
                                    <SelectTrigger
                                        className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            <SelectValue placeholder="Profile"/>
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all"
                                                    className="justify-center text-center font-medium opacity-70">
                                            --- All ---
                                        </SelectItem>
                                        {availableProfiles.map(profile => (
                                            <SelectItem key={profile} value={profile}>
                                                {profile.toUpperCase()}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>

                                <Select
                                    value={pendingChunkFilter}
                                    onValueChange={(val) => setPendingChunkFilter(val === 'all' ? '' : val)}
                                >
                                    <SelectTrigger
                                        className="w-[140px] h-8 rounded-btn-sm focus:ring-1 focus:ring-ring focus:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            <SelectValue placeholder="Chunk"/>
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all"
                                                    className="justify-center text-center font-medium opacity-70">
                                            --- All ---
                                        </SelectItem>
                                        <SelectItem value="true">True (Chunk)</SelectItem>
                                        <SelectItem value="false">False (Full)</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-3 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={handleReset}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2"/>
                                        Reset
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        onClick={handleSearch}
                                    >
                                        <Search className="h-4 w-4 mr-2"/>
                                        Search
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* ═══ Stats Cards ════════════════════════════ */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
                {
                    [
                        {
                            label: "Active Jobs",
                            value: totalStats.total,
                            color: "sky",
                            icon: Video,
                            desc: "Total tasks registered"
                        },
                        {
                            label: "In Queue",
                            value: totalStats.pending,
                            color: "amber",
                            icon: Clock,
                            desc: "Tasks waiting for worker"
                        },
                        {
                            label: "Completed",
                            value: totalStats.success,
                            color: "emerald",
                            icon: CheckCircle2,
                            desc: "Successfully finished"
                        },
                        {
                            label: "Failed",
                            value: totalStats.failed,
                            color: "rose",
                            icon: XCircle,
                            desc: "Tasks requiring attention"
                        },
                    ].map((card, index) => (
                        <Card key={card.label}
                              className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
                            <CardContent className="p-5">
                                <div className="flex items-start justify-between">
                                    <div className="space-y-1">
                                        <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground dark:text-slate-500">{card.label}</p>
                                        <h3 className={`text-3xl font-bold tabular-nums text-${card.color}-600 dark:text-${card.color}-400`}>{card.value}</h3>
                                        <p className="text-[10px] text-muted-foreground dark:text-slate-500 font-medium">{card.desc}</p>
                                    </div>
                                    <div
                                        className={`p-2.5 rounded-xl bg-${card.color}-50 dark:bg-${card.color}-950/30 text-${card.color}-500 dark:text-${card.color}-400`}>
                                        <card.icon className="h-6 w-6"/>
                                    </div>
                                </div>
                                <div className={`absolute bottom-0 left-0 h-1 bg-${card.color}-500 w-full opacity-10`}/>
                            </CardContent>
                        </Card>
                    ))
                }
            </div>

            {/* ═══ Tabs ═════════════════════════════════ */}
            <div className="flex flex-wrap gap-2 my-4">
                {tabs.map((t) => (
                    <Button
                        key={t.value}
                        variant={activeTab === t.value ? "default" : "outline"}
                        size="sm"
                        className="rounded-full px-4 py-1.5"
                        onClick={() => handleTabChange(t.value)}
                    >
                        {t.label}
                        <span className="ml-2 tabular-nums text-xs opacity-60">({t.count})</span>
                    </Button>
                ))}
            </div>

            {/* ═══ Task Table ═════════════════════════════ */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                        <div>
                            <CardTitle>Task List</CardTitle>
                            <CardDescription>
                                {filteredTasks.length} task{filteredTasks.length !== 1 ? 's' : ''} found
                            </CardDescription>
                        </div>
                        <div className="flex flex-wrap items-center gap-2">
                            {selectedRows.length > 0 && (
                                <div className="flex items-center gap-2">
                                    <span className="text-sm text-muted-foreground">
                                        {selectedRows.length} selected
                                    </span>
                                    <Separator orientation="vertical" className="h-6"/>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={handleBatchRetry}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-1"/>
                                        Retry All
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => setSelectedRows([])}
                                    >
                                        <XCircle className="h-4 w-4 mr-1"/>
                                        Clear
                                    </Button>
                                </div>
                            )}
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button variant="outline" size="sm">
                                        <Settings className="h-4 w-4 mr-2"/>
                                        Options
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuLabel>Batch Actions</DropdownMenuLabel>
                                    <DropdownMenuSeparator/>
                                    <DropdownMenuItem>
                                        <Download className="h-4 w-4 mr-2"/>
                                        Export Tasks
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="px-0">
                    {!filteredTasks?.length ? (
                        <div className="py-16 text-center space-y-4">
                            <Film className="w-16 h-16 mx-auto mb-4 opacity-15"/>
                            <h3 className="text-lg font-medium">No transcoding tasks found</h3>
                            <p className="text-sm text-muted-foreground max-w-md mx-auto">
                                {urlMediaId
                                    ? `No transcoding tasks found for Media #${urlMediaId}. Try uploading a new media file to start transcoding.`
                                    : searchQuery || profileFilter !== ''
                                        ? "No tasks match your search criteria. Try adjusting your filters."
                                        : "No transcoding tasks found. Upload media files to generate encoding tasks."}
                            </p>
                            {!urlMediaId && !searchQuery && profileFilter === '' && (
                                <Button variant="default" size="sm" className="mt-2"
                                        onClick={() => navigate('/admin/media')}>
                                    <Video className="w-4 h-4 mr-2"/>
                                    Go to Media Library
                                </Button>
                            )}
                        </div>
                    ) : (
                        <div className="overflow-x-auto">
                            <Table>
                                <TableHeader>
                                    <TableRow className="border-b hover:bg-transparent">
                                        <TableHead className="w-[50px]">
                                            <Checkbox
                                                checked={filteredTasks.length > 0 && selectedRows.length === filteredTasks.length}
                                                onCheckedChange={toggleSelectAll}
                                                aria-label="Select all"
                                            />
                                        </TableHead>
                                        <TableHead
                                            className="max-w-[320px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">MEDIA</TableHead>
                                        <TableHead
                                            className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
                                            onClick={() => handleSort('profile_id')}
                                        >
                                            <div className="flex items-center gap-1">
                                                PROFILE
                                                <ArrowUpDown className="h-3 w-3"/>
                                            </div>
                                        </TableHead>
                                        <TableHead
                                            className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground"
                                        >
                                            <div className="flex items-center gap-1">
                                                PROGRESS
                                            </div>
                                        </TableHead>
                                        <TableHead
                                            className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
                                            onClick={() => handleSort('status')}
                                        >
                                            <div className="flex items-center gap-1">
                                                STATUS
                                                <ArrowUpDown className="h-3 w-3"/>
                                            </div>
                                        </TableHead>
                                        <TableHead
                                            className="w-[80px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">OUTPUT</TableHead>
                                        <TableHead
                                            className="w-[120px] py-2.5 text-right text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
                                            onClick={() => handleSort('update_time' as any)}
                                        >
                                            <div className="flex items-center justify-end gap-1">
                                                TIME
                                                <ArrowUpDown className="h-3 w-3"/>
                                            </div>
                                        </TableHead>
                                        <TableHead
                                            className="w-[180px] py-2.5 text-right text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                                            ACTION
                                        </TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {filteredTasks.map((task) => (
                                        <TaskRow key={task.id}
                                                 task={task}
                                                 onRetry={() => handleRetryTask(String(task.id))}
                                                 isRetrying={retryingTaskId === task.id}
                                                 isSelected={selectedRows.includes(task.id)}
                                                 onToggleSelect={() => toggleSelectRow(task.id)}
                                        />
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Pagination */}
            {filteredData && (
                <TablePagination
                    page={page}
                    pageSize={filteredData.page_size ?? 25}
                    total={filteredData.total}
                    onPageChange={setPage}
                />
            )}
        </div>
    );
}
