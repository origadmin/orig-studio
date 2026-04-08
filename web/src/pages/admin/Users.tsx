/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 用户管理页面
 */

import {useState, useEffect} from 'react';
import {Search, Plus, User, MoreVertical, Trash2, Edit, Shield, Mail, Eye} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
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
import {userApi} from '@/lib/api/user';
import {useTranslation} from 'react-i18next';
import {getFullUrl} from '@/lib/utils';

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
                setUsers(response.list || []);
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
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">{t('admin.users')}</h2>
                    <p className="text-slate-500 text-sm mt-1">{t('admin.manageUsers') || "Manage users, roles, and permissions"}</p>
                </div>
                <Button className="bg-blue-600 hover:bg-blue-700">
                    <Plus className="w-4 h-4 mr-2"/>
                    {t('admin.addUser') || "Add User"}
                </Button>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.totalUsers') || "Total Users"}</p>
                                <p className="text-2xl font-bold">{users.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <User className="w-6 h-6 text-blue-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.activeUsers') || "Active Users"}</p>
                                <p className="text-2xl font-bold">{users.filter(u => u.status === 'active').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6 text-green-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.admins') || "Admins"}</p>
                                <p className="text-2xl font-bold">{users.filter(u => u.role === 'admin').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.editors') || "Editors"}</p>
                                <p className="text-2xl font-bold">{users.filter(u => u.role === 'editor').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
                                <Edit className="w-6 h-6 text-orange-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"/>
                    <Input
                        placeholder={t('admin.search') || "Search users..."}
                        className="pl-10"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <select
                    className="h-10 px-3 rounded-md border border-input bg-background text-sm"
                    value={roleFilter}
                    onChange={(e) => setRoleFilter(e.target.value)}
                >
                    <option value="all">{t('admin.allRoles') || "All Roles"}</option>
                    <option value="admin">{t('admin.admin') || "Admin"}</option>
                    <option value="editor">{t('admin.editor') || "Editor"}</option>
                    <option value="user">{t('admin.user') || "User"}</option>
                </select>
            </div>

            {/* Users Table */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('admin.allUsers') || "All Users"}</CardTitle>
                    <CardDescription>{t('admin.manageUserAccounts') || "Manage user accounts and permissions"}</CardDescription>
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