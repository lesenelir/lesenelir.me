import pluginNext from '@next/eslint-plugin-next'
import tanstackQueryPlugin from '@tanstack/eslint-plugin-query'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import pluginReact from 'eslint-plugin-react'
import pluginReactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'

import baseConfig from './base.js'

export default [
  ...baseConfig,
  pluginReact.configs.flat.recommended,
  jsxA11y.flatConfigs?.recommended,
  {
    languageOptions: {
      ...pluginReact.configs.flat.recommended.languageOptions,
      ...jsxA11y.flatConfigs?.recommended.languageOptions,
      globals: {
        ...globals.serviceworker,
        ...globals.browser
      }
    }
  },
  {
    plugins: {
      'react-hooks': pluginReactHooks,
      '@tanstack/query': tanstackQueryPlugin,
      '@next/next': pluginNext
    },
    settings: { react: { version: 'detect' } },
    rules: {
      ...pluginReactHooks.configs.recommended.rules,
      ...tanstackQueryPlugin.configs.recommended.rules,
      ...pluginNext.configs.recommended.rules,
      ...pluginNext.configs['core-web-vitals'].rules,

      // React scope no longer necessary with new JSX transform.
      'react/react-in-jsx-scope': 'off'
    }
  }
]
