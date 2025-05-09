module.exports = {
  // The root directory that Jest should scan for tests and modules
  roots: ["<rootDir>/src"],

  // A list of paths to directories that Jest should use to search for files in
  testMatch: ["**/__tests__/**/*.js", "**/?(*.)+(spec|test).js"],

  // An array of regexp pattern strings that are matched against all test paths
  // matched tests are skipped
  testPathIgnorePatterns: ["/node_modules/"],

  // An array of regexp pattern strings that are matched against all source file paths
  // matched files will skip transformation
  transformIgnorePatterns: ["/node_modules/"],

  // The directory where Jest should output its coverage files
  coverageDirectory: "coverage",

  // An array of regexp pattern strings used to skip coverage collection
  coveragePathIgnorePatterns: ["/node_modules/"],

  // Indicates whether each individual test should be reported during the run
  verbose: true,

  // Sets the test environment
  testEnvironment: "jsdom",

  // Setup files
  setupFilesAfterEnv: ["<rootDir>/src/setupTests.js"],
  
  // Mock file extensions
  moduleNameMapper: {
    "\\.(css|less|scss|sass)$": "identity-obj-proxy"
  }
};