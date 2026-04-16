/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import typescriptEslint from '@typescript-eslint/eslint-plugin';
import typescriptParser from '@typescript-eslint/parser';

export default [
  {
    root: true,
    parser: typescriptParser,
    plugins: {
      '@typescript-eslint': typescriptEslint,
    },
    extends: [
      'eslint:recommended',
      'plugin:@typescript-eslint/recommended',
    ],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: ['../*'],
        },
      ],
      'no-console': 'warn',
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
    },
    ignorePatterns: ['dist/', 'node_modules/', '*.gen.ts'],
  },
];

