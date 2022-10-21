/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      boxShadow: {
        top: '0 0px 8px',
      },
      backgroundImage: {
        'phonon-card': "url('./assets/images/card-bg.png')",
        'phonon-card-light': "url('./assets/images/card-bg-light.png')",
        'phonon-card-blue': "url('./assets/images/card-bg-blue.png')",
      },
      fontFamily: {
        'bandeins-sans': ['BandeinsSansRegular'],
        'bandeins-sans-semibold': ['BandeinsSansSemiBold'],
        'bandeins-sans-bold': ['BandeinsSansBold'],
        'bandeins-sans-light': ['BandeinsSansLight'],
        'noto-sans-mono': ['Noto Sans Mono', 'monospace'],
      },
      fontSize: {
        'phonon-card': '2.85rem',
        xxs: '0.7rem',
      },
      rotate: {
        30: '30deg',
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
