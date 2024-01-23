import { defineConfig, devices } from "@playwright/test"
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
})
