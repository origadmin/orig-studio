/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React, {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';
import './i18n';
import App from './App';
import './index.css';

import {QueryClient, QueryClientProvider} from '@tanstack/react-query';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            refetchOnWindowFocus: false,
            retry: 1,
            staleTime: 5 * 60 * 1000, // 5 minutes
        },
    },
});

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');

createRoot(rootElement).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <App/>
        </QueryClientProvider>
    </StrictMode>
);
