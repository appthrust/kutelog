import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { createHtmlPlugin } from "vite-plugin-html";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
	plugins: [
		tailwindcss(),
		viteSingleFile(),
		createHtmlPlugin({
			minify: {
				collapseWhitespace: true,
				removeComments: true,
				removeRedundantAttributes: true,
				removeScriptTypeAttributes: true,
				removeStyleLinkTypeAttributes: true,
				useShortDoctype: true,
				minifyCSS: true,
				minifyJS: true,
			},
		}),
	],
	build: {
		outDir: "dist",
		sourcemap: "inline",
		minify: "terser",
		cssMinify: true,
		cssCodeSplit: false,
		rollupOptions: {
			input: "index.html",
			output: {
				inlineDynamicImports: true,
				format: "iife",
			},
		},
		terserOptions: {
			format: {
				comments: false,
			},
		},
	},
	experimental: {
		renderBuiltUrl(filename) {
			return { relative: true };
		},
	},
});
