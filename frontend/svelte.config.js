import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),

  kit: {
    adapter: adapter({
      // Output directory for embedded Go assets
      pages: "build",
      assets: "build",
      fallback: "index.html",
      precompress: false,
      strict: false,
    }),
    paths: {
      // Base path if serving from a subpath
      base: "",
    },
    prerender: {
      entries: ["/", "/config", "/series"],
      handleMissingId: "ignore",
      handleUnseenRoutes: "ignore",
    },
  },
};

export default config;
