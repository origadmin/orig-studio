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
        title: 'OrigCMS - Shared Platform',
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
    server: {
        port: 18080,
        historyApiFallback: true, // 确保客户端路由能正确回退到 index.html
        proxy: {
            '/api': {
                target: 'http://localhost:9090',
                changeOrigin: true,
            },
        },
    },
    output: {
        assetPrefix: '/', // 确保嵌套路由下的资源加载使用绝对路径
    },
});
