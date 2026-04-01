import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Separator} from '@/components/ui/separator';
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
    RefreshCw
} from 'lucide-react';

// 模拟系统数据
const systemInfo = {
    version: '1.0.0',
    goVersion: 'Go 1.21',
    database: 'PostgreSQL 15',
    os: 'Linux 5.15',
    uptime: '15天 8小时 23分钟',
    totalMemory: '16 GB',
    usedMemory: '6.2 GB',
    cpuUsage: '23%',
};

const Settings: React.FC = () => {
    const {t} = useTranslation();
    const [activeTab, setActiveTab] = useState('general');

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">{t('settings.title')}</h1>
                    <p className="text-muted-foreground">{t('settings.desc')}</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline">
                        <RefreshCw className="mr-2 h-4 w-4"/>
                        {t('settings.restart')}
                    </Button>
                    <Button>
                        <Save className="mr-2 h-4 w-4"/>
                        {t('settings.save')}
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
                                    <Input defaultValue="OrigCMS"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteDesc')}</label>
                                    <Input defaultValue="现代媒体内容管理系统"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.siteUrl')}</label>
                                    <Input defaultValue="https://example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.timezone')}</label>
                                    <Input defaultValue="Asia/Shanghai"/>
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
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="light">{t('settings.light')}</option>
                                        <option value="dark">{t('settings.dark')}</option>
                                        <option value="system">{t('settings.system')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.themeColor')}</label>
                                    <Input type="color" defaultValue="#3b82f6" className="h-10"/>
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
                                    <Input defaultValue="/var/media/uploads"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.encodeDir')}</label>
                                    <Input defaultValue="/var/media/encoded"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.thumbDir')}</label>
                                    <Input defaultValue="/var/media/thumbnails"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxFileSize')}</label>
                                    <Input type="number" defaultValue="2048"/>
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
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="local">本地存储</option>
                                        <option value="s3">Amazon S3</option>
                                        <option value="minio">MinIO</option>
                                        <option value="oss">Aliyun OSS</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.bucketName')}</label>
                                    <Input defaultValue="origcms-media"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.accessKey')}</label>
                                    <Input type="password" defaultValue="AKIAIOSFODNN7EXAMPLE"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.secretKey')}</label>
                                    <Input type="password" defaultValue="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"/>
                                </div>
                                <div className="space-y-2 md:col-span-2">
                                    <label className="text-sm font-medium">{t('settings.endpointUrl')}</label>
                                    <Input defaultValue="https://s3.amazonaws.com"/>
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
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.transcodeMethod')}</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
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
                                    <Input defaultValue="mp4, webm, mkv, avi, mov"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.imageFormats')}</label>
                                    <Input defaultValue="jpg, png, gif, webp"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxDuration')}</label>
                                    <Input type="number" defaultValue="120"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.maxFileSize2')}</label>
                                    <Input type="number" defaultValue="2048"/>
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
                                    <Input defaultValue="smtp.example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.smtpPort')}</label>
                                    <Input type="number" defaultValue="587"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.username')}</label>
                                    <Input defaultValue="noreply@example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.password')}</label>
                                    <Input type="password" defaultValue="password123"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.senderName')}</label>
                                    <Input defaultValue="OrigCMS"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.useTls')}</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
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
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.requireEmailVerify')}</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.minPasswordLen')}</label>
                                    <Input type="number" defaultValue="8"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.jwtExpiry')}</label>
                                    <Input type="number" defaultValue="7"/>
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
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">{t('settings.enable')}</option>
                                        <option value="false">{t('settings.disable')}</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">{t('settings.rateLimit')}</label>
                                    <Input type="number" defaultValue="60"/>
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
                                <div className="space-y-3">
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.version')}</span>
                                        <span className="font-medium">{systemInfo.version}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.goVersion')}</span>
                                        <span className="font-medium">{systemInfo.goVersion}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.database')}</span>
                                        <span className="font-medium">{systemInfo.database}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.os')}</span>
                                        <span className="font-medium">{systemInfo.os}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.uptime')}</span>
                                        <span className="font-medium">{systemInfo.uptime}</span>
                                    </div>
                                </div>
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
                                <div className="space-y-3">
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.totalMemory')}</span>
                                        <span className="font-medium">{systemInfo.totalMemory}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.usedMemory')}</span>
                                        <span className="font-medium">{systemInfo.usedMemory}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">{t('settings.cpuUsage')}</span>
                                        <span className="font-medium">{systemInfo.cpuUsage}</span>
                                    </div>
                                </div>
                                <div className="mt-4 space-y-2">
                                    <div className="flex justify-between text-sm">
                                        <span>{t('settings.memoryUsage')}</span>
                                        <span>38.75%</span>
                                    </div>
                                    <div className="h-2 bg-muted rounded-full overflow-hidden">
                                        <div className="h-full bg-blue-600 rounded-full" style={{width: '38.75%'}}/>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default Settings;