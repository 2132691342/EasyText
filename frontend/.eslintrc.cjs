// ESLint 配置 — EasyText 前端
//
// 重点规则:
//   - no-explicit-any: 禁止 any 类型擦除（必须显式 unknown 或具体类型）
//   - no-empty: 禁止空 catch 块（至少 console.warn）
//   - @typescript-eslint/no-unused-vars: 死代码检测
//
// 运行: npm run lint
// 修复: npm run lint -- --fix

module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true,
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module',
    extraFileExtensions: ['.vue'],
  },
  plugins: ['@typescript-eslint', 'vue'],
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:vue/vue3-recommended',
  ],
  rules: {
    // 强制：禁止 any（用 unknown 替代）
    '@typescript-eslint/no-explicit-any': 'error',
    // 强制：禁止空 catch（必须 console.warn 或 throw）
    'no-empty': ['error', { allowEmptyCatch: false }],
    // 允许未使用以 _ 开头的变量
    '@typescript-eslint/no-unused-vars': ['warn', {
      argsIgnorePattern: '^_',
      varsIgnorePattern: '^_',
    }],
    // 允许 console.warn/error（项目故意使用做错误日志）
    'no-console': ['warn', { allow: ['warn', 'error', 'info'] }],
    // Vue: 允许 v-html 在受信任场景使用（Markdown 渲染等）
    'vue/no-v-html': 'warn',
    // TypeScript: 允许 .ts 文件使用 namespace
    '@typescript-eslint/no-namespace': 'off',
    // 关闭过度严格的 Vue prop 类型检查（Vue 3 + TS 已用 defineProps<Props>() 模式）
    'vue/require-default-prop': 'off',
    'vue/attribute-hyphenation': 'off',
    'vue/v-on-event-hyphenation': 'off',
  },
  overrides: [
    {
      // 自动生成的 wailsjs 绑定：禁止修改
      files: ['wailsjs/**'],
      rules: {
        '@typescript-eslint/no-explicit-any': 'off',
        'no-empty': 'off',
      },
    },
  ],
  ignorePatterns: ['dist/**', 'node_modules/**', 'wailsjs/**'],
}