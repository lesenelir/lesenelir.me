/** @type {import('prettier').Config} */
module.exports = {
  printWidth: 100,
  tabWidth: 2,
  semi: false,
  singleQuote: true,
  quoteProps: 'consistent',
  trailingComma: 'none',
  arrowParens: 'always',
  endOfLine: 'lf',
  plugins: [
    '@ianvs/prettier-plugin-sort-imports',
    'prettier-plugin-sort-json',
    'prettier-plugin-packagejson',
    'prettier-plugin-tailwindcss' // MUST come last
  ],
  // options for prettier-plugin-sort-json
  ...{
    jsonRecursiveSort: true,
  },
  // options for prettier-plugin-tailwindcss
  ...{
    tailwindFunctions: ['cva', 'cn'],
  },
  // options for @ianvs/prettier-plugin-sort-imports
  importOrder: [
    '^react$',
    '',

    '<THIRD_PARTY_MODULES>',
    '<TYPES>',
    '',

    '^@/',
    '<TYPES>^@/',
    '',

    '^~/',
    '<TYPES>^~/',
    '',

    '^[.]',
    '<TYPES>^[.]',
    '',

    '^(?!.*[.]css$)[./].*$',
    '.css$',
    '',
  ]
}
