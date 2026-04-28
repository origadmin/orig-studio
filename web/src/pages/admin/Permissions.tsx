import {useState} from 'react';
import {
    Shield, Plus, Edit, Trash2, Users, ToggleLeft, ToggleRight,
    ChevronRight, Loader2, Search, ArrowLeft,
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
    Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog';
import {
    AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
    AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {Label} from '@/components/ui/label';
import {Textarea} from '@/components/ui/textarea';
import {Separator} from '@/components/ui/separator';
import {
    usePermissionGroups, usePermissionGroup, useGroupMembers,
    useUserPermissions,
} from '@/hooks/queries';
import {adminPermissionApi, type CreatePermissionGroupRequest, type UpdatePermissionGroupRequest} from '@/lib/api/permission';
import {useQueryClient} from '@tanstack/react-query';
import {TablePagination} from '@/components/common/TablePagination';

export default function PermissionsPage() {
    const queryClient = useQueryClient();
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
    const [view, setView] = useState<'list' | 'detail' | 'members' | 'user-perms'>('list');

    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [editingGroup, setEditingGroup] = useState<{id: string; name: string; description: string; permissions: string[]; category_scope: string[]} | null>(null);
    const [createForm, setCreateForm] = useState<CreatePermissionGroupRequest>({name: '', description: '', permissions: [], category_scope: []});
    const [editForm, setEditForm] = useState<UpdatePermissionGroupRequest>({});

    const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false);
    const [addMemberIds, setAddMemberIds] = useState('');
    const [userPermId, setUserPermId] = useState<string | null>(null);

    const {data: groupsData, isLoading} = usePermissionGroups({page, page_size: 20});
    const {data: groupDetail} = usePermissionGroup(selectedGroupId);
    const {data: membersData} = useGroupMembers(selectedGroupId, {page: 1, page_size: 50});
    const {data: userPermsData} = useUserPermissions(userPermId);

    const groups = groupsData?.items || [];
    const members = membersData?.items || [];

    const filteredGroups = search
        ? groups.filter(g => g.name.toLowerCase().includes(search.toLowerCase()))
        : groups;

    const handleCreate = async () => {
        try {
            await adminPermissionApi.create(createForm);
            setCreateDialogOpen(false);
            setCreateForm({name: '', description: '', permissions: [], category_scope: []});
            queryClient.invalidateQueries({queryKey: ['permissionGroups']});
        } catch (err) {
            console.error('Failed to create group:', err);
        }
    };

    const handleUpdate = async () => {
        if (!editingGroup) return;
        try {
            await adminPermissionApi.update(editingGroup.id, editForm);
            setEditDialogOpen(false);
            queryClient.invalidateQueries({queryKey: ['permissionGroups']});
            queryClient.invalidateQueries({queryKey: ['permissionGroup', editingGroup.id]});
        } catch (err) {
            console.error('Failed to update group:', err);
        }
    };

    const handleDelete = async () => {
        if (!editingGroup) return;
        try {
            await adminPermissionApi.delete(editingGroup.id);
            setDeleteDialogOpen(false);
            setSelectedGroupId(null);
            setView('list');
            queryClient.invalidateQueries({queryKey: ['permissionGroups']});
        } catch (err) {
            console.error('Failed to delete group:', err);
        }
    };

    const handleToggle = async (id: string, isActive: boolean) => {
        try {
            await adminPermissionApi.toggle(id, {is_active: !isActive});
            queryClient.invalidateQueries({queryKey: ['permissionGroups']});
        } catch (err) {
            console.error('Failed to toggle group:', err);
        }
    };

    const handleAddMembers = async () => {
        if (!selectedGroupId || !addMemberIds.trim()) return;
        try {
            const ids = addMemberIds.split(',').map(s => s.trim()).filter(Boolean);
            await adminPermissionApi.addMembers(selectedGroupId, {user_ids: ids});
            setAddMemberDialogOpen(false);
            setAddMemberIds('');
            queryClient.invalidateQueries({queryKey: ['groupMembers', selectedGroupId]});
        } catch (err) {
            console.error('Failed to add members:', err);
        }
    };

    const handleRemoveMember = async (userId: string) => {
        if (!selectedGroupId) return;
        try {
            await adminPermissionApi.removeMember(selectedGroupId, userId);
            queryClient.invalidateQueries({queryKey: ['groupMembers', selectedGroupId]});
        } catch (err) {
            console.error('Failed to remove member:', err);
        }
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <Shield className="h-6 w-6"/>权限管理
                    </h1>
                    <p className="text-muted-foreground text-sm mt-1">管理权限组、成员分配和用户权限</p>
                </div>
                {view !== 'list' && (
                    <Button variant="outline" size="sm" onClick={() => { setView('list'); setSelectedGroupId(null); }}>
                        <ArrowLeft className="w-4 h-4 mr-2"/>返回列表
                    </Button>
                )}
                {view === 'list' && (
                    <Button size="sm" onClick={() => { setCreateForm({name: '', description: '', permissions: [], category_scope: []}); setCreateDialogOpen(true); }}>
                        <Plus className="w-4 h-4 mr-2"/>创建权限组
                    </Button>
                )}
            </div>

            {view === 'list' && (
                <>
                    <div className="flex items-center gap-2">
                        <div className="relative flex-1 max-w-sm">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                            <Input placeholder="搜索权限组..." value={search} onChange={e => setSearch(e.target.value)} className="pl-10 h-8 rounded-btn-sm"/>
                        </div>
                    </div>
                    <Card>
                        <CardContent className="p-0">
                            {isLoading ? (
                                <div className="py-12 text-center text-muted-foreground">加载中...</div>
                            ) : (
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>名称</TableHead>
                                            <TableHead>描述</TableHead>
                                            <TableHead>权限数</TableHead>
                                            <TableHead>成员数</TableHead>
                                            <TableHead>状态</TableHead>
                                            <TableHead className="text-right">操作</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {filteredGroups.length > 0 ? filteredGroups.map(group => (
                                            <TableRow key={group.id} className="cursor-pointer" onClick={() => { setSelectedGroupId(group.id); setView('detail'); }}>
                                                <TableCell className="font-medium">{group.name}</TableCell>
                                                <TableCell className="text-muted-foreground text-sm max-w-[200px] truncate">{group.description || '-'}</TableCell>
                                                <TableCell><Badge variant="outline">{group.permissions?.length || 0}</Badge></TableCell>
                                                <TableCell><Badge variant="secondary">{group.member_count || 0}</Badge></TableCell>
                                                <TableCell>
                                                    <Badge variant={group.is_active ? 'default' : 'secondary'}>
                                                        {group.is_active ? '启用' : '禁用'}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <div className="flex items-center justify-end gap-1" onClick={e => e.stopPropagation()}>
                                                        <Button variant="ghost" size="icon-sm"
                                                            onClick={() => handleToggle(group.id, group.is_active)}>
                                                            {group.is_active ? <ToggleRight className="h-4 w-4 text-success"/> : <ToggleLeft className="h-4 w-4 text-muted-foreground"/>}
                                                        </Button>
                                                        <Button variant="ghost" size="icon-sm"
                                                            onClick={() => { setEditingGroup({id: group.id, name: group.name, description: group.description || '', permissions: group.permissions, category_scope: group.category_scope || []}); setEditForm({name: group.name, description: group.description, permissions: group.permissions, category_scope: group.category_scope}); setEditDialogOpen(true); }}>
                                                            <Edit className="h-4 w-4"/>
                                                        </Button>
                                                        <Button variant="ghost" size="icon-sm" className="text-destructive"
                                                            onClick={() => { setEditingGroup({id: group.id, name: group.name, description: group.description || '', permissions: group.permissions, category_scope: group.category_scope || []}); setDeleteDialogOpen(true); }}>
                                                            <Trash2 className="h-4 w-4"/>
                                                        </Button>
                                                    </div>
                                                </TableCell>
                                            </TableRow>
                                        )) : (
                                            <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">暂无权限组</TableCell></TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            )}
                        </CardContent>
                    </Card>

                    <TablePagination
                        page={page}
                        pageSize={20}
                        total={groupsData?.total || 0}
                        onPageChange={setPage}
                    />
                </>
            )}

            {view === 'detail' && groupDetail && (
                <div className="space-y-4">
                    <Card>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle>{groupDetail.name}</CardTitle>
                                    <CardDescription>{groupDetail.description || '无描述'}</CardDescription>
                                </div>
                                <Badge variant={groupDetail.is_active ? 'default' : 'secondary'}>
                                    {groupDetail.is_active ? '启用' : '禁用'}
                                </Badge>
                            </div>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div>
                                <Label className="text-sm font-medium">权限列表</Label>
                                <div className="flex flex-wrap gap-2 mt-2">
                                    {groupDetail.permissions.map(p => (
                                        <Badge key={p} variant="outline">{p}</Badge>
                                    ))}
                                    {groupDetail.permissions.length === 0 && <span className="text-sm text-muted-foreground">无权限</span>}
                                </div>
                            </div>
                            {groupDetail.category_scope && groupDetail.category_scope.length > 0 && (
                                <div>
                                    <Label className="text-sm font-medium">分类范围</Label>
                                    <div className="flex flex-wrap gap-2 mt-2">
                                        {groupDetail.category_scope.map(c => (
                                            <Badge key={c} variant="secondary">{c}</Badge>
                                        ))}
                                    </div>
                                </div>
                            )}
                            <Separator/>
                            <div className="flex gap-2">
                                <Button variant="outline" size="sm" onClick={() => setView('members')}>
                                    <Users className="w-4 h-4 mr-2"/>管理成员 ({groupDetail.member_count || 0})
                                </Button>
                                <Button variant="outline" size="sm" onClick={() => { setUserPermId(null); setView('user-perms'); }}>
                                    <Shield className="w-4 h-4 mr-2"/>查看用户权限
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            )}

            {view === 'members' && (
                <Card>
                    <CardHeader>
                        <div className="flex items-center justify-between">
                            <CardTitle>组成员管理</CardTitle>
                            <Button size="sm" onClick={() => setAddMemberDialogOpen(true)}>
                                <Plus className="w-4 h-4 mr-2"/>添加成员
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-0">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>用户ID</TableHead>
                                    <TableHead>用户名</TableHead>
                                    <TableHead>加入时间</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {members.length > 0 ? members.map(m => (
                                    <TableRow key={m.id}>
                                        <TableCell className="font-mono text-sm">{m.user_id}</TableCell>
                                        <TableCell>{m.username}</TableCell>
                                        <TableCell className="text-sm text-muted-foreground">{m.joined_at}</TableCell>
                                        <TableCell className="text-right">
                                            <Button variant="ghost" size="sm" className="text-destructive"
                                                onClick={() => handleRemoveMember(m.user_id)}>
                                                <Trash2 className="w-4 h-4 mr-1"/>移除
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">暂无成员</TableCell></TableRow>
                                )}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {view === 'user-perms' && (
                <Card>
                    <CardHeader>
                        <CardTitle>用户权限查看</CardTitle>
                        <CardDescription>输入用户ID查看其权限详情</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex gap-2">
                            <Input placeholder="输入用户ID" value={userPermId || ''} onChange={e => setUserPermId(e.target.value)} className="max-w-xs"/>
                        </div>
                        {userPermsData && (
                            <div className="space-y-4">
                                <div>
                                    <Label className="text-sm font-medium">角色</Label>
                                    <Badge variant="outline" className="ml-2">{userPermsData.role}</Badge>
                                </div>
                                <div>
                                    <Label className="text-sm font-medium mb-2 block">所属权限组</Label>
                                    <div className="flex flex-wrap gap-2">
                                        {userPermsData.groups.map(g => (
                                            <Badge key={g.id} variant={g.is_active ? 'default' : 'secondary'}>
                                                {g.name}
                                            </Badge>
                                        ))}
                                        {userPermsData.groups.length === 0 && <span className="text-sm text-muted-foreground">无权限组</span>}
                                    </div>
                                </div>
                                <div>
                                    <Label className="text-sm font-medium mb-2 block">有效权限</Label>
                                    <div className="space-y-2">
                                        {Object.entries(userPermsData.effective_permissions).map(([key, val]) => (
                                            <div key={key} className="flex items-center gap-2 p-2 rounded bg-muted/50">
                                                <Badge variant="outline">{key}</Badge>
                                                <span className="text-xs text-muted-foreground">来源: {val.sources.join(', ')}</span>
                                                {val.scope && <span className="text-xs text-muted-foreground">范围: {val.scope.join(', ')}</span>}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>
            )}

            {/* Create Dialog */}
            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>创建权限组</DialogTitle>
                        <DialogDescription>创建新的权限组并配置权限</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2">
                            <Label>名称</Label>
                            <Input value={createForm.name} onChange={e => setCreateForm({...createForm, name: e.target.value})} placeholder="权限组名称"/>
                        </div>
                        <div className="grid gap-2">
                            <Label>描述</Label>
                            <Textarea value={createForm.description || ''} onChange={e => setCreateForm({...createForm, description: e.target.value})} placeholder="权限组描述" rows={2}/>
                        </div>
                        <div className="grid gap-2">
                            <Label>权限 (逗号分隔)</Label>
                            <Input value={createForm.permissions.join(', ')} onChange={e => setCreateForm({...createForm, permissions: e.target.value.split(',').map(s => s.trim()).filter(Boolean)})} placeholder="media:read, media:write"/>
                        </div>
                        <div className="grid gap-2">
                            <Label>分类范围 (逗号分隔, 可选)</Label>
                            <Input value={(createForm.category_scope || []).join(', ')} onChange={e => setCreateForm({...createForm, category_scope: e.target.value.split(',').map(s => s.trim()).filter(Boolean)})} placeholder="category1, category2"/>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>取消</Button>
                        <Button onClick={handleCreate} disabled={!createForm.name}>创建</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit Dialog */}
            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>编辑权限组</DialogTitle>
                        <DialogDescription>修改权限组 "{editingGroup?.name}" 的配置</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2">
                            <Label>名称</Label>
                            <Input value={editForm.name || ''} onChange={e => setEditForm({...editForm, name: e.target.value})}/>
                        </div>
                        <div className="grid gap-2">
                            <Label>描述</Label>
                            <Textarea value={editForm.description || ''} onChange={e => setEditForm({...editForm, description: e.target.value})} rows={2}/>
                        </div>
                        <div className="grid gap-2">
                            <Label>权限 (逗号分隔)</Label>
                            <Input value={(editForm.permissions || []).join(', ')} onChange={e => setEditForm({...editForm, permissions: e.target.value.split(',').map(s => s.trim()).filter(Boolean)})}/>
                        </div>
                        <div className="grid gap-2">
                            <Label>分类范围 (逗号分隔)</Label>
                            <Input value={(editForm.category_scope || []).join(', ')} onChange={e => setEditForm({...editForm, category_scope: e.target.value.split(',').map(s => s.trim()).filter(Boolean)})}/>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
                        <Button onClick={handleUpdate}>保存</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Dialog */}
            <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认删除？</AlertDialogTitle>
                        <AlertDialogDescription>将永久删除权限组 "{editingGroup?.name}"，此操作无法撤销。</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">确认删除</AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            {/* Add Member Dialog */}
            <Dialog open={addMemberDialogOpen} onOpenChange={setAddMemberDialogOpen}>
                <DialogContent className="sm:max-w-[400px]">
                    <DialogHeader>
                        <DialogTitle>添加成员</DialogTitle>
                        <DialogDescription>输入用户ID，多个用逗号分隔</DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <Textarea value={addMemberIds} onChange={e => setAddMemberIds(e.target.value)} placeholder="user_id_1, user_id_2" rows={3}/>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setAddMemberDialogOpen(false)}>取消</Button>
                        <Button onClick={handleAddMembers} disabled={!addMemberIds.trim()}>添加</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
