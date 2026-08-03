import { defineConfig } from "astro/config";
import zuedocs from "zuedocs/astro";

export default defineConfig({
  site: "https://pluck-delta.vercel.app",
  output: "static",
  integrations: [zuedocs()]
});
