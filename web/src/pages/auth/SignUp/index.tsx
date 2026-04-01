// SignUp 注册页面 - 使用 shadcn/ui 组件
import {useState} from "react";
import {useNavigate, Link} from "@tanstack/react-router";
import {api, setAuth} from "@/lib/request";
import {useAuth} from "@/hooks/useAuth";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card";

interface AuthResponse {
    access_token: string;
    token_type: string;
    expires_in: number;
    user: {
        id: number;
        username: string;
        nickname?: string;
        is_staff?: boolean;
    };
}

export default function SignUpPage() {
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const {login} = useAuth();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");

        if (password !== confirmPassword) {
            setError("Passwords do not match");
            return;
        }
        if (password.length < 6) {
            setError("Password must be at least 6 characters");
            return;
        }

        setLoading(true);
        try {
            const res = await api.post<AuthResponse>("/auth/signup", {username, email, password});
            setAuth({
                access_token: res.access_token,
                expires_in: res.expires_in,
                token_type: res.token_type,
            });
            login(res.access_token, {
                id: res.user.id,
                username: res.user.username,
                displayName: res.user.nickname || res.user.username,
                roles: res.user.is_staff ? ["admin"] : ["user"],
            });
            navigate({to: "/"});
        } catch (err) {
            setError(err instanceof Error ? err.message : "Registration failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex">
            {/* 左侧背景 */}
            <div
                className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-500 flex-col justify-between p-12">
                <div className="flex items-center gap-2 text-white">
                    <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                              d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                    </svg>
                    <span className="text-2xl font-bold">OrigCMS</span>
                </div>
                <blockquote className="text-white/90 text-xl italic">
                    "Join us and start managing your content and media with ease."
                </blockquote>
                <div className="text-white/40 text-sm">© 2026 OrigCMS. All rights reserved.</div>
            </div>

            {/* 右侧表单 */}
            <div className="flex-1 flex items-center justify-center p-8 bg-gray-50">
                <Card className="w-full max-w-md">
                    <CardHeader>
                        <CardTitle>Create an account</CardTitle>
                        <CardDescription>Enter your details to get started</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {error && (
                            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
                                {error}
                            </div>
                        )}
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="space-y-2">
                                <label htmlFor="username" className="text-sm font-medium">Username</label>
                                <Input
                                    id="username"
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder="Choose a username"
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <label htmlFor="email" className="text-sm font-medium">Email</label>
                                <Input
                                    id="email"
                                    type="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    placeholder="Enter your email"
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <label htmlFor="password" className="text-sm font-medium">Password</label>
                                <Input
                                    id="password"
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder="Create a password"
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <label htmlFor="confirmPassword" className="text-sm font-medium">Confirm
                                    Password</label>
                                <Input
                                    id="confirmPassword"
                                    type="password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    placeholder="Confirm your password"
                                    required
                                />
                            </div>
                            <Button type="submit" className="w-full" disabled={loading}>
                                {loading ? "Creating account..." : "Create Account"}
                            </Button>
                        </form>
                    </CardContent>
                    <CardFooter className="justify-center">
                        <p className="text-sm text-muted-foreground">
                            Already have an account?{" "}
                            <Link to="/auth/signin" className="text-primary hover:underline">Sign in</Link>
                        </p>
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}