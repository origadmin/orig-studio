import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Separator} from '@/components/ui/separator';
import {Skeleton} from '@/components/ui/skeleton';
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
    AlertCircle
} from 'lucide-react';
import {settingsApi, type SystemSettings} from '@/lib/api/system';

interface FormData {
    site_name: string;
    site_description: string;
    site_url: string;
    timezone: string;
    default_theme: string;
    theme_color: string;
    upload_dir: string;
    encode_dir: string;
    thumb_dir: string;
    max_file_size: string;
    storage_type: string;
    bucket_name: string;
    access_key: string;
    secret_key: string;
    endpoint_url: string;
    auto_transcode: string;
    transcode_method: string;
    video_formats: string;
    image_formats: string;
    max_duration: string;
    max_file_size2: string;
    smtp_server: string;
    smtp_port: string;
    smtp_username: string;
    smtp_password: string;
    sender_name: string;
    use_tls: string;
    enable_register: string;
    require_email_verify: string;
    min_password_len: string;
    jwt_expiry: string;
    enable_rest_api: string;
    rate_limit: string;
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
}

const defaultFormData: FormData = {
    site_name: '',
    site_description: '',
    site_url: '',
    timezone: 'Asia/Shanghai',
    default_theme: 'system',
    theme_color: '#3b82f6',
    upload_dir: '/var/media/uploads',
    encode_dir: '/var/media/encoded',
    thumb_dir: '/var/media/thumbnails',
    max_file_size: '2048',
    storage_type: 'local',
    bucket_name: '',
    access_key: '',
    secret_key: '',
    endpoint_url: '',
    auto_transcode: 'true',
    transcode_method: 'ffmpeg',
    video_formats: 'mp4, webm, mkv, avi, mov',
    image_formats: 'jpg, png, gif, webp',
    max_duration: '120',
    max_file_size2: '2048',
    smtp_server: '',
    smtp_port: '587',
    smtp_username: '',
    smtp_password: '',
    sender_name: '',
    use_tls: 'true',
    enable_register: 'true',
    require_email_verify: 'false',
    min_password_len: '8',
    jwt_expiry: '7',
    enable_rest_api: 'true',
    rate_limit: '60',
};

