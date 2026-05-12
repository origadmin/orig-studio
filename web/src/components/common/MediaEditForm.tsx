import {Label} from '@/components/ui/label';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Separator} from '@/components/ui/separator';
import {formatFileSize, formatDuration} from '@/lib/format';
import type {Media} from '@/lib/api/media';

export interface MediaEditFormState {
    title: string;
    description: string;
    category_id: string | number;
    tags: string;
    privacy: number;
    state: string;
    enable_comments: boolean;
    allow_download: boolean;
    featured?: boolean;
    listable?: boolean;
}

interface MediaEditFormProps {
    form: MediaEditFormState;
    setForm: (form: MediaEditFormState) => void;
    media: Media;
    categories: any;
    isAdmin: boolean;
    /** Whether to show admin-only fields (featured, listable). Portal should pass false even for admin users. */
    showAdminOnlyFields?: boolean;
}

export function MediaEditForm({form, setForm, media, categories, isAdmin, showAdminOnlyFields = true}: MediaEditFormProps) {
    const categoriesList = (categories as any)?.items
        ? (categories as any).items
        : Array.isArray(categories) ? categories : [];

    const techResolution = media.width && media.height ? `${media.width} x ${media.height}` : 'N/A';
    const techDuration = media.duration ? formatDuration(media.duration) : 'N/A';
    const techMimeType = media.mime_type || 'N/A';
    const techFileSize = media.size ? formatFileSize(Number(media.size)) : 'N/A';
    const techExtension = media.extension || 'N/A';

    return (
        <div className="space-y-6">
            <div className="space-y-2">
                <Label htmlFor="title">Title</Label>
                <Input
                    id="title"
                    value={form.title}
                    onChange={e => setForm({...form, title: e.target.value})}
                />
            </div>

            <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <textarea
                    id="description"
                    className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    value={form.description}
                    onChange={e => setForm({...form, description: e.target.value})}
                    placeholder="Describe your media..."
                />
            </div>

            <div className="space-y-2">
                <Label>Category</Label>
                <Select
                    value={form.category_id !== '' && form.category_id !== undefined ? String(form.category_id) : '_none_'}
                    onValueChange={val => setForm({...form, category_id: val === '_none_' ? '' : val})}
                >
                    <SelectTrigger>
                        <SelectValue placeholder="Select category"/>
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="_none_">No category</SelectItem>
                        {categoriesList.map((cat: any) => (
                            <SelectItem key={cat.id} value={String(cat.id)}>
                                {cat.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            <div className="space-y-2">
                <Label htmlFor="tags">Tags (comma separated)</Label>
                <Input
                    id="tags"
                    value={form.tags}
                    onChange={e => setForm({...form, tags: e.target.value})}
                    placeholder="e.g. tutorial, coding, devops"
                />
                {form.tags && (
                    <div className="flex flex-wrap gap-1 mt-2">
                        {form.tags.split(',').map((tag, i) => tag.trim() && (
                            <Badge key={i} variant="secondary" className="text-xs">{tag.trim()}</Badge>
                        ))}
                    </div>
                )}
            </div>

            <Separator/>

            <div className="space-y-3">
                <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">Technical Info</h3>
                <div className="grid grid-cols-2 gap-x-6 gap-y-3">
                    <div className="space-y-1">
                        <Label className="text-xs text-muted-foreground">Resolution</Label>
                        <p className="text-sm font-mono">{techResolution}</p>
                    </div>
                    <div className="space-y-1">
                        <Label className="text-xs text-muted-foreground">Duration</Label>
                        <p className="text-sm font-mono">{techDuration}</p>
                    </div>
                    <div className="space-y-1">
                        <Label className="text-xs text-muted-foreground">MIME Type</Label>
                        <p className="text-sm font-mono">{techMimeType}</p>
                    </div>
                    <div className="space-y-1">
                        <Label className="text-xs text-muted-foreground">File Size</Label>
                        <p className="text-sm font-mono">{techFileSize}</p>
                    </div>
                    <div className="space-y-1">
                        <Label className="text-xs text-muted-foreground">Extension</Label>
                        <p className="text-sm font-mono">{techExtension}</p>
                    </div>
                </div>
            </div>

            <Separator/>

            <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                    <Label>Privacy</Label>
                    <Select
                        value={String(form.privacy)}
                        onValueChange={val => setForm({...form, privacy: Number(val)})}
                    >
                        <SelectTrigger>
                            <SelectValue/>
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="1">Public</SelectItem>
                            <SelectItem value="3">Unlisted</SelectItem>
                            <SelectItem value="2">Private</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                {isAdmin && (
                    <div className="space-y-2">
                        <Label>State</Label>
                        <Select
                            value={form.state}
                            onValueChange={val => setForm({...form, state: val})}
                        >
                            <SelectTrigger>
                                <SelectValue/>
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="draft">Draft</SelectItem>
                                <SelectItem value="active">Published</SelectItem>
                                <SelectItem value="deleted">Deleted</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                )}
            </div>

            <Separator/>

            <div className="grid grid-cols-2 gap-6">
                <div className="flex items-center gap-3">
                    <input
                        type="checkbox"
                        id="enable_comments"
                        checked={form.enable_comments}
                        onChange={e => setForm({...form, enable_comments: e.target.checked})}
                        className="h-4 w-4 rounded border-input text-primary focus:ring-primary"
                    />
                    <div>
                        <Label htmlFor="enable_comments" className="cursor-pointer">Allow Comments</Label>
                        <p className="text-xs text-muted-foreground">Users can leave comments</p>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <input
                        type="checkbox"
                        id="allow_download"
                        checked={form.allow_download}
                        onChange={e => setForm({...form, allow_download: e.target.checked})}
                        className="h-4 w-4 rounded border-input text-primary focus:ring-primary"
                    />
                    <div>
                        <Label htmlFor="allow_download" className="cursor-pointer">Allow Download</Label>
                        <p className="text-xs text-muted-foreground">Users can download the original file</p>
                    </div>
                </div>
            </div>

            {showAdminOnlyFields && (
                <>
                    <Separator/>
                    <div className="grid grid-cols-2 gap-6">
                        <div className="flex items-center gap-3">
                            <input
                                type="checkbox"
                                id="featured"
                                checked={form.featured ?? false}
                                onChange={e => setForm({...form, featured: e.target.checked})}
                                className="h-4 w-4 rounded border-input text-primary focus:ring-primary"
                            />
                            <div>
                                <Label htmlFor="featured" className="cursor-pointer">Featured</Label>
                                <p className="text-xs text-muted-foreground">Show in featured section</p>
                            </div>
                        </div>
                        <div className="flex items-center gap-3">
                            <input
                                type="checkbox"
                                id="listable"
                                checked={form.listable ?? false}
                                onChange={e => setForm({...form, listable: e.target.checked})}
                                className="h-4 w-4 rounded border-input text-primary focus:ring-primary"
                            />
                            <div>
                                <Label htmlFor="listable" className="cursor-pointer">Listable</Label>
                                <p className="text-xs text-muted-foreground">Show in video listings</p>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}
