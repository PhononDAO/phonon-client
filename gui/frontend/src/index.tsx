import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { ErrorBoundary } from 'react-error-boundary';
import { ErrorFallback } from './components/ErrorFallback';
import { ErrorHandler } from './constants/ErrorHandler';
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n';
import './styles.css';

const container = document.getElementById('root');
// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
const root = createRoot(container!);
root.render(
  <React.StrictMode>
    <ErrorBoundary FallbackComponent={ErrorFallback} onError={ErrorHandler}>
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
    </ErrorBoundary>
  </React.StrictMode>
);
