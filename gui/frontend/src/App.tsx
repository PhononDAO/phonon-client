import React, { Suspense } from 'react';
import { NotificationTray } from './components/NotificationTray';
import { Header } from './components/Header';
import { Stage } from './components/Stage';

const App = () => {
  return (
    <Suspense fallback="loading">
      <NotificationTray />
      <div className="w-full overflow-hidden flex flex-col h-screen bg-black relative">
        <Header />
        <Stage />
      </div>
    </Suspense>
  );
};

export default App;
