const path = require('path')

module.exports = {
  env: {
    browser: true,
    es6: true,
    node: true,
  },
  extends: [
    'plugin:@typescript-eslint/recommended',
    'plugin:import/errors',
    'plugin:import/warnings',
    'plugin:import/typescript',
  ],
  globals: {
    AnnConfig: true,
    ga: true,
    gon: true,
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2018,
    project: path.resolve(__dirname, './tsconfig.json'),
    sourceType: 'module',
    tsconfigRootDir: __dirname,
  },
  rules: {
    '@typescript-eslint/indent': ['error', 2],
    '@typescript-eslint/no-unused-vars': ['error', { 'argsIgnorePattern': '^_' }],
    'import/no-default-export': 'error',
  },
}
