import { defineConfig } from "astro/config";
import zuedocs from "zuedocs/astro";

export default defineConfig({
  site: "https://pluck.ashray.xyz",
  output: "static",
  integrations: [zuedocs()]
});
