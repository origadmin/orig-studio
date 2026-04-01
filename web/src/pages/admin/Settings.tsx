import React, {useState} from 'react';
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
    const [activeTab, setActiveTab] = useState('general');

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">系统设置</h1>
                    <p className="text-muted-foreground">配置和管理系统各项参数</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline">
                        <RefreshCw className="mr-2 h-4 w-4"/>
                        重启服务
                    </Button>
                    <Button>
                        <Save className="mr-2 h-4 w-4"/>
                        保存配置
                    </Button>
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab}>
                <TabsList className="grid w-full grid-cols-6 lg:w-[800px]">
                    <TabsTrigger value="general">通用</TabsTrigger>
                    <TabsTrigger value="storage">存储</TabsTrigger>
                    <TabsTrigger value="media">媒体</TabsTrigger>
                    <TabsTrigger value="email">邮件</TabsTrigger>
                    <TabsTrigger value="security">安全</TabsTrigger>
                    <TabsTrigger value="system">系统</TabsTrigger>
                </TabsList>

                {/* 通用设置 */}
                <TabsContent value="general" className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Globe className="h-5 w-5"/>
                                站点信息
                            </CardTitle>
                            <CardDescription>配置站点的基本信息和名称</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">站点名称</label>
                                    <Input defaultValue="OrigCMS"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">站点描述</label>
                                    <Input defaultValue="现代媒体内容管理系统"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">站点URL</label>
                                    <Input defaultValue="https://example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">时区</label>
                                    <Input defaultValue="Asia/Shanghai"/>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Palette className="h-5 w-5"/>
                                外观
                            </CardTitle>
                            <CardDescription>自定义站点的外观和主题</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">默认主题</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="light">浅色模式</option>
                                        <option value="dark">深色模式</option>
                                        <option value="system">跟随系统</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">主题色</label>
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
                                本地存储
                            </CardTitle>
                            <CardDescription>配置本地文件存储路径和限制</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">上传目录</label>
                                    <Input defaultValue="/var/media/uploads"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">转码目录</label>
                                    <Input defaultValue="/var/media/encoded"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">缩略图目录</label>
                                    <Input defaultValue="/var/media/thumbnails"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">单文件大小限制 (MB)</label>
                                    <Input type="number" defaultValue="2048"/>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Server className="h-5 w-5"/>
                                对象存储
                            </CardTitle>
                            <CardDescription>配置 S3/MinIO 等对象存储服务</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">存储类型</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="local">本地存储</option>
                                        <option value="s3">Amazon S3</option>
                                        <option value="minio">MinIO</option>
                                        <option value="oss">阿里云 OSS</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">存储桶名称</label>
                                    <Input defaultValue="origcms-media"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">访问密钥</label>
                                    <Input type="password" defaultValue="AKIAIOSFODNN7EXAMPLE"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">秘密密钥</label>
                                    <Input type="password" defaultValue="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"/>
                                </div>
                                <div className="space-y-2 md:col-span-2">
                                    <label className="text-sm font-medium">端点 URL</label>
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
                                转码设置
                            </CardTitle>
                            <CardDescription>配置视频转码参数和分辨率选项</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">启用自动转码</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">启用</option>
                                        <option value="false">禁用</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">转码方式</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="ffmpeg">FFmpeg</option>
                                        <option value="handbrake">HandBrake</option>
                                    </select>
                                </div>
                            </div>
                            <Separator/>
                            <div className="space-y-3">
                                <label className="text-sm font-medium">输出分辨率</label>
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
                            <CardTitle>上传限制</CardTitle>
                            <CardDescription>限制用户上传的格式和大小</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">允许的视频格式</label>
                                    <Input defaultValue="mp4, webm, mkv, avi, mov"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">允许的图片格式</label>
                                    <Input defaultValue="jpg, png, gif, webp"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">最大视频时长 (分钟)</label>
                                    <Input type="number" defaultValue="120"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">最大文件大小 (MB)</label>
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
                                SMTP 设置
                            </CardTitle>
                            <CardDescription>配置邮件发送服务</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">SMTP 服务器</label>
                                    <Input defaultValue="smtp.example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">SMTP 端口</label>
                                    <Input type="number" defaultValue="587"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">用户名</label>
                                    <Input defaultValue="noreply@example.com"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">密码</label>
                                    <Input type="password" defaultValue="password123"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">发件人名称</label>
                                    <Input defaultValue="OrigCMS"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">使用 TLS</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">是</option>
                                        <option value="false">否</option>
                                    </select>
                                </div>
                            </div>
                            <div className="flex justify-end">
                                <Button variant="outline">发送测试邮件</Button>
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
                                认证与授权
                            </CardTitle>
                            <CardDescription>配置用户认证和安全策略</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">启用用户注册</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">启用</option>
                                        <option value="false">禁用</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">注册需要邮箱验证</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">启用</option>
                                        <option value="false">禁用</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">最小密码长度</label>
                                    <Input type="number" defaultValue="8"/>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">JWT 过期时间 (天)</label>
                                    <Input type="number" defaultValue="7"/>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>API 访问</CardTitle>
                            <CardDescription>配置 API 访问限制</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">启用 REST API</label>
                                    <select className="w-full px-3 py-2 border rounded-md bg-background">
                                        <option value="true">启用</option>
                                        <option value="false">禁用</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">API 速率限制 (请求/分钟)</label>
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
                                    服务器信息
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-3">
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">版本</span>
                                        <span className="font-medium">{systemInfo.version}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">Go 版本</span>
                                        <span className="font-medium">{systemInfo.goVersion}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">数据库</span>
                                        <span className="font-medium">{systemInfo.database}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">操作系统</span>
                                        <span className="font-medium">{systemInfo.os}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">运行时间</span>
                                        <span className="font-medium">{systemInfo.uptime}</span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Database className="h-5 w-5"/>
                                    资源使用
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-3">
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">总内存</span>
                                        <span className="font-medium">{systemInfo.totalMemory}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">已用内存</span>
                                        <span className="font-medium">{systemInfo.usedMemory}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">CPU 使用率</span>
                                        <span className="font-medium">{systemInfo.cpuUsage}</span>
                                    </div>
                                </div>
                                <div className="mt-4 space-y-2">
                                    <div className="flex justify-between text-sm">
                                        <span>内存使用</span>
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