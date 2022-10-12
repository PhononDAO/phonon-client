import React, { Suspense } from 'react';
import { Header } from './components/Header';
import { NotificationTray } from './components/NotificationTray';
import { Stage } from './components/Stage';

const App = () => {
  return (
    <Suspense fallback="loading">
      <NotificationTray />
      <div className="w-100 flex flex-col h-screen bg-black relative">
        <Header />
        <Stage />
      </div>
    </Suspense>
  );
};

export default App;
