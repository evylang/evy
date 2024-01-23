const { defineConfig, devices } = require("@playwright/test")
module.exports = defineConfig({
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
})
