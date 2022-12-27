import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { ErrorBoundary } from 'react-error-boundary';
import { ErrorFallback } from './components/ErrorFallback';
import { ErrorHandler } from './constants/ErrorHandler';
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n';
import { ChakraProvider } from '@chakra-ui/react';
import theme from './chakra-theme';
import './styles.css';

const container = document.getElementById('root');

const root = createRoot(container);
root.render(
  <React.StrictMode>
    <ErrorBoundary FallbackComponent={ErrorFallback} onError={ErrorHandler}>
      <I18nextProvider i18n={i18n}>
        <ChakraProvider theme={theme}>
          <App />
        </ChakraProvider>
      </I18nextProvider>
    </ErrorBoundary>
  </React.StrictMode>
);
