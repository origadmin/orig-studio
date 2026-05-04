import React, {useState, useCallback, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {useQueryClient} from '@tanstack/react-query';
import {channelApi, type CreateChannelInput, type ChannelLimits} from '@/lib/api/channel';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Label} from '@/components/ui/label';
import {Badge} from '@/components/ui/badge';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {useAuth} from '@/hooks/useAuth';
import {useChannelLimits} from '@/hooks/queries';

interface CreateChannelDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess?: (channel: {id: string; handle: string; short_token?: string}) => void;
}

const HANDLE_REGEX = /^[a-zA-Z][a-zA-Z0-9-]{2,38}$/;
const MAX_TAGS = 10;

export function CreateChannelDialog({open, onOpenChange, onSuccess}: CreateChannelDialogProps) {
    const {t} = useTranslation();
    const {isAuthenticated} = useAuth();
    const queryClient = useQueryClient();

    const {data: limits} = useChannelLimits(isAuthenticated && open);

    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<CreateChannelInput>({
        name: '',
        handle: '',
        description: '',
        privacy: 'PUBLIC',
        tags: [],
    });
    const [tagInput, setTagInput] = useState('');
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [handleStatus, setHandleStatus] = useState<'idle' | 'checking' | 'available' | 'taken' | 'invalid'>('idle');

    // Reset form when dialog opens
    useEffect(() => {
        if (open) {
            setFormData({
                name: '',
                handle: '',
                description: '',
                privacy: 'PUBLIC',
                tags: [],
            });
            setTagInput('');
            setErrors({});
            setHandleStatus('idle');
        }
    }, [open]);

    // Handle uniqueness check with debounce
    useEffect(() => {
        if (!formData.handle || !HANDLE_REGEX.test(formData.handle)) {
            setHandleStatus(formData.handle ? 'invalid' : 'idle');
            return;
        }

        setHandleStatus('checking');
        const timer = setTimeout(async () => {
            try {
                const res = await channelApi.validateHandle(formData.handle);
                const data = (res as any)?.data ?? res;
                setHandleStatus(data?.available ? 'available' : 'taken');
            } catch {
                setHandleStatus('idle');
            }
        }, 500);

        return () => clearTimeout(timer);
    }, [formData.handle]);

    // Auto-generate handle from name
    const handleNameChange = useCallback((name: string) => {
        setFormData(prev => {
            const autoHandle = name
                .toLowerCase()
                .replace(/[^a-z0-9-]/g, '-')
                .replace(/-+/g, '-')
                .replace(/^-|-$/g, '')
                .slice(0, 39);

            return {
                ...prev,
                name,
                handle: prev.handle === autoGenerateHandle(prev.name) ? autoHandle : prev.handle,
            };
        });
    }, []);

    // Helper to generate handle from name (for comparison)
    function autoGenerateHandle(name: string): string {
        return name
            .toLowerCase()
            .replace(/[^a-z0-9-]/g, '-')
            .replace(/-+/g, '-')
            .replace(/^-|-$/g, '')
            .slice(0, 39);
    }

    const addTag = useCallback(() => {
        const tag = tagInput.trim().toLowerCase();
        if (tag && formData.tags && formData.tags.length < MAX_TAGS && !formData.tags.includes(tag)) {
            setFormData(prev => ({...prev, tags: [...(prev.tags || []), tag]}));
            setTagInput('');
        }
    }, [tagInput, formData.tags]);

    const removeTag = useCallback((tag: string) => {
        setFormData(prev => ({...prev, tags: (prev.tags || []).filter(t => t !== tag)}));
    }, []);

    const validate = useCallback((): boolean => {
        const newErrors: Record<string, string> = {};

        if (!formData.name || formData.name.length < 3) {
            newErrors.name = t('channel.create.errors.name_min_length', {min: 3});
        }
        if (formData.name.length > 150) {
            newErrors.name = t('channel.create.errors.name_max_length', {max: 150});
        }
        if (!formData.handle || !HANDLE_REGEX.test(formData.handle)) {
            newErrors.handle = t('channel.create.errors.handle_format');
        }
        if (handleStatus === 'taken') {
            newErrors.handle = t('channel.create.errors.handle_taken');
        }
        if (handleStatus === 'checking') {
            newErrors.handle = t('channel.create.errors.handle_checking');
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    }, [formData, handleStatus, t]);

    const handleSubmit = useCallback(async () => {
        if (!validate()) return;

        setLoading(true);
        try {
            const res = await channelApi.create(formData);
            const channel = (res as any)?.data ?? res;

            // Invalidate channel queries
            queryClient.invalidateQueries({queryKey: ['channels', 'me']});
            queryClient.invalidateQueries({queryKey: ['channel', 'limits']});

            onOpenChange(false);
            onSuccess?.({id: channel?.id, handle: channel?.handle || formData.handle, short_token: channel?.short_token});
        } catch (err: any) {
            const msg = err?.response?.data?.message || err?.message || t('channel.create.errors.generic');
            if (msg.includes('channel_limit_reached')) {
                setErrors(prev => ({...prev, _form: t('channel.create.errors.limit_reached')}));
            } else if (msg.includes('handle_already_taken')) {
                setErrors(prev => ({...prev, handle: t('channel.create.errors.handle_taken')}));
            } else {
                setErrors(prev => ({...prev, _form: msg}));
            }
        } finally {
            setLoading(false);
        }
    }, [formData, validate, queryClient, onOpenChange, onSuccess, t]);

    // Check if creation is allowed
    const canCreate = limits ? (limits as ChannelLimits).can_create : true;
    const currentCount = limits ? (limits as ChannelLimits).current_count : 0;
    const maxChannels = limits ? (limits as ChannelLimits).max_channels : -1;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>{t('channel.create.title')}</DialogTitle>
                    <DialogDescription>
                        {canCreate
                            ? t('channel.create.description', {
                                current: currentCount,
                                max: maxChannels === -1 ? t('common.unlimited') : maxChannels,
                            })
                            : t('channel.create.limit_reached')
                        }
                    </DialogDescription>
                </DialogHeader>

                {!canCreate ? (
                    <div className="py-4 text-center text-muted-foreground">
                        {t('channel.create.limit_reached_message', {max: maxChannels})}
                    </div>
                ) : (
                    <div className="grid gap-4 py-4">
                        {errors._form && (
                            <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
                                {errors._form}
                            </div>
                        )}

                        <div className="grid gap-2">
                            <Label htmlFor="channel-name">{t('channel.create.name_label')}</Label>
                            <Input
                                id="channel-name"
                                value={formData.name}
                                onChange={e => handleNameChange(e.target.value)}
                                placeholder={t('channel.create.name_placeholder')}
                                maxLength={150}
                            />
                            {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
                        </div>

                        <div className="grid gap-2">
                            <Label htmlFor="channel-handle">{t('channel.create.handle_label')}</Label>
                            <div className="flex items-center gap-2">
                                <span className="text-muted-foreground">@</span>
                                <Input
                                    id="channel-handle"
                                    value={formData.handle}
                                    onChange={e => setFormData(prev => ({...prev, handle: e.target.value}))}
                                    placeholder={t('channel.create.handle_placeholder')}
                                    maxLength={39}
                                    className={handleStatus === 'available' ? 'border-green-500' : handleStatus === 'taken' ? 'border-red-500' : ''}
                                />
                                {handleStatus === 'checking' && (
                                    <span className="text-xs text-muted-foreground">{t('channel.create.checking')}</span>
                                )}
                                {handleStatus === 'available' && (
                                    <Badge variant="secondary" className="bg-green-100 text-green-800">
                                        {t('channel.create.handle_available')}
                                    </Badge>
                                )}
                                {handleStatus === 'taken' && (
                                    <Badge variant="destructive">{t('channel.create.handle_taken')}</Badge>
                                )}
                            </div>
                            {errors.handle && <p className="text-sm text-destructive">{errors.handle}</p>}
                            <p className="text-xs text-muted-foreground">
                                {t('channel.create.handle_hint')}
                            </p>
                        </div>

                        <div className="grid gap-2">
                            <Label htmlFor="channel-description">{t('channel.create.description_label')}</Label>
                            <Textarea
                                id="channel-description"
                                value={formData.description}
                                onChange={e => setFormData(prev => ({...prev, description: e.target.value}))}
                                placeholder={t('channel.create.description_placeholder')}
                                rows={3}
                            />
                        </div>

                        <div className="grid gap-2">
                            <Label>{t('channel.create.privacy_label')}</Label>
                            <Select
                                value={formData.privacy}
                                onValueChange={value => setFormData(prev => ({...prev, privacy: value}))}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="PUBLIC">{t('channel.privacy.public')}</SelectItem>
                                    <SelectItem value="PRIVATE">{t('channel.privacy.private')}</SelectItem>
                                    <SelectItem value="UNLISTED">{t('channel.privacy.unlisted')}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="grid gap-2">
                            <Label>{t('channel.create.tags_label')}</Label>
                            <div className="flex gap-2">
                                <Input
                                    value={tagInput}
                                    onChange={e => setTagInput(e.target.value)}
                                    placeholder={t('channel.create.tags_placeholder')}
                                    onKeyDown={e => {
                                        if (e.key === 'Enter') {
                                            e.preventDefault();
                                            addTag();
                                        }
                                    }}
                                    disabled={(formData.tags?.length || 0) >= MAX_TAGS}
                                />
                                <Button type="button" variant="outline" size="sm" onClick={addTag}>
                                    {t('common.add')}
                                </Button>
                            </div>
                            {formData.tags && formData.tags.length > 0 && (
                                <div className="flex flex-wrap gap-1 mt-1">
                                    {formData.tags.map(tag => (
                                        <Badge key={tag} variant="secondary" className="cursor-pointer" onClick={() => removeTag(tag)}>
                                            {tag} &times;
                                        </Badge>
                                    ))}
                                </div>
                            )}
                            <p className="text-xs text-muted-foreground">
                                {t('channel.create.tags_hint', {max: MAX_TAGS, current: formData.tags?.length || 0})}
                            </p>
                        </div>
                    </div>
                )}

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
                        {t('common.cancel')}
                    </Button>
                    {canCreate && (
                        <Button
                            onClick={handleSubmit}
                            disabled={loading || handleStatus === 'taken' || handleStatus === 'checking'}
                        >
                            {loading ? t('common.creating') : t('channel.create.submit')}
                        </Button>
                    )}
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
