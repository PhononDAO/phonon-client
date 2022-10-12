/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        'bandeins-sans': ['BandeinsSansRegular'],
        'bandeins-sans-semibold': ['BandeinsSansSemiBold'],
        'bandeins-sans-bold': ['BandeinsSansBold'],
        'bandeins-sans-light': ['BandeinsSansLight'],
      },
      keyframes: {
        dismissIndicator: {
          '0%': { width: '100%' },
          '100%': { width: '0px' },
        },
      },
      animation: {
        dismissIndicator: 'dismissIndicator 8s ease-out 1',
      },
    },
  },
  plugins: [],
};