const Settings: React.FC = () => {
    const {t} = useTranslation();
    const [activeTab, setActiveTab] = useState('general');
    const [formData, setFormData] = useState<FormData>(defaultFormData);
    const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [initialLoad, setInitialLoad] = useState(true);
    const [message, setMessage] = useState<{type: 'success' | 'error', text: string} | null>(null);

    useEffect(() => {
        fetchSettings();
        fetchSystemInfo();
    }, []);

    const fetchSettings = async () => {
        try {
            const settings = await settingsApi.get();
            if (settings) {
                setFormData(prev => ({
                    ...prev,
                    site_name: settings.site_name || prev.site_name,
                    site_description: settings.site_description || prev.site_description,
                }));
            }
        } catch (error) {
            console.error('Failed to fetch settings:', error);
            setMessage({type: 'error', text: t('settings.loadFailed') || 'Failed to load settings'});
            setTimeout(() => setMessage(null), 3000);
        } finally {
            setLoading(false);
            setTimeout(() => setInitialLoad(false), 100);
        }
    };

    const fetchSystemInfo = async () => {
        try {
            const response = await fetch('/api/v1/system/info');
            if (response.ok) {
                const info = await response.json();
                if (info.data) {
                    setSystemInfo(info.data);
                }
            }
        } catch (error) {
            console.error('Failed to fetch system info:', error);
            setSystemInfo(null);
        }
    };

    const handleInputChange = (field: keyof FormData, value: string) => {
        setFormData(prev => ({...prev, [field]: value}));
    };

    const handleSave = async () => {
        try {
            setSaving(true);
            await settingsApi.update({
                site_name: formData.site_name,
                site_description: formData.site_description,
                allow_register: formData.enable_register === 'true',
                allow_upload: true,
                max_upload_size: parseInt(formData.max_file_size) * 1024 * 1024,
            });
            setMessage({type: 'success', text: t('settings.saveSuccess') || 'Settings saved successfully'});
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to save settings:', error);
            setMessage({type: 'error', text: t('settings.saveFailed') || 'Failed to save settings'});
            setTimeout(() => setMessage(null), 3000);
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="space-y-6 p-4 md:p-6">
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
            {/* Message Alert */}
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
                        {t('settings.refresh') || '刷新'}
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                        {saving ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                        ) : (
                            <Save className="mr-2 h-4 w-4"/>
                        )}
                        {saving ? (t('settings.saving') || '保存中...') : t('settings.save')}
                    </Button>
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab}>
                <TabsList className="grid w-full grid-cols-6 lg:w-[800px]">
                    <TabsTrigger value="general">{t('settings.tabGeneral')}</TabsTrigger>
                    <TabsTrigger value="storage">{t('settings.tabStorage')}</TabsTrigger>
                    <TabsTrigger value="media">{t('settings.tabMedia')}</TabsTrigger>
                    <TabsTrigger value="email">{t('settings.tabEmail')}</TabsTrigger>
                    <TabsTrigger value="security">{t('settings.tabSecurity')}</TabsTrigger>
                    <TabsTrigger value="system">{t('settings.tabSystem')}</TabsTrigger>
                </TabsList>

                {/* 通用设置 */}
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
                                        placeholder={t('settings.enterSiteName') || '请输入站点名称'}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteDesc')}</label>
                                    <Input
                                        value={formData.site_description}
                                        onChange={(e) => handleInputChange('site_description', e.target.value)}
                                        placeholder={t('settings.enterSiteDesc') || '请输入站点描述'}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteUrl')}</label>
                                    <Input
                                        value={formData.site_url}
                                        onChange={(e) => handleInputChange('site_url', e.target.value)}
                                        placeholder="https://example.com"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.timezone')}</label>
                                    <Input
                                        value={formData.timezone}
                                        onChange={(e) => handleInputChange('timezone', e.target.value)}
                                    />
                                </div>
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
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.defaultTheme')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.default_theme}
                                        onChange={(e) => handleInputChange('default_theme', e.target.value)}
                                    >
                                        <option value="light">{t('settings.light')}</option>
                                        <option value="dark">{t('settings.dark')}</option>
                                        <option value="system">{t('settings.system')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.themeColor')}</label>
                                    <Input
                                        type="color"
                                        value={formData.theme_color}
                                        onChange={(e) => handleInputChange('theme_color', e.target.value)}
                                        className="h-10"
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 存储设置 */}
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
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.uploadDir')}</label>
                                    <Input
                                        value={formData.upload_dir}
                                        onChange={(e) => handleInputChange('upload_dir', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.encodeDir')}</label>
                                    <Input
                                        value={formData.encode_dir}
                                        onChange={(e) => handleInputChange('encode_dir', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.thumbDir')}</label>
                                    <Input
                                        value={formData.thumb_dir}
                                        onChange={(e) => handleInputChange('thumb_dir', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxFileSize')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_file_size}
                                        onChange={(e) => handleInputChange('max_file_size', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Server className="h-5 w-5"/>
                                {t('settings.objectStorage')}
                            </CardTitle>
                            <CardDescription>{t('settings.objectStorageDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.storageType')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.storage_type}
                                        onChange={(e) => handleInputChange('storage_type', e.target.value)}
                                    >
                                        <option value="local">本地存储</option>
                                        <option value="s3">Amazon S3</option>
                                        <option value="minio">MinIO</option>
                                        <option value="oss">Aliyun OSS</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.bucketName')}</label>
                                    <Input
                                        value={formData.bucket_name}
                                        onChange={(e) => handleInputChange('bucket_name', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.accessKey')}</label>
                                    <Input
                                        type="password"
                                        value={formData.access_key}
                                        onChange={(e) => handleInputChange('access_key', e.target.value)}
                                        placeholder="••••••••"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.secretKey')}</label>
                                    <Input
                                        type="password"
                                        value={formData.secret_key}
                                        onChange={(e) => handleInputChange('secret_key', e.target.value)}
                                        placeholder="••••••••"
                                    />
                                </div>
                                <div className="space-y-2 md:col-span-2">
                                    <label className="text-sm font-medium">{t('settings.endpointUrl')}</label>
                                    <Input
                                        value={formData.endpoint_url}
                                        onChange={(e) => handleInputChange('endpoint_url', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 媒体设置 */}
                <TabsContent value="media" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <SettingsIcon className="h-5 w-5"/>
                                {t('settings.transcode')}
                            </CardTitle>
                            <CardDescription>{t('settings.transcodeDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.autoTranscode')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.auto_transcode}
                                        onChange={(e) => handleInputChange('auto_transcode', e.target.value)}
                                    >
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.transcodeMethod')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.transcode_method}
                                        onChange={(e) => handleInputChange('transcode_method', e.target.value)}
                                    >
                                        <option value="ffmpeg">FFmpeg</option>
                                        <option value="handbrake">HandBrake</option>
                                    </select>
                                </div>
                            </div>
                            <Separator/>
                            <div className="space-y-3">
                                <label className="text-sm font-medium">{t('settings.outputRes')}</label>
                                <div className="flex flex-wrap gap-2">
                                    {['2160p (4K)', '1080p', '720p', '480p', '360p'].map((res) => (
                                        <Badge key={res} variant="outline" className="cursor-pointer">
                                            {res}
                                        </Badge>
                                    ))}
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>{t('settings.uploadLimit')}</CardTitle>
                            <CardDescription>{t('settings.uploadLimitDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.videoFormats')}</label>
                                    <Input
                                        value={formData.video_formats}
                                        onChange={(e) => handleInputChange('video_formats', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.imageFormats')}</label>
                                    <Input
                                        value={formData.image_formats}
                                        onChange={(e) => handleInputChange('image_formats', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxDuration')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_duration}
                                        onChange={(e) => handleInputChange('max_duration', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxFileSize2')}</label>
                                    <Input
                                        type="number"
                                        value={formData.max_file_size2}
                                        onChange={(e) => handleInputChange('max_file_size2', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 邮件设置 */}
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
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.smtpServer')}</label>
                                    <Input
                                        value={formData.smtp_server}
                                        onChange={(e) => handleInputChange('smtp_server', e.target.value)}
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
                                        value={formData.smtp_username}
                                        onChange={(e) => handleInputChange('smtp_username', e.target.value)}
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
                                        value={formData.sender_name}
                                        onChange={(e) => handleInputChange('sender_name', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.useTls')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.use_tls}
                                        onChange={(e) => handleInputChange('use_tls', e.target.value)}
                                    >
                                        <option value="true">{t('settings.yes')}</option>
                                        <option value="false">{t('settings.no')}</option>
                                    </select>
                                </div>
                            </div>
                            <div className="flex justify-end">
                                <Button variant="outline">{t('settings.sendTest')}</Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 安全设置 */}
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
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.enableRegister')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.enable_register}
                                        onChange={(e) => handleInputChange('enable_register', e.target.value)}
                                    >
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.requireEmailVerify')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.require_email_verify}
                                        onChange={(e) => handleInputChange('require_email_verify', e.target.value)}
                                    >
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.minPasswordLen')}</label>
                                    <Input
                                        type="number"
                                        value={formData.min_password_len}
                                        onChange={(e) => handleInputChange('min_password_len', e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.jwtExpiry')}</label>
                                    <Input
                                        type="number"
                                        value={formData.jwt_expiry}
                                        onChange={(e) => handleInputChange('jwt_expiry', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>{t('settings.apiAccess')}</CardTitle>
                            <CardDescription>{t('settings.apiAccessDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.enableRestApi')}</label>
                                    <select
                                        className="w-full px-3 py-2 border rounded-md bg-background"
                                        value={formData.enable_rest_api}
                                        onChange={(e) => handleInputChange('enable_rest_api', e.target.value)}
                                    >
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.rateLimit')}</label>
                                    <Input
                                        type="number"
                                        value={formData.rate_limit}
                                        onChange={(e) => handleInputChange('rate_limit', e.target.value)}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 系统信息 */}
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
                                    </div>
                                ) : (
                                    <div className="space-y-3">
                                        {[1, 2, 3, 4, 5].map(i => (
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
