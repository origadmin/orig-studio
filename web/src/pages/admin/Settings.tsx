import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Separator} from '@/components/ui/separator';
import {Skeleton} from '@/components/ui/skeleton';
import {Switch} from '@/components/ui/switch';
import {Label} from '@/components/ui/label';
import {ThemeSwitcher} from '@/themes';
import {
    Settings as SettingsIcon,
    Database,
    Server,
    Mail,
    Shield,
    HardDrive,
    Globe,
    Palette,
    Save,
    RefreshCw,
    Loader2,
    CheckCircle,
    AlertCircle,
    Blocks,
    LayoutGrid,
    Plus,
    X,
    Send
} from 'lucide-react';
import {settingsApi, type GroupedSettings} from '@/lib/api/system';
import {api} from '@/lib/request';

interface FormData {
    site_name: string;
    site_description: string;
    base_urls: string[];
    primary_url: string;
    allow_registration: string;
    allow_upload: string;
    storage_base_path: string;
    storage_type: string;
    s3_endpoint: string;
    s3_region: string;
    s3_bucket: string;
    s3_access_key: string;
    s3_secret_key: string;
    s3_use_path_style: string;
    max_upload_size_video: string;
    max_upload_size_image: string;
    auto_transcode: string;
    transcode_method: string;
    allowed_video_formats: string;
    allowed_image_formats: string;
    max_video_duration: string;
    sprite_frame_interval: string;
    sprite_columns: string;
    sprite_frame_width: string;
    sprite_frame_height: string;
    sprite_max_frames: string;
    thumbnail_quality: string;
    thumbnail_resolution: string;
    thumbnail_position: string;
    auto_approve: string;
    require_review: string;
    smtp_host: string;
    smtp_port: string;
    smtp_user: string;
    smtp_password: string;
    smtp_sender_name: string;
    smtp_use_tls: string;
    min_password_length: string;
    require_email_verification: string;
    api_rate_limit: string;
    module_articles: boolean;
    module_videos: boolean;
    module_music: boolean;
    homepage_layout: string;
}

interface SystemInfo {
    version: string;
    goVersion: string;
    database: string;
    os: string;
    uptime: string;
    totalMemory: string;
    usedMemory: string;
    cpuUsage: string;
    memoryUsage: number;
    numCPU: number;
    numGoroutine: number;
}

interface StorageCapabilities {
    current_type: string;
    available_types: string[];
    s3_configured: boolean;
    s3_available: boolean;
    hybrid_available: boolean;
}

interface EmailStatus {
    configured: boolean;
}

const defaultFormData: FormData = {
    site_name: '',
    site_description: '',
    base_urls: [],
    primary_url: '',
    allow_registration: 'true',
    allow_upload: 'true',
    storage_base_path: '/var/media',
    storage_type: 'local',
    s3_endpoint: '',
    s3_region: '',
    s3_bucket: '',
    s3_access_key: '',
    s3_secret_key: '',
    s3_use_path_style: 'false',
    max_upload_size_video: '5120',
    max_upload_size_image: '20',
    auto_transcode: 'true',
    transcode_method: 'ffmpeg',
    allowed_video_formats: 'mp4, webm, mkv, avi, mov',
    allowed_image_formats: 'jpg, png, gif, webp',
    max_video_duration: '120',
    sprite_frame_interval: '10',
    sprite_columns: '10',
    sprite_frame_width: '120',
    sprite_frame_height: '68',
    sprite_max_frames: '100',
    thumbnail_quality: '85',
    thumbnail_resolution: '320x180',
    thumbnail_position: '00:00:01',
    auto_approve: 'true',
    require_review: 'false',
    smtp_host: '',
    smtp_port: '587',
    smtp_user: '',
    smtp_password: '',
    smtp_sender_name: '',
    smtp_use_tls: 'true',
    min_password_length: '8',
    require_email_verification: 'false',
    api_rate_limit: '60',
    module_articles: true,
    module_videos: true,
    module_music: false,
    homepage_layout: 'auto',
};

