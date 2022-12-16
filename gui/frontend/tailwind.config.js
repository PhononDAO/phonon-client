const { scale } = require('style-value-types');

/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      margin: {
        full: '96%',
      },
      height: {
        144: '48rem',
      },
      boxShadow: {
        top: '0 0px 8px',
      },
      backgroundImage: {
        'phonon-card': "url('./assets/images/card-bg.png')",
        'phonon-card-gray': "url('./assets/images/card-bg-gray.png')",
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
        success: {
          '50%': { transform: 'scale(1.2)' },
        },
        outgoing: {
          '0%': { opacity: '0', transform: 'translateX(0rem)' },
          '50%': { opacity: '1', transform: 'translateX(10rem)' },
          '100%': { opacity: '0', transform: 'translateX(32rem)' },
        },
        incoming: {
          '0%': { opacity: '0', transform: 'translateX(32rem)' },
          '50%': { opacity: '1', transform: 'translateX(22rem)' },
          '100%': { opacity: '0', transform: 'translateX(0rem)' },
        },
      },
      animation: {
        dismissIndicator: 'dismissIndicator 8s ease-out 1',
        errorShake: 'errorShake 0.8s cubic-bezier(0.97,0.19,0.07,0.36) both',
        success: 'success 0.4s linear 1',
        outgoing: 'outgoing 4s linear infinite',
        incoming: 'incoming 4s linear infinite',
      },
    },
  },
  plugins: [],
};
