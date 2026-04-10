/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState, useMemo} from "react";
import {mediaApi, encodingApi, type EncodeProfile} from "../../lib/api/media";
import {Button} from "../../components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "../../components/ui/table";
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from "../../components/ui/card";
import {Badge} from "../../components/ui/badge";
import {Input} from "../../components/ui/input";
import {Label} from "../../components/ui/label";
import {Checkbox} from "../../components/ui/checkbox";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogFooter,
    DialogTrigger
} from "../../components/ui/dialog";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../../components/ui/select";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "../../components/ui/dropdown-menu";
import {
    PlusCircle, Edit, Trash2, CheckCircle, XCircle, Play, Pause,
    Search, Filter, MoreVertical, Download, Upload, Copy,
    Trash, CheckSquare, Square, ArrowUpDown, ChevronDown, ChevronUp
} from "lucide-react";
import {Separator} from "../../components/ui/separator";

export default function TranscodingProfiles() {
    const [profiles, setProfiles] = useState<EncodeProfile[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingProfile, setEditingProfile] = useState<Partial<EncodeProfile> | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    // 新增的状态
    const [searchQuery, setSearchQuery] = useState('');
    const [statusFilter, setStatusFilter] = useState<string>('');
    const [codecFilter, setCodecFilter] = useState<string>('');
    const [extensionFilter, setExtensionFilter] = useState<string>('');
    const [resolutionFilter, setResolutionFilter] = useState<string>('');
    const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);
    const [selectedRows, setSelectedRows] = useState<number[]>([]);
    const [sortConfig, setSortConfig] = useState<{ key: keyof EncodeProfile, direction: 'asc' | 'desc' } | null>(null);

    const fetchProfiles = async () => {
        try {
            setLoading(true);
            const response = await encodingApi.profiles.list();
            setProfiles(response.profiles || []);
        } catch (error) {
            console.error("Failed to fetch profiles:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchProfiles();
    }, []);

    // 筛选和排序逻辑
    const filteredProfiles = useMemo(() => {
        let result = [...profiles];

        // 搜索筛选
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            result = result.filter(p =>
                p.name.toLowerCase().includes(query) ||
                p.resolution.toLowerCase().includes(query) ||
                p.video_codec.toLowerCase().includes(query)
            );
        }

        // 状态筛选
        if (statusFilter && statusFilter !== '-') {
            result = result.filter(p =>
                statusFilter === 'active' ? p.is_active : !p.is_active
            );
        }

        // 编码筛选
        if (codecFilter && codecFilter !== '-') {
            result = result.filter(p => p.video_codec.includes(codecFilter));
        }

        // 扩展名筛选
        if (extensionFilter && extensionFilter !== '-') {
            result = result.filter(p => p.extension === extensionFilter);
        }

        // 分辨率筛选（只匹配高度部分）
        if (resolutionFilter && resolutionFilter !== '-') {
            result = result.filter(p => {
                const height = p.resolution.split('x')[1] || p.resolution;
                return height === resolutionFilter;
            });
        }

        // 排序
        if (sortConfig) {
            result.sort((a, b) => {
                const aVal = a[sortConfig.key];
                const bVal = b[sortConfig.key];

                if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
                if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
                return 0;
            });
        } else {
            // 默认按 ID 升序排序，保持稳定的顺序
            result.sort((a, b) => a.id - b.id);
        }

        return result;
    }, [profiles, searchQuery, statusFilter, codecFilter, extensionFilter, resolutionFilter, sortConfig]);

    // 获取所有可用的编码
    const availableCodecs = useMemo(() => {
        const codecs = new Set<string>();
        profiles.forEach(p => {
            const codec = p.video_codec.toLowerCase();
            if (codec.includes('h264')) codecs.add('h264');
            if (codec.includes('h265') || codec.includes('hevc')) codecs.add('h265');
            if (codec.includes('vp9')) codecs.add('vp9');
        });
        return Array.from(codecs);
    }, [profiles]);

    // 获取所有可用的扩展名
    const availableExtensions = useMemo(() => {
        const extensions = new Set<string>();
        profiles.forEach(p => extensions.add(p.extension));
        return Array.from(extensions);
    }, [profiles]);

    // 获取所有可用的分辨率（只取高度部分）
    const availableResolutions = useMemo(() => {
        const resolutions = new Set<string>();
        profiles.forEach(p => {
            const height = p.resolution.split('x')[1] || p.resolution;
            if (height && height !== '-') {
                resolutions.add(height);
            }
        });
        return Array.from(resolutions).sort((a, b) => parseInt(a) - parseInt(b));
    }, [profiles]);

    const handleSort = (key: keyof EncodeProfile) => {
        setSortConfig(current => {
            if (current?.key === key) {
                return {key, direction: current.direction === 'asc' ? 'desc' : 'asc'};
            }
            return {key, direction: 'asc'};
        });
    };

    // 批量选择
    const toggleSelectAll = () => {
        if (selectedRows.length === filteredProfiles.length) {
            setSelectedRows([]);
        } else {
            setSelectedRows(filteredProfiles.map(p => p.id));
        }
    };

    const toggleSelectRow = (id: number) => {
        setSelectedRows(prev =>
            prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
        );
    };

    // 批量操作
    const handleBatchActivate = async () => {
        try {
            // 乐观更新：先更新本地状态
            setProfiles(prev => prev.map(p =>
                selectedRows.includes(p.id) ? {...p, is_active: true} : p
            ));
            // 然后发送 API 请求
            for (const id of selectedRows) {
                const profile = profiles.find(p => p.id === id);
                if (profile) {
                    await encodingApi.profiles.update(profile.id, {...profile, is_active: true});
                }
            }
            setSelectedRows([]);
        } catch (error) {
            console.error("Failed to batch activate profiles:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };

    const handleBatchDeactivate = async () => {
        try {
            // 乐观更新：先更新本地状态
            setProfiles(prev => prev.map(p =>
                selectedRows.includes(p.id) ? {...p, is_active: false} : p
            ));
            // 然后发送 API 请求
            for (const id of selectedRows) {
                const profile = profiles.find(p => p.id === id);
                if (profile) {
                    await encodingApi.profiles.update(profile.id, {...profile, is_active: false});
                }
            }
            setSelectedRows([]);
        } catch (error) {
            console.error("Failed to batch deactivate profiles:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };

    const handleBatchDelete = async () => {
        if (!confirm(`Are you sure you want to delete ${selectedRows.length} profiles?`)) return;
        try {
            // 乐观更新：先更新本地状态
            setProfiles(prev => prev.filter(p => !selectedRows.includes(p.id)));
            // 然后发送 API 请求
            for (const id of selectedRows) {
                await encodingApi.profiles.delete(id);
            }
            setSelectedRows([]);
        } catch (error) {
            console.error("Failed to batch delete profiles:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };

    const handleSave = async () => {
        if (!editingProfile) return;

        // 验证必选字段
        if (!editingProfile.name || !editingProfile.extension) {
            alert("Name and Extension are required fields");
            return;
        }
        
        try {
            if (editingProfile.id) {
                // 更新现有配置
                setProfiles(prev => prev.map(p =>
                    p.id === editingProfile.id ? {...p, ...editingProfile} as EncodeProfile : p
                ));
                await encodingApi.profiles.update(editingProfile.id, editingProfile);
            } else {
                // 创建新配置 - 这里需要等待返回的ID，所以不做乐观更新
                await encodingApi.profiles.create(editingProfile);
                fetchProfiles();
            }
            setIsDialogOpen(false);
        } catch (error) {
            console.error("Failed to save profile:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Are you sure you want to delete this profile?")) return;
        try {
            // 乐观更新：先更新本地状态
            setProfiles(prev => prev.filter(p => p.id !== id));
            // 然后发送 API 请求
            await encodingApi.profiles.delete(id);
        } catch (error) {
            console.error("Failed to delete profile:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };

    const handleToggleActive = async (profile: EncodeProfile) => {
        try {
            // 乐观更新：先更新本地状态
            setProfiles(prev => prev.map(p =>
                p.id === profile.id ? {...p, is_active: !p.is_active} : p
            ));
            // 然后发送 API 请求
            await encodingApi.profiles.update(profile.id, {...profile, is_active: !profile.is_active});
        } catch (error) {
            console.error("Failed to toggle profile status:", error);
            // 出错时回滚
            fetchProfiles();
        }
    };


    return (
        <div className="space-y-6">
            {/* 页面标题和操作区域 */}
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Encoding Profiles</h2>
                    <p className="text-sm text-muted-foreground mt-1">
                        Manage and configure your video encoding presets
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    {selectedRows.length > 0 && (
                        <div className="flex items-center gap-2">
                            <span className="text-sm text-muted-foreground">
                                {selectedRows.length} selected
                            </span>
                            <Separator orientation="vertical" className="h-6"/>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleBatchActivate}
                            >
                                <Play className="h-4 w-4 mr-1"/>
                                Activate
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleBatchDeactivate}
                            >
                                <Pause className="h-4 w-4 mr-1"/>
                                Deactivate
                            </Button>
                            <Button
                                variant="destructive"
                                size="sm"
                                onClick={handleBatchDelete}
                            >
                                <Trash className="h-4 w-4 mr-1"/>
                                Delete
                            </Button>
                            <Separator orientation="vertical" className="h-6"/>
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setSelectedRows([])}
                            >
                                Clear
                            </Button>
                        </div>
                    )}
                    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                        <DialogTrigger asChild>
                            <Button onClick={() => setEditingProfile({is_active: true})}>
                                <PlusCircle className="mr-2 h-4 w-4"/>
                                Add Profile
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="max-w-lg">
                            <DialogHeader>
                                <DialogTitle>{editingProfile?.id ? "Edit Profile" : "Add Profile"}</DialogTitle>
                                <DialogDescription>
                                    {editingProfile?.id ? "Update the profile settings" : "Create a new transcoding profile"}
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="name" className="text-right">Name <span
                                        className="text-destructive">*</span></Label>
                                    <Input id="name" value={editingProfile?.name || ""}
                                           onChange={(e) => setEditingProfile({
                                               ...editingProfile,
                                               name: e.target.value
                                           })} className="col-span-3"/>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="extension" className="text-right">Extension <span
                                        className="text-destructive">*</span></Label>
                                    <Select
                                        value={editingProfile?.extension || ""}
                                        onValueChange={(value) => setEditingProfile({
                                            ...editingProfile,
                                            extension: value
                                        })}
                                    >
                                        <SelectTrigger className="col-span-3">
                                            <SelectValue placeholder="Select extension"/>
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="mp4">MP4</SelectItem>
                                            <SelectItem value="webm">WebM</SelectItem>
                                            <SelectItem value="mkv">Matroska</SelectItem>
                                            <SelectItem value="mov">QuickTime</SelectItem>
                                            <SelectItem value="gif">GIF</SelectItem>
                                            <SelectItem value="avi">AVI</SelectItem>
                                            <SelectItem value="flv">FLV</SelectItem>
                                            <SelectItem value="wmv">WMV</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="res" className="text-right">Resolution</Label>
                                    <Select
                                        value={editingProfile?.resolution?.split('x')[1] || editingProfile?.resolution || ""}
                                        onValueChange={(value) => {
                                            // 根据高度自动选择完整分辨率
                                            const resolutionMap: Record<string, string> = {
                                                '240': '426x240',
                                                '360': '640x360',
                                                '480': '854x480',
                                                '720': '1280x720',
                                                '1080': '1920x1080',
                                                '1440': '2560x1440',
                                                '2160': '3840x2160'
                                            };
                                            const fullResolution = resolutionMap[value] || value;
                                            setEditingProfile({...editingProfile, resolution: fullResolution});
                                        }}
                                    >
                                        <SelectTrigger className="col-span-3">
                                            <SelectValue placeholder="Select resolution"/>
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="240">240</SelectItem>
                                            <SelectItem value="360">360</SelectItem>
                                            <SelectItem value="480">480</SelectItem>
                                            <SelectItem value="720">720</SelectItem>
                                            <SelectItem value="1080">1080</SelectItem>
                                            <SelectItem value="1440">1440</SelectItem>
                                            <SelectItem value="2160">2160 (4K)</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="vcodec" className="text-right">Video Codec</Label>
                                    <Select
                                        value={editingProfile?.video_codec || ""}
                                        onValueChange={(value) => setEditingProfile({
                                            ...editingProfile,
                                            video_codec: value
                                        })}
                                    >
                                        <SelectTrigger className="col-span-3">
                                            <SelectValue placeholder="Select video codec"/>
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="libx264">libx264 (H.264)</SelectItem>
                                            <SelectItem value="libx265">libx265 (H.265/HEVC)</SelectItem>
                                            <SelectItem value="libvpx-vp9">libvpx-vp9 (VP9)</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <div className="col-span-4">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => setShowAdvancedOptions(!showAdvancedOptions)}
                                            className="w-full flex items-center justify-center gap-1 text-sm text-muted-foreground"
                                        >
                                            {showAdvancedOptions ? <ChevronUp className="h-4 w-4"/> :
                                                <ChevronDown className="h-4 w-4"/>}
                                            {showAdvancedOptions ? "Hide Advanced Options" : "Show Advanced Options"}
                                        </Button>
                                    </div>
                                </div>
                                {showAdvancedOptions && (
                                    <>
                                        <div className="grid grid-cols-4 items-center gap-4">
                                            <Label htmlFor="vbitrate" className="text-right">Video Bitrate</Label>
                                            <Select
                                                value={editingProfile?.video_bitrate || ""}
                                                onValueChange={(value) => setEditingProfile({
                                                    ...editingProfile,
                                                    video_bitrate: value
                                                })}
                                            >
                                                <SelectTrigger className="col-span-3">
                                                    <SelectValue placeholder="Select video bitrate"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="200k">200k</SelectItem>
                                                    <SelectItem value="400k">400k</SelectItem>
                                                    <SelectItem value="800k">800k</SelectItem>
                                                    <SelectItem value="1500k">1500k</SelectItem>
                                                    <SelectItem value="3000k">3000k</SelectItem>
                                                    <SelectItem value="5000k">5000k</SelectItem>
                                                    <SelectItem value="8000k">8000k</SelectItem>
                                                    <SelectItem value="12000k">12000k</SelectItem>
                                                    <SelectItem value="16000k">16000k</SelectItem>
                                                    <SelectItem value="25000k">25000k</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <div className="grid grid-cols-4 items-center gap-4">
                                            <Label htmlFor="acodec" className="text-right">Audio Codec</Label>
                                            <Select
                                                value={editingProfile?.audio_codec || ""}
                                                onValueChange={(value) => setEditingProfile({
                                                    ...editingProfile,
                                                    audio_codec: value
                                                })}
                                            >
                                                <SelectTrigger className="col-span-3">
                                                    <SelectValue placeholder="Select audio codec"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="aac">aac</SelectItem>
                                                    <SelectItem value="mp3">mp3</SelectItem>
                                                    <SelectItem value="opus">opus</SelectItem>
                                                    <SelectItem value="vorbis">vorbis</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <div className="grid grid-cols-4 items-center gap-4">
                                            <Label htmlFor="abitrate" className="text-right">Audio Bitrate</Label>
                                            <Select
                                                value={editingProfile?.audio_bitrate || ""}
                                                onValueChange={(value) => setEditingProfile({
                                                    ...editingProfile,
                                                    audio_bitrate: value
                                                })}
                                            >
                                                <SelectTrigger className="col-span-3">
                                                    <SelectValue placeholder="Select audio bitrate"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="64k">64k</SelectItem>
                                                    <SelectItem value="96k">96k</SelectItem>
                                                    <SelectItem value="128k">128k</SelectItem>
                                                    <SelectItem value="192k">192k</SelectItem>
                                                    <SelectItem value="256k">256k</SelectItem>
                                                    <SelectItem value="320k">320k</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </>
                                )}
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="active" className="text-right">Active</Label>
                                    <div className="col-span-3 flex items-center">
                                        <Checkbox
                                            id="active"
                                            checked={editingProfile?.is_active ?? true}
                                            onCheckedChange={(checked) => setEditingProfile({
                                                ...editingProfile,
                                                is_active: checked
                                            })}
                                        />
                                        <label htmlFor="active" className="ml-2 text-sm font-medium leading-none">
                                            Enable this profile
                                        </label>
                                    </div>
                                </div>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                                <Button onClick={handleSave}>Save</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>
            {/* 搜索和筛选区域 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex flex-col lg:flex-row gap-3">
                        <div className="flex-1 lg:max-w-xs">
                            <div className="relative">
                                <Search
                                    className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                <Input
                                    placeholder="Search..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10 h-9"
                                />
                            </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <Select value={extensionFilter} onValueChange={setExtensionFilter}>
                                <SelectTrigger className="w-[130px] h-9">
                                    <div className="flex items-center gap-2">
                                        <Filter className="h-4 w-4"/>
                                        <SelectValue placeholder="Extension"/>
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="-">-</SelectItem>
                                    {availableExtensions.map(ext => {
                                        const getExtensionDisplayName = (e: string) => {
                                            switch (e.toLowerCase()) {
                                                case 'mp4':
                                                    return 'MP4';
                                                case 'webm':
                                                    return 'WebM';
                                                case 'mkv':
                                                    return 'Matroska';
                                                case 'mov':
                                                    return 'QuickTime';
                                                case 'gif':
                                                    return 'GIF';
                                                case 'avi':
                                                    return 'AVI';
                                                case 'flv':
                                                    return 'FLV';
                                                case 'wmv':
                                                    return 'WMV';
                                                default:
                                                    return e.toUpperCase();
                                            }
                                        };
                                        return (
                                            <SelectItem key={ext} value={ext}>
                                                {getExtensionDisplayName(ext)}
                                            </SelectItem>
                                        );
                                    })}
                                </SelectContent>
                            </Select>
                            <Select value={resolutionFilter} onValueChange={setResolutionFilter}>
                                <SelectTrigger className="w-[140px] h-9">
                                    <div className="flex items-center gap-2">
                                        <Filter className="h-4 w-4"/>
                                        <SelectValue placeholder="Resolution"/>
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="-">-</SelectItem>
                                    {availableResolutions.map(res => (
                                        <SelectItem key={res} value={res}>
                                            {res}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <Select value={codecFilter} onValueChange={setCodecFilter}>
                                <SelectTrigger className="w-[120px] h-9">
                                    <div className="flex items-center gap-2">
                                        <Filter className="h-4 w-4"/>
                                        <SelectValue placeholder="Codec"/>
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="-">-</SelectItem>
                                    {availableCodecs.map(codec => (
                                        <SelectItem key={codec} value={codec}>
                                            {codec.toUpperCase()}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger className="w-[120px] h-9">
                                    <div className="flex items-center gap-2">
                                        <Filter className="h-4 w-4"/>
                                        <SelectValue placeholder="Status"/>
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="-">-</SelectItem>
                                    <SelectItem value="active">Active</SelectItem>
                                    <SelectItem value="inactive">Inactive</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 表格区域 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>Profile Management</CardTitle>
                            <CardDescription>
                                {filteredProfiles.length} profile{filteredProfiles.length !== 1 ? 's' : ''} found
                            </CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button variant="outline" size="sm">
                                        <MoreVertical className="h-4 w-4"/>
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                    <DropdownMenuSeparator/>
                                    <DropdownMenuItem>
                                        <Download className="h-4 w-4 mr-2"/>
                                        Export Profiles
                                    </DropdownMenuItem>
                                    <DropdownMenuItem>
                                        <Upload className="h-4 w-4 mr-2"/>
                                        Import Profiles
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="px-0">
                    <div className="overflow-x-auto">
                        <Table>
                            <TableHeader>
                                <TableRow className="bg-muted/50">
                                    <TableHead className="w-[50px]">
                                        <Checkbox
                                            checked={filteredProfiles.length > 0 && selectedRows.length === filteredProfiles.length}
                                            onCheckedChange={toggleSelectAll}
                                            aria-label="Select all"
                                        />
                                    </TableHead>
                                    <TableHead
                                        className="cursor-pointer hover:bg-muted/80"
                                        onClick={() => handleSort('name')}
                                    >
                                        <div className="flex items-center gap-1">
                                            Name
                                            <ArrowUpDown className="h-3 w-3"/>
                                        </div>
                                    </TableHead>
                                    <TableHead
                                        className="cursor-pointer hover:bg-muted/80"
                                        onClick={() => handleSort('extension')}
                                    >
                                        <div className="flex items-center gap-1">
                                            Extension
                                            <ArrowUpDown className="h-3 w-3"/>
                                        </div>
                                    </TableHead>
                                    <TableHead
                                        className="cursor-pointer hover:bg-muted/80"
                                        onClick={() => handleSort('resolution')}
                                    >
                                        <div className="flex items-center gap-1">
                                            Resolution
                                            <ArrowUpDown className="h-3 w-3"/>
                                        </div>
                                    </TableHead>
                                    <TableHead
                                        className="cursor-pointer hover:bg-muted/80"
                                        onClick={() => handleSort('video_codec')}
                                    >
                                        <div className="flex items-center gap-1">
                                            Video Codec
                                            <ArrowUpDown className="h-3 w-3"/>
                                        </div>
                                    </TableHead>
                                    <TableHead
                                        className="cursor-pointer hover:bg-muted/80"
                                        onClick={() => handleSort('is_active')}
                                    >
                                        <div className="flex items-center gap-1">
                                            Status
                                            <ArrowUpDown className="h-3 w-3"/>
                                        </div>
                                    </TableHead>
                                    <TableHead className="text-right w-[150px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8">
                                            Loading profiles...
                                        </TableCell>
                                    </TableRow>
                                ) : filteredProfiles.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8">
                                            <div
                                                className="flex flex-col items-center justify-center text-muted-foreground">
                                                <Search className="h-12 w-12 mb-4 opacity-20"/>
                                                <p className="text-lg font-medium">No profiles found</p>
                                                <p className="text-sm">Try adjusting your search or filters</p>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredProfiles.map((p) => {
                                        // 显示完整的扩展名名称
                                        const getExtensionDisplayName = (ext: string) => {
                                            switch (ext.toLowerCase()) {
                                                case 'mp4':
                                                    return 'MP4';
                                                case 'webm':
                                                    return 'WebM';
                                                case 'mkv':
                                                    return 'Matroska';
                                                case 'mov':
                                                    return 'QuickTime';
                                                case 'gif':
                                                    return 'GIF';
                                                case 'avi':
                                                    return 'AVI';
                                                case 'flv':
                                                    return 'FLV';
                                                case 'wmv':
                                                    return 'WMV';
                                                default:
                                                    return ext.toUpperCase();
                                            }
                                        };
                                        return (
                                            <TableRow
                                                key={p.id}
                                                className={selectedRows.includes(p.id) ? 'bg-muted/30' : ''}
                                            >
                                                <TableCell>
                                                    <Checkbox
                                                        checked={selectedRows.includes(p.id)}
                                                        onCheckedChange={() => toggleSelectRow(p.id)}
                                                        aria-label={`Select ${p.name}`}
                                                    />
                                                </TableCell>
                                                <TableCell className="font-medium">{p.name}</TableCell>
                                                <TableCell>
                                                    <Badge variant="outline">
                                                        {getExtensionDisplayName(p.extension)}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell>
                                                    <Badge variant="outline" className="font-mono">
                                                        {p.resolution.split('x')[1] || p.resolution}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell>
                                                    <div className="flex flex-col">
                                                        <span>{p.video_codec}</span>
                                                        <span className="text-xs text-muted-foreground">
                                                            {p.video_bitrate}
                                                        </span>
                                                    </div>
                                                </TableCell>
                                                <TableCell>
                                                    <Badge
                                                        variant={p.is_active ? "default" : "secondary"}
                                                        className="flex items-center gap-1 w-fit"
                                                    >
                                                        {p.is_active ? (
                                                            <CheckCircle className="h-3 w-3"/>
                                                        ) : (
                                                            <XCircle className="h-3 w-3"/>
                                                        )}
                                                        {p.is_active ? "Active" : "Inactive"}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <div className="flex items-center justify-end gap-1">
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            onClick={() => handleToggleActive(p)}
                                                            title={p.is_active ? "Deactivate" : "Activate"}
                                                            className="h-8 w-8"
                                                        >
                                                            {p.is_active ? <Pause className="h-4 w-4"/> :
                                                                <Play className="h-4 w-4"/>}
                                                        </Button>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            onClick={() => {
                                                                setEditingProfile(p);
                                                                setIsDialogOpen(true);
                                                            }}
                                                            className="h-8 w-8"
                                                        >
                                                            <Edit className="h-4 w-4"/>
                                                        </Button>
                                                        <DropdownMenu>
                                                            <DropdownMenuTrigger asChild>
                                                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                                                    <MoreVertical className="h-4 w-4"/>
                                                                </Button>
                                                            </DropdownMenuTrigger>
                                                            <DropdownMenuContent align="end">
                                                                <DropdownMenuItem onClick={() => {
                                                                    setEditingProfile({...p});
                                                                    setIsDialogOpen(true);
                                                                }}>
                                                                    <Copy className="h-4 w-4 mr-2"/>
                                                                    Duplicate
                                                                </DropdownMenuItem>
                                                                <DropdownMenuSeparator/>
                                                                <DropdownMenuItem
                                                                    className="text-destructive focus:text-destructive"
                                                                    onClick={() => handleDelete(p.id)}
                                                                >
                                                                    <Trash2 className="h-4 w-4 mr-2"/>
                                                                    Delete
                                                                </DropdownMenuItem>
                                                            </DropdownMenuContent>
                                                        </DropdownMenu>
                                                    </div>
                                                </TableCell>
                                            </TableRow>
                                        );
                                    })
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
