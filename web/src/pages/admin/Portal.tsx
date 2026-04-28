import React, {useState} from 'react';
import {
    Layout, Image, Plus, Edit, Trash2, ToggleLeft, ToggleRight,
    GripVertical, Loader2, ExternalLink, ArrowUp, ArrowDown,
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Spinner} from '@/components/ui/spinner';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
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
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {
    useAdminNavItems, useAdminBanners,
    useCreateNavItem, useUpdateNavItem, useDeleteNavItem,
    useCreateBanner, useUpdateBanner, useToggleBanner,
} from '@/hooks/queries';
import {type NavItem, type Banner, type CreateNavItemRequest, type CreateBannerRequest} from '@/lib/api/portal';
import {useQueryClient} from '@tanstack/react-query';

export default function PortalConfigPage() {
    return (
        <div className="space-y-4 p-4 md:p-6">
            <div>
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <Layout className="h-6 w-6"/>门户配置
                </h1>
                <p className="text-muted-foreground text-sm mt-1">管理导航栏、Banner和首页展示内容</p>
            </div>

            <Tabs defaultValue="navigation">
                <TabsList>
                    <TabsTrigger value="navigation">导航管理</TabsTrigger>
                    <TabsTrigger value="banners">Banner管理</TabsTrigger>
                </TabsList>
                <TabsContent value="navigation" className="mt-4">
                    <NavigationTab/>
                </TabsContent>
                <TabsContent value="banners" className="mt-4">
                    <BannersTab/>
                </TabsContent>
            </Tabs>
        </div>
    );
}

