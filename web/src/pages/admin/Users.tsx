/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 用户管理页面
 */

import {useState} from 'react';
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

// 模拟用户数据
const mockUsers = [
    {
        id: "1",
        username: "gopher_expert",
        nickname: "Gopher Expert",
        email: "gopher@example.com",
        avatar: "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=100",
        role: "user",
        status: "active",
        created_at: "2023-06-15"
    },
    {
        id: "2",
        username: "react_master",
        nickname: "React Master",
        email: "react@example.com",
        avatar: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=100",
        role: "admin",
        status: "active",
        created_at: "2023-07-20"
    },
    {
        id: "3",
        username: "devops_pro",
        nickname: "DevOps Pro",
        email: "devops@example.com",
        avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&q=80&w=100",
        role: "user",
        status: "active",
        created_at: "2023-08-10"
    },
    {
        id: "4",
        username: "typescript_guru",
        nickname: "TypeScript Guru",
        email: "typescript@example.com",
        avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=100",
        role: "editor",
        status: "active",
        created_at: "2023-09-05"
    },
    {
        id: "5",
        username: "cloud_expert",
        nickname: "Cloud Expert",
        email: "cloud@example.com",
        avatar: "https://images.unsplash.com/photo-1560250097-0b93528c311a?auto=format&fit=crop&q=80&w=100",
        role: "user",
        status: "inactive",
        created_at: "2023-10-12"
    }
];

export default function UsersPage() {
    const [users] = useState(mockUsers);
    const [searchTerm, setSearchTerm] = useState('');
    const [roleFilter, setRoleFilter] = useState('all');

    const filteredUsers = users.filter(user => {
        const matchesSearch =
            user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
            user.nickname.toLowerCase().includes(searchTerm.toLowerCase()) ||
            user.email.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesRole = roleFilter === 'all' || user.role === roleFilter;
        return matchesSearch && matchesRole;
    });

    const getRoleBadge = (role: string) => {
        const roles: Record<string, { variant: "default" | "secondary" | "outline", label: string }> = {
            admin: {variant: "default", label: "Admin"},
            editor: {variant: "secondary", label: "Editor"},
            user: {variant: "outline", label: "User"}
        };
        return roles[role] || {variant: "outline", label: role};
    };

    const getStatusBadge = (status: string) => {
        return status === "active"
            ? <Badge variant="default" className="bg-green-500">Active</Badge>
            : <Badge variant="secondary">Inactive</Badge>;
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">User Management</h2>
                    <p className="text-slate-500 text-sm mt-1">Manage users, roles, and permissions</p>
                </div>
                <Button className="bg-blue-600 hover:bg-blue-700">
                    <Plus className="w-4 h-4 mr-2"/>
                    Add User
                </Button>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">Total Users</p>
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
                                <p className="text-sm text-slate-500">Active Users</p>
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
                                <p className="text-sm text-slate-500">Admins</p>
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
                                <p className="text-sm text-slate-500">Editors</p>
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
                        placeholder="Search users..."
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
                    <option value="all">All Roles</option>
                    <option value="admin">Admin</option>
                    <option value="editor">Editor</option>
                    <option value="user">User</option>
                </select>
            </div>

            {/* Users Table */}
            <Card>
                <CardHeader>
                    <CardTitle>All Users</CardTitle>
                    <CardDescription>Manage user accounts and permissions</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>User</TableHead>
                                <TableHead>Email</TableHead>
                                <TableHead>Role</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Joined</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredUsers.map((user) => (
                                <TableRow key={user.id}>
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <Avatar className="w-10 h-10">
                                                <AvatarImage src={user.avatar}/>
                                                <AvatarFallback>{user.nickname.charAt(0)}</AvatarFallback>
                                            </Avatar>
                                            <div>
                                                <p className="font-medium">{user.nickname}</p>
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
                                                    View Profile
                                                </DropdownMenuItem>
                                                <DropdownMenuItem>
                                                    <Edit className="w-4 h-4 mr-2"/>
                                                    Edit
                                                </DropdownMenuItem>
                                                <DropdownMenuItem>
                                                    <Shield className="w-4 h-4 mr-2"/>
                                                    Change Role
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