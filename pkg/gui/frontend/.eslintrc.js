// @ts-ignore
module.exports = {
  parser: '@typescript-eslint/parser', // Specifies the ESLint parser
  ignorePatterns: [
    '**/*.pcss',
    '**/*.css',
    '**/*.less',
    '**/*.json',
    '**/*.otf',
    '.eslintrc.js',
  ],
  parserOptions: {
    ecmaVersion: 2020, // Allows for the parsing of modern ECMAScript features
    sourceType: 'module', // Allows for the use of imports,
    ecmaFeatures: {
      jsx: true, // Allows for the parsing of JSX,
    },
    project: 'tsconfig.json',
    tsconfigRootDir: __dirname,
    projectFolderIgnoreList: [
      'node_modules/*',
      'node_modules',
      'dist',
      'build',
      '.yarn',
      'build-utils',
      'docs',
      './src/assets/**/*',
    ],
  },
  env: {
    browser: true,
    es6: true,
    node: true,
  },
  extends: [
    'plugin:import/errors',
    'plugin:import/warnings',
    'plugin:import/typescript',
    'plugin:react-hooks/recommended',
    'plugin:react/recommended',
    'plugin:prettier/recommended',
    'plugin:@typescript-eslint/eslint-recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking',
  ],
  plugins: [
    'prettier',
    'react',
    'react-hooks',
    'unused-imports',
    'testing-library',
  ],
  rules: {
    '@typescript-eslint/no-unsafe-call': 'off',
    '@typescript-eslint/no-unsafe-argument': 'off',
    '@typescript-eslint/no-unsafe-assignment': 'off',
    '@typescript-eslint/no-unsafe-member-access': 'off',
    '@typescript-eslint/no-explicit-any': 'off',
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
    'comma-dangle': 0,
    'prettier/prettier': [
      'error',
      {
        semi: true,
        trailingComma: 'es5',
        endofLine: 'auto',
        singleQuote: true,
        printWidth: 80,
        tabWidth: 2,
      },
    ],
  },
  settings: {
    react: {
      createClass: 'createReactClass', // Regex for Component Factory to use,
      // default to "createReactClass"
      pragma: 'React', // Pragma to use, default to "React"
      fragment: 'Fragment', // Fragment to use (may be a property of <pragma>), default to "Fragment"
      version: 'detect', // React version. "detect" automatically picks the version you have installed.
      // You can also use `16."off"`, `16.3`, etc, if you want to override the detected value.
      // default to latest and warns if missing
      // It will default to "detect" in the future
    },
  },
};
