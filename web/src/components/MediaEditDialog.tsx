import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {mediaApi} from '@/lib/api/media';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Label} from '@/components/ui/label';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';

interface MediaEditDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    media: {
        id: number;
        title: string;
        description?: string;
        tags?: string[];
        state: string;
    } | null;
    onSuccess?: () => void;
}

export function MediaEditDialog({open, onOpenChange, media, onSuccess}: MediaEditDialogProps) {
    const {t} = useTranslation();
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        title: '',
        description: '',
        tags: '',
        state: 'active',
    });

    React.useEffect(() => {
        if (media) {
            setFormData({
                title: media.title || '',
                description: media.description || '',
                tags: media.tags?.join(', ') || '',
                state: media.state || 'active',
            });
        }
    }, [media]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!media) return;

        setLoading(true);
        try {
            const tags = formData.tags.split(',').map(t => t.trim()).filter(Boolean);
            await mediaApi.update(media.id.toString(), {
                title: formData.title,
                description: formData.description,
                tags,
                state: formData.state,
            });
            onSuccess?.();
            onOpenChange(false);
        } catch (error) {
            console.error('Failed to update media:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>{t('admin.edit') || '编辑'}</DialogTitle>
                    <DialogDescription>
                        {t('media.editDesc') || '修改媒体信息'}
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit}>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="title">{t('upload.titleLabel') || '标题'}</Label>
                            <Input
                                id="title"
                                value={formData.title}
                                onChange={(e) => setFormData({...formData, title: e.target.value})}
                                required
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="description">{t('upload.descLabel') || '描述'}</Label>
                            <Textarea
                                id="description"
                                value={formData.description}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                                rows={3}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="tags">{t('upload.tagLabel') || '标签'}</Label>
                            <Input
                                id="tags"
                                value={formData.tags}
                                onChange={(e) => setFormData({...formData, tags: e.target.value})}
                                placeholder="tag1, tag2, tag3"
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="state">{t('admin.status') || '状态'}</Label>
                            <Select
                                value={formData.state}
                                onValueChange={(value) => setFormData({...formData, state: value})}
                            >
                                <SelectTrigger>
                                    <SelectValue/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">{t('admin.enabled') || '启用'}</SelectItem>
                                    <SelectItem value="draft">Draft</SelectItem>
                                    <SelectItem value="deleted">{t('admin.delete') || '删除'}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            {t('common.cancel') || '取消'}
                        </Button>
                        <Button type="submit" disabled={loading}>
                            {loading ? t('common.loading') : t('common.save') || '保存'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}