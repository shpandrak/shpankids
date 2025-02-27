import {defineConfig} from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    base: "/ui",
    // build: {
    //     // Enable filename hashing for cache busting
    //     rollupOptions: {
    //         output: {
    //             entryFileNames: "assets/[name].[hash].js",
    //             chunkFileNames: "assets/[name].[hash].js",
    //             assetFileNames: "assets/[name].[hash].[ext]",
    //         }
    //     }
    // }
});
