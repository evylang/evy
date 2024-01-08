const { defineConfig, devices } = require("@playwright/test")
module.exports = defineConfig({
  use: {
    baseURL: process.env.BASEURL || "http://localhost:8080",
  },
})
