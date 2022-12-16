import { PhononWallet } from './PhononWallet';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { CardManagementContextProvider } from '../contexts/CardManagementContext';
import { DeckManager } from './DeckManager';
import { CardDragLayer } from './CardDragLayer';
import { PhononDragLayer } from './PhononDragLayer';

export const Stage = () => {
  return (
    <main className="bg-zinc-900 shadow-top shadow-gray-600 font-bandeins-sans px-6 py-4 flex-grow relative">
      <CardManagementContextProvider>
        <DndProvider backend={HTML5Backend}>
          <PhononWallet />
          <DeckManager />
          <CardDragLayer />
          <PhononDragLayer />
        </DndProvider>
      </CardManagementContextProvider>
    </main>
  );
};
