// SignIn page - uses shadcn/ui components
import {useState} from "react";
import {useNavigate, useSearch, Link} from "@tanstack/react-router";
import {useTranslation} from 'react-i18next';
import {api, setAuth} from "@/lib/request";
import {useAuth} from "@/hooks/useAuth";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card";

interface AuthResponse {
    access_token: string;
    refresh_token?: string;
    token_type: string;
    expires_in: number;
    user: {
        id: string;
        username: string;
        nickname?: string;
        is_staff?: boolean;
    };
}

export default function SignInPage() {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const {login} = useAuth();
    const {t} = useTranslation();

    // Read redirect param from search params (set by _authenticated layout route)
    const search = useSearch({strict: false}) as { redirect?: string };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);

        try {
            const res = await api.post<AuthResponse>("/auth/signin", {username, password});
            setAuth({
                access_token: res.access_token,
                refresh_token: res.refresh_token,
                expires_in: res.expires_in,
                token_type: res.token_type,
            });
            login(res.access_token, res.refresh_token || '', {
                id: res.user.id,
                username: res.user.username,
                displayName: res.user.nickname || res.user.username,
                roles: res.user.is_staff ? ["admin"] : ["user"],
            });

            // Navigate to redirect target or home page
            // Security: only allow internal paths (starting with /)
            const redirectUrl = search.redirect;
            if (redirectUrl && redirectUrl.startsWith('/')) {
                navigate({to: redirectUrl});
            } else {
                navigate({to: "/"});
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : "Login failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex">
            {/* Left background panel */}
            <div
                className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-500 flex-col justify-between p-12">
                <div className="flex items-center gap-2 text-white">
                    <img src="/logo.svg" alt="OrigStudio" className="h-8 w-8" />
                    <span className="text-2xl font-bold">OrigStudio</span>
                </div>
                <blockquote className="text-white/90 text-xl italic">
                    "A powerful content management system that helps you manage your media and content with ease."
                </blockquote>
                <div className="text-white/40 text-sm">&copy; 2026 OrigStudio. All rights reserved.</div>
            </div>

            {/* Right form panel */}
            <div className="flex-1 flex items-center justify-center p-8 bg-gray-50">
                <Card className="w-full max-w-md">
                    <CardHeader>
                        <CardTitle>{t('auth.signIn.title')}</CardTitle>
                        <CardDescription>{t('auth.signIn.desc')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {error && (
                            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-destructive text-sm">
                                {error}
                            </div>
                        )}
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="space-y-2">
                                <label htmlFor="username" className="text-sm font-medium">{t('auth.signIn.usernameLabel')}</label>
                                <Input
                                    id="username"
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder={t('auth.signIn.usernamePlaceholder')}
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <label htmlFor="password" className="text-sm font-medium">{t('auth.signIn.passwordLabel')}</label>
                                <Input
                                    id="password"
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder={t('auth.signIn.passwordPlaceholder')}
                                    required
                                />
                            </div>
                            <Button type="submit" className="w-full" disabled={loading}>
                                {loading ? t('auth.signIn.submitting') : t('auth.signIn.submit')}
                            </Button>
                        </form>
                    </CardContent>
                    <CardFooter className="justify-center">
                        <p className="text-sm text-muted-foreground">
                            {t('auth.signIn.noAccount')}{" "}
                            <Link to="/auth/signup" className="text-primary hover:underline">{t('auth.signIn.signUp')}</Link>
                        </p>
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}
