// AdminLayout.tsx - 管理后台布局（使用 shadcn/ui）
import {useState} from "react";
import {Link, useRouterState} from "@tanstack/react-router";
import {signOut} from "@/lib/auth";
import {useAuth} from "@/hooks/useAuth";
import {useTranslation} from "react-i18next";
import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback} from "@/components/ui/avatar";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    LayoutDashboard,
    Users,
    Film,
    FileText,
    LogOut,
    ArrowLeft,
    Tags,
    FolderTree,
    Radio,
    MessageSquare,
    PlayCircle,
    Settings,
} from "lucide-react";

export default function AdminPageLayout({children}: { children?: React.ReactNode }) {
    const {t} = useTranslation();
    const state = useRouterState();
    const {user} = useAuth();

    const menuItems = [
        {id: "dashboard", icon: LayoutDashboard, label: t('admin.dashboard'), path: "/admin"},
        {id: "media", icon: Film, label: t('admin.media'), path: "/admin/media"},
        {id: "users", icon: Users, label: t('admin.users'), path: "/admin/users"},
        {id: "categories", icon: FolderTree, label: t('admin.categories'), path: "/admin/categories"},
        {id: "channels", icon: Radio, label: t('admin.channels'), path: "/admin/channels"},
        {id: "tags", icon: Tags, label: t('admin.tags'), path: "/admin/tags"},
        {id: "comments", icon: MessageSquare, label: t('admin.comments'), path: "/admin/comments"},
        {id: "playlists", icon: PlayCircle, label: t('admin.playlists'), path: "/admin/playlists"},
        {id: "content", icon: FileText, label: t('admin.content'), path: "/admin/content"},
        {id: "settings", icon: Settings, label: t('admin.settings'), path: "/admin/settings"},
    ];

    const isActive = (path: string) =>
        path === "/admin"
            ? state.location.pathname === "/admin"
            : state.location.pathname.startsWith(path);

    const handleSignOut = async () => {
        await signOut();
        window.location.href = "/auth/signin";
    };

    const getInitials = (name: string) => {
        return name ? name.charAt(0).toUpperCase() : "U";
    };

    return (
        <div className="min-h-screen flex">
            {/* 侧边栏 */}
            <aside className="w-64 bg-slate-900 text-white flex flex-col">
                <div className="p-4 border-b border-slate-700">
                    <h1 className="text-xl font-bold flex items-center gap-2">
            <span
                className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-bold">
              O
            </span>
                        OrigCMS
                    </h1>
                </div>
                <nav className="flex-1 p-4 space-y-1">
                    {menuItems.map((item) => (
                        <Link
                            key={item.path}
                            to={item.path}
                            className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                                isActive(item.path)
                                    ? "bg-primary text-white"
                                    : "hover:bg-slate-800 text-slate-300 hover:text-white"
                            }`}
                        >
                            <item.icon size={20}/>
                            <span>{item.label}</span>
                        </Link>
                    ))}
                </nav>
                <div className="p-4 border-t border-slate-700">
                    <Link
                        to="/"
                        className="flex items-center gap-2 px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
                    >
                        <ArrowLeft size={18}/>
                        <span>{t('admin.exitAdmin')}</span>
                    </Link>
                </div>
            </aside>

            {/* 主内容区 */}
            <div className="flex-1 flex flex-col">
                <header className="h-16 bg-background border-b flex items-center justify-between px-6">
                    <div className="text-sm text-muted-foreground">{t('admin.welcomeAdmin')}</div>
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="flex items-center gap-2 px-2">
                                <Avatar className="h-8 w-8">
                                    <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                                        {getInitials(user?.displayName || user?.username || "Admin")}
                                    </AvatarFallback>
                                </Avatar>
                                <span className="text-sm font-medium text-foreground">
                  {user?.displayName || user?.username || "管理员"}
                </span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-56">
                            <DropdownMenuLabel>{t('admin.myAccount')}</DropdownMenuLabel>
                            <DropdownMenuSeparator/>
                            <DropdownMenuItem onClick={handleSignOut} className="text-destructive">
                                <LogOut className="mr-2 h-4 w-4"/>
                                <span>{t('admin.logout')}</span>
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </header>
                <main className="flex-1 p-6 bg-muted/40 overflow-auto">{children}</main>
            </div>
        </div>
    );
}