const NavigationTab: React.FC = () => {
    const queryClient = useQueryClient();
    const {data: navData, isLoading} = useAdminNavItems();
    const createMutation = useCreateNavItem();
    const updateMutation = useUpdateNavItem();
    const deleteMutation = useDeleteNavItem();

    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [editingItem, setEditingItem] = useState<NavItem | null>(null);
    const [createForm, setCreateForm] = useState<CreateNavItemRequest>({
        type: 'internal_link', label: '', url: '', sequence: 0, open_new_tab: false,
    });
    const [editForm, setEditForm] = useState({label: '', url: '', type: 'internal_link' as "internal_link" | "external_link" | "category", open_new_tab: false});

    const navItems = navData?.items || [];

    const handleCreate = async () => {
        try {
            await createMutation.mutateAsync(createForm);
            setCreateDialogOpen(false);
            setCreateForm({type: 'internal_link', label: '', url: '', sequence: 0, open_new_tab: false});
        } catch (err) {
            console.error('Failed to create nav item:', err);
        }
    };

    const handleUpdate = async () => {
        if (!editingItem) return;
        try {
            await updateMutation.mutateAsync({id: editingItem.id, data: editForm});
            setEditDialogOpen(false);
        } catch (err) {
            console.error('Failed to update nav item:', err);
        }
    };

    const handleDelete = async () => {
        if (!editingItem) return;
        try {
            await deleteMutation.mutateAsync(editingItem.id);
            setDeleteDialogOpen(false);
        } catch (err) {
            console.error('Failed to delete nav item:', err);
        }
    };

    const moveItem = async (index: number, direction: 'up' | 'down') => {
        const newItems = [...navItems];
        const targetIndex = direction === 'up' ? index - 1 : index + 1;
        if (targetIndex < 0 || targetIndex >= newItems.length) return;
        [newItems[index], newItems[targetIndex]] = [newItems[targetIndex], newItems[index]];
        try {
            await adminPortalApi_reorder(newItems.map(i => i.id));
        } catch (err) {
            console.error('Failed to reorder:', err);
        }
    };

    const adminPortalApi_reorder = async (ids: string[]) => {
        const {adminPortalApi} = await import('@/lib/api/portal');
        await adminPortalApi.reorderNavItems({ids});
        queryClient.invalidateQueries({queryKey: ['adminNavItems']});
        queryClient.invalidateQueries({queryKey: ['portalConfig']});
    };

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>导航项</CardTitle>
                            <CardDescription>管理顶部导航栏的链接项</CardDescription>
                        </div>
                        <Button size="sm" onClick={() => setCreateDialogOpen(true)}>
                            <Plus className="w-4 h-4 mr-2"/>添加导航项
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="py-12 text-center"><Spinner className="mx-auto"/></div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[40px]">排序</TableHead>
                                    <TableHead>标签</TableHead>
                                    <TableHead>类型</TableHead>
                                    <TableHead>链接</TableHead>
                                    <TableHead>新标签页</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {navItems.length > 0 ? navItems.map((item, idx) => (
                                    <TableRow key={item.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-1">
                                                <GripVertical className="w-4 h-4 text-muted-foreground"/>
                                                <Button variant="ghost" size="icon-sm" disabled={idx === 0}
                                                    onClick={() => moveItem(idx, 'up')}><ArrowUp className="w-3 h-3"/></Button>
                                                <Button variant="ghost" size="icon-sm" disabled={idx === navItems.length - 1}
                                                    onClick={() => moveItem(idx, 'down')}><ArrowDown className="w-3 h-3"/></Button>
                                            </div>
                                        </TableCell>
                                        <TableCell className="font-medium">{item.label}</TableCell>
                                        <TableCell><Badge variant="outline">{item.type}</Badge></TableCell>
                                        <TableCell className="text-sm text-muted-foreground max-w-[200px] truncate">{item.url}</TableCell>
                                        <TableCell>{item.open_new_tab ? <Badge variant="secondary">是</Badge> : <span className="text-xs text-muted-foreground">否</span>}</TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex items-center justify-end gap-1">
                                                <Button variant="ghost" size="icon-sm"
                                                    onClick={() => { setEditingItem(item); setEditForm({label: item.label, url: item.url, type: item.type, open_new_tab: item.open_new_tab}); setEditDialogOpen(true); }}>
                                                    <Edit className="w-4 h-4"/>
                                                </Button>
                                                <Button variant="ghost" size="icon-sm" className="text-destructive"
                                                    onClick={() => { setEditingItem(item); setDeleteDialogOpen(true); }}>
                                                    <Trash2 className="w-4 h-4"/>
                                                </Button>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">暂无导航项</TableCell></TableRow>
                                )}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
                <DialogContent className="sm:max-w-[450px]">
                    <DialogHeader><DialogTitle>添加导航项</DialogTitle></DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2"><Label>标签</Label><Input value={createForm.label} onChange={e => setCreateForm({...createForm, label: e.target.value})} placeholder="导航标签"/></div>
                        <div className="grid gap-2"><Label>类型</Label>
                            <Select value={createForm.type} onValueChange={v => setCreateForm({...createForm, type: v as any})}>
                                <SelectTrigger><SelectValue/></SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="internal_link">内部链接</SelectItem>
                                    <SelectItem value="external_link">外部链接</SelectItem>
                                    <SelectItem value="category">分类入口</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2"><Label>URL</Label><Input value={createForm.url} onChange={e => setCreateForm({...createForm, url: e.target.value})} placeholder="/featured 或 https://..."/></div>
                        <div className="grid gap-2"><Label>排序</Label><Input type="number" value={createForm.sequence} onChange={e => setCreateForm({...createForm, sequence: Number(e.target.value)})}/></div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>取消</Button>
                        <Button onClick={handleCreate} disabled={!createForm.label || !createForm.url}>添加</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
                <DialogContent className="sm:max-w-[450px]">
                    <DialogHeader><DialogTitle>编辑导航项</DialogTitle></DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2"><Label>标签</Label><Input value={editForm.label} onChange={e => setEditForm({...editForm, label: e.target.value})}/></div>
                        <div className="grid gap-2"><Label>类型</Label>
                            <Select value={editForm.type} onValueChange={v => setEditForm({...editForm, type: v as any})}>
                                <SelectTrigger><SelectValue/></SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="internal_link">内部链接</SelectItem>
                                    <SelectItem value="external_link">外部链接</SelectItem>
                                    <SelectItem value="category">分类入口</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2"><Label>URL</Label><Input value={editForm.url} onChange={e => setEditForm({...editForm, url: e.target.value})}/></div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
                        <Button onClick={handleUpdate}>保存</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader><AlertDialogTitle>确认删除？</AlertDialogTitle><AlertDialogDescription>将删除导航项 "{editingItem?.label}"</AlertDialogDescription></AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">确认删除</AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
};

const BannersTab: React.FC = () => {
    const {data: bannerData, isLoading} = useAdminBanners();
    const createMutation = useCreateBanner();
    const updateMutation = useUpdateBanner();
    const toggleMutation = useToggleBanner();

    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [editingBanner, setEditingBanner] = useState<Banner | null>(null);
    const [createForm, setCreateForm] = useState<CreateBannerRequest>({title: ''});
    const [editForm, setEditForm] = useState({title: '', subtitle: '', primary_btn_text: '', primary_btn_url: ''});

    const banners = bannerData?.items || [];

    const handleCreate = async () => {
        try {
            await createMutation.mutateAsync(createForm);
            setCreateDialogOpen(false);
            setCreateForm({title: ''});
        } catch (err) {
            console.error('Failed to create banner:', err);
        }
    };

    const handleUpdate = async () => {
        if (!editingBanner) return;
        try {
            await updateMutation.mutateAsync({id: editingBanner.id, data: editForm});
            setEditDialogOpen(false);
        } catch (err) {
            console.error('Failed to update banner:', err);
        }
    };

    const handleToggle = async (id: string) => {
        try {
            await toggleMutation.mutateAsync(id);
        } catch (err) {
            console.error('Failed to toggle banner:', err);
        }
    };

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>Banner管理</CardTitle>
                            <CardDescription>管理首页Banner轮播图</CardDescription>
                        </div>
                        <Button size="sm" onClick={() => setCreateDialogOpen(true)}>
                            <Plus className="w-4 h-4 mr-2"/>添加Banner
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="py-12 text-center"><Spinner className="mx-auto"/></div>
                    ) : banners.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {banners.map(banner => (
                                <Card key={banner.id} className={`overflow-hidden ${!banner.is_active ? 'opacity-60' : ''}`}>
                                    <div className="h-32 bg-gradient-to-r from-emerald-500 to-teal-600 relative">
                                        {banner.image_url && <img src={banner.image_url} alt="" className="w-full h-full object-cover"/>}
                                        <div className="absolute inset-0 bg-black/30 flex items-center justify-center">
                                            <div className="text-center text-white">
                                                <h3 className="font-bold text-lg">{banner.title}</h3>
                                                {banner.subtitle && <p className="text-sm opacity-80">{banner.subtitle}</p>}
                                            </div>
                                        </div>
                                    </div>
                                    <CardContent className="p-4">
                                        <div className="flex items-center justify-between">
                                            <div>
                                                <Badge variant={banner.is_active ? 'default' : 'secondary'}>
                                                    {banner.is_active ? '启用' : '禁用'}
                                                </Badge>
                                                {banner.badge_text && <Badge variant="outline" className="ml-2">{banner.badge_text}</Badge>}
                                            </div>
                                            <div className="flex items-center gap-1">
                                                <Button variant="ghost" size="icon-sm"
                                                    onClick={() => handleToggle(banner.id)}>
                                                    {banner.is_active ? <ToggleRight className="w-4 h-4 text-success"/> : <ToggleLeft className="w-4 h-4 text-muted-foreground"/>}
                                                </Button>
                                                <Button variant="ghost" size="icon-sm"
                                                    onClick={() => { setEditingBanner(banner); setEditForm({title: banner.title, subtitle: banner.subtitle || '', primary_btn_text: banner.primary_btn_text || '', primary_btn_url: banner.primary_btn_url || ''}); setEditDialogOpen(true); }}>
                                                    <Edit className="w-4 h-4"/>
                                                </Button>
                                            </div>
                                        </div>
                                        {banner.primary_btn_text && (
                                            <div className="mt-2 text-xs text-muted-foreground">
                                                CTA: {banner.primary_btn_text} → {banner.primary_btn_url}
                                            </div>
                                        )}
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    ) : (
                        <div className="py-12 text-center text-muted-foreground">暂无Banner</div>
                    )}
                </CardContent>
            </Card>

            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader><DialogTitle>创建Banner</DialogTitle></DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2"><Label>标题</Label><Input value={createForm.title} onChange={e => setCreateForm({...createForm, title: e.target.value})} placeholder="Banner标题"/></div>
                        <div className="grid gap-2"><Label>副标题</Label><Input value={createForm.subtitle || ''} onChange={e => setCreateForm({...createForm, subtitle: e.target.value})} placeholder="副标题(可选)"/></div>
                        <div className="grid gap-2"><Label>徽章文字</Label><Input value={createForm.badge_text || ''} onChange={e => setCreateForm({...createForm, badge_text: e.target.value})} placeholder="如: HOT, NEW"/></div>
                        <div className="grid gap-2"><Label>图片URL</Label><Input value={createForm.image_url || ''} onChange={e => setCreateForm({...createForm, image_url: e.target.value})} placeholder="/uploads/banners/xxx.jpg"/></div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="grid gap-2"><Label>主按钮文字</Label><Input value={createForm.primary_btn_text || ''} onChange={e => setCreateForm({...createForm, primary_btn_text: e.target.value})}/></div>
                            <div className="grid gap-2"><Label>主按钮链接</Label><Input value={createForm.primary_btn_url || ''} onChange={e => setCreateForm({...createForm, primary_btn_url: e.target.value})}/></div>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>取消</Button>
                        <Button onClick={handleCreate} disabled={!createForm.title}>创建</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader><DialogTitle>编辑Banner</DialogTitle></DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="grid gap-2"><Label>标题</Label><Input value={editForm.title} onChange={e => setEditForm({...editForm, title: e.target.value})}/></div>
                        <div className="grid gap-2"><Label>副标题</Label><Input value={editForm.subtitle} onChange={e => setEditForm({...editForm, subtitle: e.target.value})}/></div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="grid gap-2"><Label>主按钮文字</Label><Input value={editForm.primary_btn_text} onChange={e => setEditForm({...editForm, primary_btn_text: e.target.value})}/></div>
                            <div className="grid gap-2"><Label>主按钮链接</Label><Input value={editForm.primary_btn_url} onChange={e => setEditForm({...editForm, primary_btn_url: e.target.value})}/></div>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
                        <Button onClick={handleUpdate}>保存</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
};
