import React, { Suspense } from 'react';
import { NotificationTray } from './components/NotificationTray';
import { Header } from './components/Header';
import { Stage } from './components/Stage';
import { PageLoading } from './components/PageLoading';

const App = () => {
  return (
    <Suspense fallback={<PageLoading />}>
      <NotificationTray />
      <div className="w-full overflow-scroll flex flex-col h-screen bg-black relative">
        <Header />
        <Stage />
      </div>
    </Suspense>
  );
};

export default App;
