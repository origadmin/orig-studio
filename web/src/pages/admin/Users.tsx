import {Spinner} from "@/components/ui/spinner"
/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 用户管理页面
 */

import {useState, useEffect} from 'react';
import {Search, Plus, User as UserIcon, MoreVertical, Trash2, Edit, Shield, Mail, Eye, RotateCcw} from 'lucide-react';
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
    DropdownMenuLabel,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {userApi, User, CreateUserRequest, UpdateUserRequest} from '@/lib/api/user';
import {useTranslation} from 'react-i18next';
import {getFullUrl} from '@/lib/utils';
import {TablePagination} from '@/components/common/TablePagination';

export default function UsersPage() {
    const {t} = useTranslation();
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchParams, setSearchParams] = useState({keyword: '', role: 'all', page: 1, page_size: 20});
    const [total, setTotal] = useState(0);
    
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [showChangeRoleDialog, setShowChangeRoleDialog] = useState(false);
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    
    const [currentUser, setCurrentUser] = useState<User | null>(null);
    const [formData, setFormData] = useState<Partial<CreateUserRequest & UpdateUserRequest>>({
        username: '',
        email: '',
        password: '',
        role: 'user',
        status: 'active',
    });

    useEffect(() => {
        loadUsers();
    }, [searchParams.page]);

    const loadUsers = async (params = searchParams) => {
        try {
            setLoading(true);
            const apiParams: any = {page: params.page, page_size: params.page_size};
            if (params.keyword) {
                apiParams.keyword = params.keyword;
            }
            if (params.role && params.role !== 'all') {
                apiParams.role = params.role;
            }
            const response = await userApi.list(apiParams);
            const userList = Array.isArray(response?.items) ? response.items : [];
            setUsers(userList);
            if (response?.total !== undefined) {
                setTotal(response.total);
            }
        } catch (error) {
            console.error('Failed to fetch users:', error);
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setFormData({
            username: '',
            email: '',
            password: '',
            role: 'user',
            status: 'active',
        });
    };

    const handleCreate = async () => {
        try {
            await userApi.create(formData as CreateUserRequest);
            await loadUsers();
            setShowCreateDialog(false);
            resetForm();
        } catch (err) {
            console.error('Failed to create user:', err);
        }
    };

    const handleUpdate = async () => {
        if (!currentUser) return;

        try {
            await userApi.update(currentUser.id, formData as UpdateUserRequest);
            await loadUsers();
            setShowEditDialog(false);
            resetForm();
            setCurrentUser(null);
        } catch (err) {
            console.error('Failed to update user:', err);
        }
    };

    const handleChangeRole = async (newRole: string) => {
        if (!currentUser) return;

        try {
            await userApi.update(currentUser.id, {role: newRole});
            await loadUsers();
            setShowChangeRoleDialog(false);
            setCurrentUser(null);
        } catch (err) {
            console.error('Failed to change role:', err);
        }
    };

    const handleDelete = async () => {
        if (!currentUser) return;

        try {
            await userApi.delete(currentUser.id);
            await loadUsers();
            setShowDeleteDialog(false);
            setCurrentUser(null);
        } catch (err) {
            console.error('Failed to delete user:', err);
        }
    };

    const openCreateDialog = () => {
        resetForm();
        setShowCreateDialog(true);
    };

    const openEditDialog = (user: User) => {
        setCurrentUser(user);
        setFormData({
            username: user.username,
            email: user.email,
            role: user.role,
            status: user.status,
        });
        setShowEditDialog(true);
    };

    const openChangeRoleDialog = (user: User) => {
        setCurrentUser(user);
        setShowChangeRoleDialog(true);
    };

    const openDeleteDialog = (user: User) => {
        setCurrentUser(user);
        setShowDeleteDialog(true);
    };

    const handleViewProfile = (user: User) => {
        window.open(`/members/${user.username}`, '_blank');
    };

    const getRoleBadge = (role: string) => {
        const roles: Record<string, { variant: "default" | "secondary" | "outline", label: string }> = {
            admin: {variant: "default", label: t('admin.admin') || "Admin"},
            editor: {variant: "secondary", label: t('admin.editor') || "Editor"},
            user: {variant: "outline", label: t('admin.user') || "User"}
        };
        return roles[role] || {variant: "outline", label: role};
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
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
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
                                        value={searchParams.keyword}
                                        onChange={(e) => setSearchParams({...searchParams, keyword: e.target.value})}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={searchParams.role} onValueChange={(v) => setSearchParams({...searchParams, role: v})}>
                                    <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {searchParams.role === 'all' ? (
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
                                        onClick={() => {
                                            const newParams = {keyword: '', role: 'all', page: 1, page_size: 20};
                                            setSearchParams(newParams);
                                            loadUsers(newParams);
                                        }}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2"/>
                                        Reset
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        onClick={() => loadUsers()}
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
                <Card key="total-users" className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.totalUsers') || "Total Users"}</p>
                                <p className="text-2xl font-bold text-info dark:text-blue-400">{users.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <UserIcon className="w-6 h-6 text-info"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                </Card>
                <Card key="active-users" className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.activeUsers') || "Active Users"}</p>
                                <p className="text-2xl font-bold text-success dark:text-green-400">{users.filter(u => u.status === 'active').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6 text-success"/>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                </Card>
                <Card key="admins" className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
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
                <Card key="editors" className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
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
                            <Button size="sm" onClick={openCreateDialog}>
                                <Plus className="w-4 h-4 mr-2"/>
                                {t('admin.addUser') || "Add User"}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center">
                            <Spinner className="mx-auto" />
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
                                {users.length > 0 ? users.map((user) => (
                                    <TableRow key={user.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <Avatar className="w-10 h-10">
                                                    <AvatarImage
                                                        src={user.avatar ? getFullUrl(user.avatar) : undefined}/>
                                                    <AvatarFallback>{user.username ? user.username.charAt(0) : '?'}</AvatarFallback>
                                                </Avatar>
                                                <div>
                                                    <p className="font-medium">{user.username}</p>
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
                                            {user.status === "active" ? (
                                                <Badge variant="default" className="bg-success">{t('admin.active') || "Active"}</Badge>
                                            ) : (
                                                <Badge variant="secondary">{t('admin.inactive') || "Inactive"}</Badge>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-sm text-slate-500">{user.created_at}</TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button 
                                                        variant="ghost" 
                                                        size="icon" 
                                                        className="h-6 w-6" 
                                                        title="More Actions"
                                                    >
                                                        <MoreVertical className="h-3 w-3"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem onClick={() => handleViewProfile(user)}>
                                                        <Eye className="w-4 h-4 mr-2"/>
                                                        {t('admin.viewProfile') || "View Profile"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => openEditDialog(user)}>
                                                        <Edit className="w-4 h-4 mr-2"/>
                                                        {t('admin.edit') || "Edit"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => openChangeRoleDialog(user)}>
                                                        <Shield className="w-4 h-4 mr-2"/>
                                                        {t('admin.changeRole') || "Change Role"}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem 
                                                        className="text-destructive focus:text-destructive" 
                                                        onClick={() => openDeleteDialog(user)}
                                                    >
                                                        <Trash2 className="w-4 h-4 mr-2"/>
                                                        {t('admin.delete') || "Delete"}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow key="empty">
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

            <TablePagination
                page={searchParams.page}
                pageSize={searchParams.page_size}
                total={total}
                onPageChange={(p) => setSearchParams({...searchParams, page: p})}
            />

            {/* Create User Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.addUser') || "Add User"}</DialogTitle>
                        <DialogDescription>{t('admin.createNewUser') || "Create a new user account"}</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.username') || "Username"}</label>
                            <Input
                                value={formData.username}
                                onChange={(e) => setFormData({...formData, username: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.email') || "Email"}</label>
                            <Input
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({...formData, email: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.password') || "Password"}</label>
                            <Input
                                type="password"
                                value={formData.password}
                                onChange={(e) => setFormData({...formData, password: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.role') || "Role"}</label>
                            <Select value={formData.role} onValueChange={(v) => setFormData({...formData, role: v})}>
                                <SelectTrigger>
                                    <SelectValue/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="user">{t('admin.user') || "User"}</SelectItem>
                                    <SelectItem value="editor">{t('admin.editor') || "Editor"}</SelectItem>
                                    <SelectItem value="admin">{t('admin.admin') || "Admin"}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            {t('admin.cancel') || "Cancel"}
                        </Button>
                        <Button onClick={handleCreate}>
                            {t('admin.create') || "Create"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit User Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.editUser') || "Edit User"}</DialogTitle>
                        <DialogDescription>{t('admin.editUserDesc') || "Edit user account details"}</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.username') || "Username"}</label>
                            <Input
                                value={formData.username}
                                onChange={(e) => setFormData({...formData, username: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.email') || "Email"}</label>
                            <Input
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({...formData, email: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.role') || "Role"}</label>
                            <Select value={formData.role} onValueChange={(v) => setFormData({...formData, role: v})}>
                                <SelectTrigger>
                                    <SelectValue/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="user">{t('admin.user') || "User"}</SelectItem>
                                    <SelectItem value="editor">{t('admin.editor') || "Editor"}</SelectItem>
                                    <SelectItem value="admin">{t('admin.admin') || "Admin"}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.status') || "Status"}</label>
                            <Select value={formData.status} onValueChange={(v) => setFormData({...formData, status: v})}>
                                <SelectTrigger>
                                    <SelectValue/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">{t('admin.active') || "Active"}</SelectItem>
                                    <SelectItem value="inactive">{t('admin.inactive') || "Inactive"}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                            {t('admin.cancel') || "Cancel"}
                        </Button>
                        <Button onClick={handleUpdate}>
                            {t('admin.save') || "Save"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Change Role Dialog */}
            <Dialog open={showChangeRoleDialog} onOpenChange={setShowChangeRoleDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.changeRole') || "Change Role"}</DialogTitle>
                        <DialogDescription>{t('admin.changeRoleDesc') || "Change user role"}</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.newRole') || "New Role"}</label>
                            <Select 
                                value={currentUser?.role || 'user'} 
                                onValueChange={(v) => handleChangeRole(v)}
                            >
                                <SelectTrigger>
                                    <SelectValue/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="user">{t('admin.user') || "User"}</SelectItem>
                                    <SelectItem value="editor">{t('admin.editor') || "Editor"}</SelectItem>
                                    <SelectItem value="admin">{t('admin.admin') || "Admin"}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Delete User Alert Dialog */}
            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t('admin.deleteUser') || "Delete User"}</AlertDialogTitle>
                        <AlertDialogDescription>
                            {t('admin.deleteUserConfirm') || "Are you sure you want to delete this user? This action cannot be undone."}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>{t('admin.cancel') || "Cancel"}</AlertDialogCancel>
                        <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleDelete}>
                            {t('admin.delete') || "Delete"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}