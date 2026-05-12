import React, {StrictMode, useContext, useRef, useMemo} from 'react';
import {createRoot} from 'react-dom/client';
import {RouterProvider, createRouter} from '@tanstack/react-router';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {Toaster} from 'sonner';
import {AuthProvider} from '@/contexts/auth';
import {AuthContext} from '@/contexts/auth/AuthContext';
import {NotificationProvider} from '@/contexts/NotificationContext';
import {ThemeProvider} from '@/themes';
import {routeTree} from './routes.gen';
import './i18n';
import './index.css';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            refetchOnWindowFocus: false,
            retry: 1,
            staleTime: 5 * 60 * 1000,
        },
    },
});

/**
 * InnerApp - renders inside AuthProvider so it can access AuthContext.
 * Creates the router with auth context injected, enabling route guards
 * to read authentication state via context.auth instead of localStorage.
 */
function InnerApp() {
    const auth = useContext(AuthContext);
    if (!auth) throw new Error('InnerApp must be used within AuthProvider');

    // Use ref + getter pattern to avoid router recreation on auth changes.
    // The router reads authRef.current at access time, always getting the latest value.
    const authRef = useRef(auth);
    authRef.current = auth;

    const router = useMemo(() => createRouter({
        routeTree,
        context: {get auth() { return authRef.current; }},
        defaultPreload: 'intent',
        pathParamsAllowedCharacters: ['@'],
    }), []);

    declare module '@tanstack/react-router' {
        interface Register {
            router: typeof router;
        }
    }

    return <RouterProvider router={router}/>;
}

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');

createRoot(rootElement).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <ThemeProvider>
                <AuthProvider>
                    <NotificationProvider>
                        <InnerApp/>
                    </NotificationProvider>
                </AuthProvider>
                <Toaster position="top-right" richColors closeButton/>
            </ThemeProvider>
        </QueryClientProvider>
    </StrictMode>
);
