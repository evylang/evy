import { defineConfig, devices } from "@playwright/test"

const platform = process.env.PLATFORM_OVERRIDE || process.platform
export default defineConfig({
  use: {
    baseURL: process.env.BASEURL || "http://localhost:8080",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "ios",
      use: { ...devices["iPhone 14"] },
    },
  ],
  testMatch: "*test.js",
  snapshotPathTemplate: `{testDir}/{testFilePath}-snapshots/{arg}-{projectName}-${platform}{ext}`,
})