const Settings: React.FC = () => {
    const {t} = useTranslation();
    const [activeTab, setActiveTab] = useState('general');
    const [formData, setFormData] = useState<FormData>(defaultFormData);
    const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{type: 'success' | 'error', text: string} | null>(null);
    const [storageCaps, setStorageCaps] = useState<StorageCapabilities>({
        current_type: 'local',
        available_types: ['local'],
        s3_configured: false,
        s3_available: false,
        hybrid_available: false,
    });
    const [emailStatus, setEmailStatus] = useState<EmailStatus>({configured: false});
    const [emailTestSending, setEmailTestSending] = useState(false);
    const [emailTestTo, setEmailTestTo] = useState('');

    useEffect(() => {
        fetchSettings();
        fetchSystemInfo();
        fetchStorageCapabilities();
        fetchEmailStatus();
    }, []);

    const fetchSettings = async () => {
        try {
            const raw = await settingsApi.get();
            let grouped: GroupedSettings | null = null;
            if (raw && typeof raw === 'object') {
                if ('code' in raw && 'data' in raw && typeof (raw as Record<string, unknown>).data === 'object') {
                    grouped = (raw as Record<string, unknown>).data as GroupedSettings;
                } else {
                    grouped = raw as GroupedSettings;
                }
            }

            if (grouped) {
                const getSettingValue = (key: string): string => {
                    for (const category of Object.values(grouped!)) {
                        if (!Array.isArray(category)) continue;
                        const found = category.find(s => s.key === key);
                        if (found) return found.value;
                    }
                    return '';
                };

                const parseBytesToMB = (val: string): string => {
                    const bytes = parseInt(val);
                    if (isNaN(bytes) || bytes <= 0) return '';
                    return String(Math.round(bytes / 1024 / 1024));
                };

                const parseBaseUrls = (val: string): string[] => {
                    if (!val) return [];
                    try {
                        const parsed = JSON.parse(val);
                        return Array.isArray(parsed) ? parsed : [];
                    } catch {
                        return val ? [val] : [];
                    }
                };

                setFormData(prev => ({
                    ...prev,
                    site_name: getSettingValue('site_name') || prev.site_name,
                    site_description: getSettingValue('site_description') || prev.site_description,
                    base_urls: parseBaseUrls(getSettingValue('base_urls')),
                    primary_url: getSettingValue('primary_url') || prev.primary_url,
                    allow_registration: getSettingValue('allow_registration') || prev.allow_registration,
                    allow_upload: getSettingValue('allow_upload') || prev.allow_upload,
                    storage_base_path: getSettingValue('storage_base_path') || prev.storage_base_path,
                    storage_type: getSettingValue('storage_type') || prev.storage_type,
                    s3_endpoint: getSettingValue('s3_endpoint') || prev.s3_endpoint,
                    s3_region: getSettingValue('s3_region') || prev.s3_region,
                    s3_bucket: getSettingValue('s3_bucket') || prev.s3_bucket,
                    s3_access_key: getSettingValue('s3_access_key') || prev.s3_access_key,
                    s3_secret_key: getSettingValue('s3_secret_key') || prev.s3_secret_key,
                    s3_use_path_style: getSettingValue('s3_use_path_style') || prev.s3_use_path_style,
                    max_upload_size_video: parseBytesToMB(getSettingValue('max_upload_size_video')) || prev.max_upload_size_video,
                    max_upload_size_image: parseBytesToMB(getSettingValue('max_upload_size_image')) || prev.max_upload_size_image,
                    auto_transcode: getSettingValue('auto_transcode') || prev.auto_transcode,
                    transcode_method: getSettingValue('transcode_method') || prev.transcode_method,
                    allowed_video_formats: getSettingValue('allowed_video_formats') || prev.allowed_video_formats,
                    allowed_image_formats: getSettingValue('allowed_image_formats') || prev.allowed_image_formats,
                    max_video_duration: getSettingValue('max_video_duration') || prev.max_video_duration,
                    sprite_frame_interval: getSettingValue('sprite_frame_interval') || prev.sprite_frame_interval,
                    sprite_columns: getSettingValue('sprite_columns') || prev.sprite_columns,
                    sprite_frame_width: getSettingValue('sprite_frame_width') || prev.sprite_frame_width,
                    sprite_frame_height: getSettingValue('sprite_frame_height') || prev.sprite_frame_height,
                    sprite_max_frames: getSettingValue('sprite_max_frames') || prev.sprite_max_frames,
                    thumbnail_quality: getSettingValue('thumbnail_quality') || prev.thumbnail_quality,
                    thumbnail_resolution: getSettingValue('thumbnail_resolution') || prev.thumbnail_resolution,
                    thumbnail_position: getSettingValue('thumbnail_position') || prev.thumbnail_position,
                    auto_approve: getSettingValue('auto_approve') || prev.auto_approve,
                    require_review: getSettingValue('require_review') || prev.require_review,
                    smtp_host: getSettingValue('smtp_host') || prev.smtp_host,
                    smtp_port: getSettingValue('smtp_port') || prev.smtp_port,
                    smtp_user: getSettingValue('smtp_user') || prev.smtp_user,
                    smtp_password: getSettingValue('smtp_password') || prev.smtp_password,
                    smtp_sender_name: getSettingValue('smtp_sender_name') || prev.smtp_sender_name,
                    smtp_use_tls: getSettingValue('smtp_use_tls') || prev.smtp_use_tls,
                    min_password_length: getSettingValue('min_password_length') || prev.min_password_length,
                    require_email_verification: getSettingValue('require_email_verification') || prev.require_email_verification,
                    api_rate_limit: getSettingValue('api_rate_limit') || prev.api_rate_limit,
                    module_articles: getSettingValue('module_articles') === 'true',
                    module_videos: getSettingValue('module_videos') === 'true',
                    module_music: getSettingValue('module_music') === 'true',
                    homepage_layout: getSettingValue('homepage_layout') || prev.homepage_layout,
                }));
            }
        } catch (error) {
            console.error('Failed to fetch settings:', error);
            setMessage({type: 'error', text: t('settings.loadFailed')});
            setTimeout(() => setMessage(null), 3000);
        } finally {
            setLoading(false);
        }
    };

    const fetchSystemInfo = async () => {
        try {
            const info = await api.get<SystemInfo>('/admin/settings/info');
            setSystemInfo(info);
        } catch (error) {
            console.error('Failed to fetch system info:', error);
            setSystemInfo(null);
        }
    };

    const fetchStorageCapabilities = async () => {
        try {
            const caps = await api.get<StorageCapabilities>('/admin/settings/storage/capabilities');
            setStorageCaps(caps);
        } catch (error) {
            console.error('Failed to fetch storage capabilities:', error);
        }
    };

    const fetchEmailStatus = async () => {
        try {
            const status = await api.get<EmailStatus>('/system/settings/email/status');
            setEmailStatus(status);
        } catch (error) {
            console.error('Failed to fetch email status:', error);
        }
    };

    const handleInputChange = (field: keyof FormData, value: string | boolean) => {
        setFormData(prev => ({...prev, [field]: value}));
    };

    const handleBaseUrlChange = (index: number, value: string) => {
        setFormData(prev => {
            const newUrls = [...prev.base_urls];
            newUrls[index] = value;
            return {...prev, base_urls: newUrls};
        });
    };

    const handleAddBaseUrl = () => {
        setFormData(prev => ({...prev, base_urls: [...prev.base_urls, '']}));
    };

    const handleRemoveBaseUrl = (index: number) => {
        setFormData(prev => {
            const newUrls = prev.base_urls.filter((_, i) => i !== index);
            return {...prev, base_urls: newUrls};
        });
    };

    const handleEmailTest = async () => {
        if (!emailTestTo) return;
        try {
            setEmailTestSending(true);
            await api.post('/system/settings/email/test', {to: emailTestTo});
            setMessage({type: 'success', text: t('settings.emailTestSuccess')});
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to send test email:', error);
            setMessage({type: 'error', text: t('settings.emailTestFailed')});
            setTimeout(() => setMessage(null), 3000);
        } finally {
            setEmailTestSending(false);
        }
    };

    const handleSave = async () => {
        try {
            setSaving(true);
            const mbToBytes = (mb: string): string => {
                const val = parseInt(mb);
                if (isNaN(val) || val <= 0) return '0';
                return String(val * 1024 * 1024);
            };

            const settings = [
                {key: 'site_name', value: formData.site_name},
                {key: 'site_description', value: formData.site_description},
                {key: 'base_urls', value: JSON.stringify(formData.base_urls.filter(u => u.trim()))},
                {key: 'primary_url', value: formData.primary_url},
                {key: 'allow_registration', value: formData.allow_registration},
                {key: 'allow_upload', value: formData.allow_upload},
                {key: 'storage_base_path', value: formData.storage_base_path},
                {key: 'storage_type', value: formData.storage_type},
                {key: 's3_endpoint', value: formData.s3_endpoint},
                {key: 's3_region', value: formData.s3_region},
                {key: 's3_bucket', value: formData.s3_bucket},
                {key: 's3_access_key', value: formData.s3_access_key},
                {key: 's3_secret_key', value: formData.s3_secret_key},
                {key: 's3_use_path_style', value: formData.s3_use_path_style},
                {key: 'max_upload_size_video', value: mbToBytes(formData.max_upload_size_video)},
                {key: 'max_upload_size_image', value: mbToBytes(formData.max_upload_size_image)},
                {key: 'auto_transcode', value: formData.auto_transcode},
                {key: 'transcode_method', value: formData.transcode_method},
                {key: 'allowed_video_formats', value: formData.allowed_video_formats},
                {key: 'allowed_image_formats', value: formData.allowed_image_formats},
                {key: 'max_video_duration', value: formData.max_video_duration},
                {key: 'sprite_frame_interval', value: formData.sprite_frame_interval},
                {key: 'sprite_columns', value: formData.sprite_columns},
                {key: 'sprite_frame_width', value: formData.sprite_frame_width},
                {key: 'sprite_frame_height', value: formData.sprite_frame_height},
                {key: 'sprite_max_frames', value: formData.sprite_max_frames},
                {key: 'thumbnail_quality', value: formData.thumbnail_quality},
                {key: 'thumbnail_resolution', value: formData.thumbnail_resolution},
                {key: 'thumbnail_position', value: formData.thumbnail_position},
                {key: 'auto_approve', value: formData.auto_approve},
                {key: 'require_review', value: formData.require_review},
                {key: 'smtp_host', value: formData.smtp_host},
                {key: 'smtp_port', value: formData.smtp_port},
                {key: 'smtp_user', value: formData.smtp_user},
                {key: 'smtp_password', value: formData.smtp_password},
                {key: 'smtp_sender_name', value: formData.smtp_sender_name},
                {key: 'smtp_use_tls', value: formData.smtp_use_tls},
                {key: 'min_password_length', value: formData.min_password_length},
                {key: 'require_email_verification', value: formData.require_email_verification},
                {key: 'api_rate_limit', value: formData.api_rate_limit},
                {key: 'module_articles', value: String(formData.module_articles)},
                {key: 'module_videos', value: String(formData.module_videos)},
                {key: 'module_music', value: String(formData.module_music)},
                {key: 'homepage_layout', value: formData.homepage_layout},
            ];
            await settingsApi.update({settings});
            setMessage({type: 'success', text: t('settings.saveSuccess')});
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to save settings:', error);
            setMessage({type: 'error', text: t('settings.saveFailed')});
            setTimeout(() => setMessage(null), 3000);
        } finally {
            setSaving(false);
        }
    };

    const showS3Config = formData.storage_type === 's3' || formData.storage_type === 'hybrid';

    if (loading) {
        return (
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="space-y-2">
                        <Skeleton className="h-8 w-48"/>
                        <Skeleton className="h-4 w-64"/>
                    </div>
                    <div className="flex gap-2">
                        <Skeleton className="h-10 w-28"/>
                        <Skeleton className="h-10 w-24"/>
                    </div>
                </div>
                <Skeleton className="h-12 w-[800px]"/>
                <div className="space-y-6">
                    {[1, 2].map(i => (
                        <Card key={i}>
                            <CardHeader>
                                <Skeleton className="h-6 w-48"/>
                                <Skeleton className="h-4 w-64"/>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    {[1, 2, 3, 4].map(j => (
                                        <div key={j} className="space-y-2">
                                            <Skeleton className="h-4 w-20"/>
                                            <Skeleton className="h-10 w-full"/>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6 p-4 md:p-6">
            {message && (
                <div className={`flex items-center gap-3 p-4 rounded-lg border ${
                    message.type === 'success'
                        ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200'
                        : 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200'
                }`}>
                    {message.type === 'success' ? (
                        <CheckCircle className="w-5 h-5 flex-shrink-0"/>
                    ) : (
                        <AlertCircle className="w-5 h-5 flex-shrink-0"/>
                    )}
                    <span className="text-sm font-medium">{message.text}</span>
                </div>
            )}

            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">{t('settings.title')}</h1>
                    <p className="text-muted-foreground">{t('settings.desc')}</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" onClick={fetchSystemInfo}>
                        <RefreshCw className="mr-2 h-4 w-4"/>
                        {t('settings.refresh')}
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                        {saving ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                        ) : (
                            <Save className="mr-2 h-4 w-4"/>
                        )}
                        {saving ? t('settings.saving') : t('settings.save')}
                    </Button>
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab}>
                <TabsList className="grid w-full grid-cols-4 sm:grid-cols-7">
                    <TabsTrigger value="general">{t('settings.tabGeneral')}</TabsTrigger>
                    <TabsTrigger value="storage">{t('settings.tabStorage')}</TabsTrigger>
                    <TabsTrigger value="media">{t('settings.tabMedia')}</TabsTrigger>
                    <TabsTrigger value="email">{t('settings.tabEmail')}</TabsTrigger>
                    <TabsTrigger value="security">{t('settings.tabSecurity')}</TabsTrigger>
                    <TabsTrigger value="modules">{t('settings.tabModules')}</TabsTrigger>
                    <TabsTrigger value="system">{t('settings.tabSystem')}</TabsTrigger>
                </TabsList>

                <TabsContent value="general" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Globe className="h-5 w-5"/>
                                {t('settings.siteInfo')}
                            </CardTitle>
                            <CardDescription>{t('settings.siteInfoDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteName')}</label>
                                    <Input
                                        value={formData.site_name}
                                        onChange={(e) => handleInputChange('site_name', e.target.value)}
                                        placeholder={t('settings.enterSiteName')}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteDesc')}</label>
                                    <Input
                                        value={formData.site_description}
                                        onChange={(e) => handleInputChange('site_description', e.target.value)}
                                        placeholder={t('settings.enterSiteDesc')}
                                    />
                                </div>
                            </div>

                            <Separator/>

                            <div className="space-y-3">
                                <label className="text-sm font-medium">{t('settings.baseUrls')}</label>
                                <p className="text-xs text-muted-foreground">{t('settings.baseUrlsDesc')}</p>
                                <div className="space-y-2">
                                    {formData.base_urls.map((url, index) => (
                                        <div key={index} className="flex items-center gap-2">
                                            <div className="flex-1 relative">
                                                <Input
                                                    value={url}
                                                    onChange={(e) => handleBaseUrlChange(index, e.target.value)}
                                                    placeholder="https://example.com"
                                                />
                                                {index === 0 && (
                                                    <Badge className="absolute right-2 top-1/2 -translate-y-1/2 text-xs"
                                                           variant="secondary">
                                                        {t('settings.primaryUrlBadge')}
                                                    </Badge>
                                                )}
                                            </div>
                                            {formData.base_urls.length > 1 && (
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    onClick={() => handleRemoveBaseUrl(index)}
                                                    className="flex-shrink-0"
                                                >
                                                    <X className="h-4 w-4"/>
                                                </Button>
                                            )}
                                        </div>
                                    ))}
                                    <Button variant="outline" size="sm" onClick={handleAddBaseUrl}>
                                        <Plus className="mr-1 h-3 w-3"/>
                                        {t('settings.addUrl')}
                                    </Button>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">{t('settings.primaryUrl')}</label>
                                <Input
                                    value={formData.primary_url}
                                    onChange={(e) => handleInputChange('primary_url', e.target.value)}
                                    placeholder="https://example.com"
                                />
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Palette className="h-5 w-5"/>
                                {t('settings.appearance')}
                            </CardTitle>
                            <CardDescription>{t('settings.appearanceDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <ThemeSwitcher/>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="storage" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <HardDrive className="h-5 w-5"/>
                                {t('settings.localStorage')}
                            </CardTitle>
                            <CardDescription>{t('settings.localStorageDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium">{t('settings.storageBasePath')}</label>
                                <Input
                                    value={formData.storage_base_path}
                                    onChange={(e) => handleInputChange('storage_base_path', e.target.value)}
                                    placeholder="/var/media"
                                />
                                <p className="text-xs text-muted-foreground">{t('settings.storageRestartNote')}</p>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-muted-foreground">
                                    {t('settings.subdirectoryPreview')}
                                </label>
                                <div className="flex flex-wrap gap-2">
                                    {['originals/', 'temp/', 'thumbnails/', 'hls/', 'previews/', 'sprites/'].map(dir => (
                                        <code key={dir}
                                              className="px-2 py-1 text-xs rounded bg-muted font-mono">
                                            {formData.storage_base_path ? `${formData.storage_base_path}/${dir}` : dir}
                                        </code>
                                    ))}
                                </div>
                            </div>

                            <Separator/>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">{t('settings.storageType')}</label>
                                <select
                                    className="w-full px-3 py-2 border rounded-md bg-background"
                                    value={formData.storage_type}
                                    onChange={(e) => handleInputChange('storage_type', e.target.value)}
                                >
                                    <option value="local">{t('settings.storageLocal')}</option>
                                    <option value="s3" disabled={!storageCaps.s3_available}>
                                        {t('settings.storageS3')}
                                        {!storageCaps.s3_available ? ` (${t('settings.s3NotAvailable')})` : ''}
                                    </option>
                                    <option value="hybrid" disabled={!storageCaps.hybrid_available}>
                                        {t('settings.storageHybrid')}
                                        {!storageCaps.hybrid_available ? ` (${t('settings.s3NotAvailable')})` : ''}
                                    </option>
                                </select>
                            </div>
                        </CardContent>
                    </Card>

                    {showS3Config && (
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Server className="h-5 w-5"/>
                                    {t('settings.s3Config')}
                                </CardTitle>
                                <CardDescription>{t('settings.s3ConfigDesc')}</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div className="space-y-2 md:col-span-2">
                                        <label className="text-sm font-medium">{t('settings.endpointUrl')}</label>
                                        <Input
                                            value={formData.s3_endpoint}
                                            onChange={(e) => handleInputChange('s3_endpoint', e.target.value)}
                                            placeholder="https://s3.amazonaws.com"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-medium">{t('settings.s3Region')}</label>
                                        <Input
                                            value={formData.s3_region}
                                            onChange={(e) => handleInputChange('s3_region', e.target.value)}
                                            placeholder="us-east-1"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-medium">{t('settings.bucketName')}</label>
                                        <Input
                                            value={formData.s3_bucket}
                                            onChange={(e) => handleInputChange('s3_bucket', e.target.value)}
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-medium">{t('settings.accessKey')}</label>
                                        <Input
                                            type="password"
                                            value={formData.s3_access_key}
                                            onChange={(e) => handleInputChange('s3_access_key', e.target.value)}
                                            placeholder="••••••••"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-medium">{t('settings.secretKey')}</label>
                                        <Input
                                            type="password"
                                            value={formData.s3_secret_key}
                                            onChange={(e) => handleInputChange('s3_secret_key', e.target.value)}
                                            placeholder="••••••••"
                                        />
                                    </div>
                                    <div className="flex items-center justify-between p-3 rounded-lg border md:col-span-2">
                                        <div className="space-y-0.5">
                                            <Label className="text-sm font-medium">{t('settings.s3UsePathStyle')}</Label>
                                            <p className="text-xs text-muted-foreground">
                                                {t('settings.s3UsePathStyleDesc')}
                                            </p>
                                        </div>
                                        <Switch
                                            checked={formData.s3_use_path_style === 'true'}
                                            onCheckedChange={(checked: boolean) =>
                                                handleInputChange('s3_use_path_style', String(checked))
                                            }
                                        />
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </TabsContent>

                <TabsContent value="media" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <SettingsIcon className="h-5 w-5"/>
                                {t('settings.transcodeSettings')}
                            </CardTitle>
                            <CardDescription>{t('settings.transcodeSettingsDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between p-3 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-sm font-medium">{t('settings.autoTranscode')}</Label>
                                    <p className="text-xs text-muted-foreground">
                                        {t('settings.autoTranscodeDesc')}
                                    </p>
                                </div>
                                <Switch
                                    checked={formData.auto_transcode === 'true'}
                                    onCheckedChange={(checked: boolean) =>
                                        handleInputChange('auto_transcode', String(checked))
                                    }
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">{t('settings.transcodeMethod')}</label>
                                <select
                                    className="w-full px-3 py-2 border rounded-md bg-background"
                                    value={formData.transcode_method}
                                    onChange={(e) => handleInputChange('transcode_method', e.target.value)}
                                >
                                    <option value="ffmpeg">FFmpeg</option>
                                </select>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>{t('settings.uploadLimits')}</CardTitle>
                            <CardDescription>{t('settings.uploadLimitsDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxUploadSizeVideo')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_upload_size_video}
                                        onChange={(e) => handleInputChange('max_upload_size_video', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxUploadSizeImage')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_upload_size_image}
                                        onChange={(e) => handleInputChange('max_upload_size_image', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxVideoDuration')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_video_duration}
                                        onChange={(e) => handleInputChange('max_video_duration', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.videoFormats')}</label>
                                    <Input
                                        value={formData.allowed_video_formats}
                                        onChange={(e) => handleInputChange('allowed_video_formats', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2 md:col-span-2">
                                    <label className="text-sm font-medium">{t('settings.imageFormats')}</label>
                                    <Input
                                        value={formData.allowed_image_formats}
                                        onChange={(e) => handleInputChange('allowed_image_formats', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>{t('settings.thumbnailSprite')}</CardTitle>
                            <CardDescription>{t('settings.thumbnailSpriteDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.thumbnailQuality')}</label>
                                    <Input
                                        type="number"
                                        value={formData.thumbnail_quality}
                                        onChange={(e) => handleInputChange('thumbnail_quality', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.thumbnailResolution')}</label>
                                    <Input
                                        value={formData.thumbnail_resolution}
                                        onChange={(e) => handleInputChange('thumbnail_resolution', e.target.value)}
                                        placeholder="320x180"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.thumbnailPosition')}</label>
                                    <Input
                                        value={formData.thumbnail_position}
                                        onChange={(e) => handleInputChange('thumbnail_position', e.target.value)}
                                        placeholder="00:00:01"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.spriteFrameInterval')}</label>
                                    <Input
                                        type="number"
                                        value={formData.sprite_frame_interval}
                                        onChange={(e) => handleInputChange('sprite_frame_interval', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.spriteColumns')}</label>
                                    <Input
                                        type="number"
                                        value={formData.sprite_columns}
                                        onChange={(e) => handleInputChange('sprite_columns', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.spriteFrameWidth')}</label>
                                    <Input
                                        type="number"
                                        value={formData.sprite_frame_width}
                                        onChange={(e) => handleInputChange('sprite_frame_width', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.spriteFrameHeight')}</label>
                                    <Input
                                        type="number"
                                        value={formData.sprite_frame_height}
                                        onChange={(e) => handleInputChange('sprite_frame_height', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.spriteMaxFrames')}</label>
                                    <Input
                                        type="number"
                                        value={formData.sprite_max_frames}
                                        onChange={(e) => handleInputChange('sprite_max_frames', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="email" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Mail className="h-5 w-5"/>
                                {t('settings.smtp')}
                            </CardTitle>
                            <CardDescription>{t('settings.smtpDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center gap-2 p-3 rounded-lg border">
                                <span className="text-sm font-medium">{t('settings.emailStatus')}:</span>
                                {emailStatus.configured ? (
                                    <Badge variant="default"
                                           className="bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                                        <CheckCircle className="w-3 h-3 mr-1"/>
                                        {t('settings.emailConfigured')}
                                    </Badge>
                                ) : (
                                    <Badge variant="secondary"
                                           className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">
                                        <AlertCircle className="w-3 h-3 mr-1"/>
                                        {t('settings.emailNotConfigured')}
                                    </Badge>
                                )}
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.smtpServer')}</label>
                                    <Input
                                        value={formData.smtp_host}
                                        onChange={(e) => handleInputChange('smtp_host', e.target.value)}
                                        placeholder="smtp.example.com"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.smtpPort')}</label>
                                    <Input
                                        type="number"
                                        value={formData.smtp_port}
                                        onChange={(e) => handleInputChange('smtp_port', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.username')}</label>
                                    <Input
                                        value={formData.smtp_user}
                                        onChange={(e) => handleInputChange('smtp_user', e.target.value)}
                                        placeholder="noreply@example.com"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.password')}</label>
                                    <Input
                                        type="password"
                                        value={formData.smtp_password}
                                        onChange={(e) => handleInputChange('smtp_password', e.target.value)}
                                        placeholder="••••••••"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.senderName')}</label>
                                    <Input
                                        value={formData.smtp_sender_name}
                                        onChange={(e) => handleInputChange('smtp_sender_name', e.target.value)}
                                    />
                                </div>
                                <div className="flex items-center justify-between p-3 rounded-lg border">
                                    <div className="space-y-0.5">
                                        <Label className="text-sm font-medium">{t('settings.useTls')}</Label>
                                    </div>
                                    <Switch
                                        checked={formData.smtp_use_tls === 'true'}
                                        onCheckedChange={(checked: boolean) =>
                                            handleInputChange('smtp_use_tls', String(checked))
                                        }
                                    />
                                </div>
                            </div>

                            <Separator/>

                            <div className="space-y-3">
                                <label className="text-sm font-medium">{t('settings.sendTest')}</label>
                                <div className="flex gap-2">
                                    <Input
                                        value={emailTestTo}
                                        onChange={(e) => setEmailTestTo(e.target.value)}
                                        placeholder="test@example.com"
                                        className="flex-1"
                                    />
                                    <Button
                                        variant="outline"
                                        onClick={handleEmailTest}
                                        disabled={emailTestSending || !emailTestTo}
                                    >
                                        {emailTestSending ? (
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                                        ) : (
                                            <Send className="mr-2 h-4 w-4"/>
                                        )}
                                        {t('settings.sendTest')}
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="security" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Shield className="h-5 w-5"/>
                                {t('settings.auth')}
                            </CardTitle>
                            <CardDescription>{t('settings.authDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between p-3 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-sm font-medium">{t('settings.enableRegister')}</Label>
                                </div>
                                <Switch
                                    checked={formData.allow_registration === 'true'}
                                    onCheckedChange={(checked: boolean) =>
                                        handleInputChange('allow_registration', String(checked))
                                    }
                                />
                            </div>
                            <div className="flex items-center justify-between p-3 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-sm font-medium">{t('settings.requireEmailVerify')}</Label>
                                </div>
                                <Switch
                                    checked={formData.require_email_verification === 'true'}
                                    onCheckedChange={(checked: boolean) =>
                                        handleInputChange('require_email_verification', String(checked))
                                    }
                                />
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.minPasswordLen')}</label>
                                    <Input
                                        type="number"
                                        value={formData.min_password_length}
                                        onChange={(e) => handleInputChange('min_password_length', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.rateLimit')}</label>
                                    <Input
                                        type="number"
                                        value={formData.api_rate_limit}
                                        onChange={(e) => handleInputChange('api_rate_limit', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>{t('settings.reviewSettings')}</CardTitle>
                            <CardDescription>{t('settings.reviewSettingsDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between p-3 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-sm font-medium">{t('settings.autoApprove')}</Label>
                                    <p className="text-xs text-muted-foreground">
                                        {t('settings.autoApproveDesc')}
                                    </p>
                                </div>
                                <Switch
                                    checked={formData.auto_approve === 'true'}
                                    onCheckedChange={(checked: boolean) => {
                                        handleInputChange('auto_approve', String(checked));
                                        handleInputChange('require_review', String(!checked));
                                    }}
                                />
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="modules" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Blocks className="h-5 w-5"/>
                                {t('settings.contentModules')}
                            </CardTitle>
                            <CardDescription>
                                {t('settings.contentModulesDesc')}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="flex items-center justify-between p-4 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-base font-medium">{t('settings.moduleArticles')}</Label>
                                    <p className="text-sm text-muted-foreground">
                                        {t('settings.moduleArticlesDesc')}
                                    </p>
                                </div>
                                <Switch
                                    checked={formData.module_articles}
                                    onCheckedChange={(checked: boolean) => setFormData(prev => ({
                                        ...prev,
                                        module_articles: checked
                                    }))}
                                />
                            </div>
                            <div className="flex items-center justify-between p-4 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-base font-medium">{t('settings.moduleVideos')}</Label>
                                    <p className="text-sm text-muted-foreground">
                                        {t('settings.moduleVideosDesc')}
                                    </p>
                                </div>
                                <Switch
                                    checked={formData.module_videos}
                                    onCheckedChange={(checked: boolean) => setFormData(prev => ({
                                        ...prev,
                                        module_videos: checked
                                    }))}
                                />
                            </div>
                            <div className="flex items-center justify-between p-4 rounded-lg border">
                                <div className="space-y-0.5">
                                    <Label className="text-base font-medium">{t('settings.moduleMusic')}</Label>
                                    <p className="text-sm text-muted-foreground">
                                        {t('settings.moduleMusicDesc')}
                                    </p>
                                </div>
                                <Switch
                                    checked={formData.module_music}
                                    onCheckedChange={(checked: boolean) => setFormData(prev => ({
                                        ...prev,
                                        module_music: checked
                                    }))}
                                />
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <LayoutGrid className="h-5 w-5"/>
                                {t('settings.homepageLayout')}
                            </CardTitle>
                            <CardDescription>
                                {t('settings.homepageLayoutDesc')}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                                {[
                                    {value: 'auto', label: t('settings.layoutAuto'), desc: t('settings.layoutAutoDesc')},
                                    {value: 'video', label: t('settings.layoutVideo'), desc: t('settings.layoutVideoDesc')},
                                    {
                                        value: 'article',
                                        label: t('settings.layoutArticle'),
                                        desc: t('settings.layoutArticleDesc')
                                    },
                                    {value: 'mixed', label: t('settings.layoutMixed'), desc: t('settings.layoutMixedDesc')},
                                    {value: 'doc', label: t('settings.layoutDoc'), desc: t('settings.layoutDocDesc')},
                                    {
                                        value: 'welcome',
                                        label: t('settings.layoutWelcome'),
                                        desc: t('settings.layoutWelcomeDesc')
                                    },
                                ].map((option) => (
                                    <button
                                        key={option.value}
                                        type="button"
                                        onClick={() => setFormData(prev => ({...prev, homepage_layout: option.value}))}
                                        className={`text-left p-4 rounded-lg border-2 transition-colors ${
                                            formData.homepage_layout === option.value
                                                ? 'border-primary bg-primary/5'
                                                : 'border-muted hover:border-muted-foreground/30'
                                        }`}
                                    >
                                        <div className="font-medium text-sm">{option.label}</div>
                                        <div className="text-xs text-muted-foreground mt-1">{option.desc}</div>
                                    </button>
                                ))}
                            </div>

                            {formData.homepage_layout === 'auto' && (
                                <div className="p-3 rounded-lg bg-muted/50 text-sm text-muted-foreground">
                                    {t('settings.autoLayoutPreview')} <strong>{
                                    !formData.module_articles && !formData.module_videos && !formData.module_music
                                        ? t('settings.layoutWelcome') :
                                        formData.module_videos && !formData.module_articles
                                            ? t('settings.layoutVideo') :
                                            formData.module_articles && !formData.module_videos
                                                ? t('settings.layoutDoc') :
                                                formData.module_articles && formData.module_videos
                                                    ? t('settings.layoutMixed') :
                                                    t('settings.layoutWelcome')
                                }</strong>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="system" className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Server className="h-5 w-5"/>
                                    {t('settings.serverInfo')}
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                {systemInfo ? (
                                    <div className="space-y-3">
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.version')}</span>
                                            <span className="font-medium">{systemInfo.version || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.goVersion')}</span>
                                            <span className="font-medium">{systemInfo.goVersion || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.database')}</span>
                                            <span className="font-medium">{systemInfo.database || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.os')}</span>
                                            <span className="font-medium">{systemInfo.os || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.uptime')}</span>
                                            <span className="font-medium">{systemInfo.uptime || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.numCPU')}</span>
                                            <span className="font-medium">{systemInfo.numCPU ?? '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.numGoroutine')}</span>
                                            <span className="font-medium">{systemInfo.numGoroutine ?? '-'}</span>
                                        </div>
                                    </div>
                                ) : (
                                    <div className="space-y-3">
                                        {[1, 2, 3, 4, 5, 6, 7].map(i => (
                                            <div key={i} className="flex justify-between">
                                                <Skeleton className="h-4 w-24"/>
                                                <Skeleton className="h-4 w-16"/>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Database className="h-5 w-5"/>
                                    {t('settings.resourceUsage')}
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                {systemInfo ? (
                                    <div className="space-y-3">
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.totalMemory')}</span>
                                            <span className="font-medium">{systemInfo.totalMemory || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.usedMemory')}</span>
                                            <span className="font-medium">{systemInfo.usedMemory || '-'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">{t('settings.cpuUsage')}</span>
                                            <span className="font-medium">{systemInfo.cpuUsage || '-'}</span>
                                        </div>
                                        <div className="mt-4 space-y-2">
                                            <div className="flex justify-between text-sm">
                                                <span>{t('settings.memoryUsage')}</span>
                                                <span>{systemInfo.memoryUsage?.toFixed(2)}%</span>
                                            </div>
                                            <div className="h-2 bg-muted rounded-full overflow-hidden">
                                                <div
                                                    className="h-full bg-blue-600 rounded-full transition-all duration-500"
                                                    style={{width: `${systemInfo.memoryUsage || 0}%`}}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                ) : (
                                    <div className="space-y-3">
                                        {[1, 2, 3].map(i => (
                                            <div key={i} className="flex justify-between">
                                                <Skeleton className="h-4 w-24"/>
                                                <Skeleton className="h-4 w-16"/>
                                            </div>
                                        ))}
                                        <div className="mt-4 space-y-2">
                                            <Skeleton className="h-4 w-16"/>
                                            <Skeleton className="h-2 w-full"/>
                                        </div>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default Settings;
