const path = require("path");
const zlib = require("zlib");
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tsconfigPaths from 'vite-tsconfig-paths'
import viteCompression from 'vite-plugin-compression';
import { chunkSplitPlugin } from 'vite-plugin-chunk-split';
import { createHtmlPlugin } from 'vite-plugin-html';
import StylelintPlugin from 'vite-plugin-stylelint';
import { ViteMinifyPlugin } from 'vite-plugin-minify';
import EnvironmentPlugin from 'vite-plugin-environment';

const PRODUCTION_PLUGINS = [
    react(),
    EnvironmentPlugin('all'),
    tsconfigPaths(),
    viteCompression({
        algorithm: "brotliCompress",
        ext: ".br",
        compressionOptions: {
            params: {
                [zlib.constants.BROTLI_PARAM_QUALITY]: 11,
            },
        },
        threshold: 10240,
        minRatio: 0.8,
        deleteOriginalAssets: false,
    }),
    chunkSplitPlugin({
        strategy: 'single-vendor',
        customSplitting: {
            'react-vendor': ['react', 'react-dom'],
        }
    }),
    createHtmlPlugin({
        minify: true,
        entry: 'src/index.tsx',
        template: 'index.html',
    }),
    StylelintPlugin({
        fix: true,
        quite: true,
    }),
    ViteMinifyPlugin({}),
];

const DEVELOPMENT_PLUGINS = [
    react(),
    tsconfigPaths(),
    EnvironmentPlugin('all')
];

export default ({ mode }) => { 
    const isProduction = mode === 'production';

    return defineConfig({
        base: '/static/dist/',
        root: path.join(__dirname, "/"),
        build: {
            outDir: path.resolve(__dirname, "dist/"),
            cssCodeSplit: false,
            sourcemap: true,
        },
        plugins: isProduction ? PRODUCTION_PLUGINS : DEVELOPMENT_PLUGINS,
        resolve: {
            alias: {
                "@app": path.resolve(__dirname, "./src/app/"),
                "@static": path.resolve(__dirname, "./src/app/static/"),
                "@": path.resolve(__dirname, "./src/"),
            }
        },
        clean: true,
        minify: true,
    })
}
