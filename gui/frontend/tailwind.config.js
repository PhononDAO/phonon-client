/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      margin: {
        full: '96%',
      },
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
      transitionProperty: {
        height: 'height',
      },
      keyframes: {
        dismissIndicator: {
          '0%': { width: '100%' },
          '100%': { width: '0px' },
        },
        errorShake: {
          '10%, 90%': { transform: 'translate3d(-2px, 0, 0)' },
          '20%, 80%': { transform: 'translate3d(4px, 0, 0)' },
          '30%, 50%, 70%': { transform: 'translate3d(-6px, 0, 0)' },
          '40%, 60%': { transform: 'translate3d(6px, 0, 0)' },
        },
      },
      animation: {
        dismissIndicator: 'dismissIndicator 8s ease-out 1',
        errorShake: 'errorShake 0.8s cubic-bezier(0.97,0.19,0.07,0.36) both',
      },
    },
  },
  plugins: [],
};
