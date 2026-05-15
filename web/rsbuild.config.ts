/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {defineConfig} from '@rsbuild/core';
import {pluginReact} from '@rsbuild/plugin-react';
import * as path from 'path';

export default defineConfig({
    plugins: [pluginReact()],
    html: {
        template: './index.html',
        title: 'OrigStudio - Shared Platform',
    },
    source: {
        entry: {
            index: './src/index.tsx',
        },
    },
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    output: {
        assetPrefix: '/', // Ensure resources load with absolute paths in nested routes
    },
    server: {
        port: 18080,
        historyApiFallback: true, // Ensure client-side routes correctly fallback to index.html
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                pathRewrite: {
                    '^/api': '/api'
                }
            },
            '/thumbnails': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/uploads': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/hls': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/sprites': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
        // Configure public directory to ensure static resources are handled correctly
        publicDir: {
            name: 'public',
            copyOnBuild: true,
        },
    },
    // Configure CSS to ensure URLs are not parsed as modules
    tools: {
        cssLoader: {
            url: false, // Completely disable URL parsing in CSS
        },
    },
});
