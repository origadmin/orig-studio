/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 用户管理页面
 */

import {useState, useEffect} from 'react';
import {Search, Plus, User, MoreVertical, Trash2, Edit, Shield, Mail, Eye, RotateCcw} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Filter} from 'lucide-react';
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
import {userApi} from '@/lib/api/user';
import {useTranslation} from 'react-i18next';
import {getFullUrl} from '@/lib/utils';
import {extractList} from '@/lib/extract';

export default function UsersPage() {
    const {t} = useTranslation();
    const [users, setUsers] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [roleFilter, setRoleFilter] = useState('all');

    useEffect(() => {
        const fetchUsers = async () => {
            try {
                setLoading(true);
                const response = await userApi.list({page_size: 100});
                // 提取列表数据，防止因格式不匹配导致崩溃
                const userList = extractList<any>(response);
                setUsers(userList);
            } catch (error) {
                console.error('Failed to fetch users:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchUsers();
    }, []);

    const filteredUsers = users.filter(user => {
        const matchesSearch =
            user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
            (user.nickname && user.nickname.toLowerCase().includes(searchTerm.toLowerCase())) ||
            user.email.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesRole = roleFilter === 'all' || user.role === roleFilter;
        return matchesSearch && matchesRole;
    });

    const getRoleBadge = (role: string) => {
        const roles: Record<string, { variant: "default" | "secondary" | "outline", label: string }> = {
            admin: {variant: "default", label: t('admin.admin') || "Admin"},
            editor: {variant: "secondary", label: t('admin.editor') || "Editor"},
            user: {variant: "outline", label: t('admin.user') || "User"}
        };
        return roles[role] || {variant: "outline", label: role};
    };

    const getStatusBadge = (status: string) => {
        return status === "active"
            ? <Badge variant="default" className="bg-green-500">{t('admin.active') || "Active"}</Badge>
            : <Badge variant="secondary">{t('admin.inactive') || "Inactive"}</Badge>;
    };



    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.users')}</h2>
                                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5">
                                    {t('admin.manageUsers') || "Manage users, roles, and permissions"}
                                </p>
                            </div>
                        </div>

                        {/* 分隔线 */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        {/* 搜索和筛选 */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-[120px] max-w-[400px]">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder={t('admin.search') || "Search users..."}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={roleFilter} onValueChange={setRoleFilter}>
                                    <SelectTrigger className="w-[140px] h-9 focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {roleFilter === 'all' ? (
                                                <span className="text-muted-foreground">Roles</span>
                                            ) : (
                                                <SelectValue placeholder="Roles"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="admin">{t('admin.admin') || "Admin"}</SelectItem>
                                        <SelectItem value="editor">{t('admin.editor') || "Editor"}</SelectItem>
                                        <SelectItem value="user">{t('admin.user') || "User"}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-9 px-3"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setRoleFilter('all');
                                        }}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2"/>
                                        Reset
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        className="h-9 px-4"
                                        onClick={() => {
                                            // 这里可以添加搜索逻辑
                                        }}
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

            {/* Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.totalUsers') || "Total Users"}</p>
                                <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">{users.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <User className="w-6 h-6 text-blue-600"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-blue-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.activeUsers') || "Active Users"}</p>
                                <p className="text-2xl font-bold text-green-600 dark:text-green-400">{users.filter(u => u.status === 'active').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6 text-green-600"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-green-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.admins') || "Admins"}</p>
                                <p className="text-2xl font-bold text-purple-600 dark:text-purple-400">{users.filter(u => u.role === 'admin').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-purple-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.editors') || "Editors"}</p>
                                <p className="text-2xl font-bold text-orange-600 dark:text-orange-400">{users.filter(u => u.role === 'editor').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
                                <Edit className="w-6 h-6 text-orange-600"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-orange-500 w-full opacity-10"/>
                </Card>
            </div>

            {/* Users Table */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.allUsers') || "All Users"}</CardTitle>
                            <CardDescription>{t('admin.manageUserAccounts') || "Manage user accounts and permissions"}</CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button className="bg-blue-600 hover:bg-blue-700">
                                <Plus className="w-4 h-4 mr-2"/>
                                {t('admin.addUser') || "Add User"}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center">
                            <div
                                className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full mx-auto"/>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>{t('admin.user') || "User"}</TableHead>
                                    <TableHead>{t('admin.email') || "Email"}</TableHead>
                                    <TableHead>{t('admin.role') || "Role"}</TableHead>
                                    <TableHead>{t('admin.status') || "Status"}</TableHead>
                                    <TableHead>{t('admin.joined') || "Joined"}</TableHead>
                                    <TableHead className="text-right">{t('admin.actions') || "Actions"}</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredUsers.length > 0 ? filteredUsers.map((user) => (
                                    <TableRow key={user.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <Avatar className="w-10 h-10">
                                                    <AvatarImage
                                                        src={user.avatar ? getFullUrl(user.avatar) : undefined}/>
                                                    <AvatarFallback>{user.nickname ? user.nickname.charAt(0) : user.username.charAt(0)}</AvatarFallback>
                                                </Avatar>
                                                <div>
                                                    <p className="font-medium">{user.nickname || user.username}</p>
                                                    <p className="text-sm text-slate-500">@{user.username}</p>
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-sm text-slate-500">
                                            <div className="flex items-center gap-2">
                                                <Mail className="w-4 h-4"/>
                                                {user.email}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge {...getRoleBadge(user.role)} />
                                        </TableCell>
                                        <TableCell>
                                            {getStatusBadge(user.status)}
                                        </TableCell>
                                        <TableCell className="text-sm text-slate-500">{user.created_at}</TableCell>
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
                                                        {t('admin.viewProfile') || "View Profile"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem>
                                                        <Edit className="w-4 h-4 mr-2"/>
                                                        {t('admin.edit') || "Edit"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem>
                                                        <Shield className="w-4 h-4 mr-2"/>
                                                        {t('admin.changeRole') || "Change Role"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem className="text-red-600">
                                                        <Trash2 className="w-4 h-4 mr-2"/>
                                                        {t('admin.delete') || "Delete"}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow>
                                        <TableCell colSpan={6} className="text-center py-8">
                                            {t('admin.noUsersFound') || "No users found"}
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}