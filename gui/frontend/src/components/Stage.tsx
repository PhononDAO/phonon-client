import { CardDeck } from './CardDeck';
import { PhononWallet } from './PhononWallet';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';

export const Stage = () => {
  return (
    <main className="bg-zinc-900 shadow-top shadow-gray-600 font-bandeins-sans px-6 py-4 flex-grow relative">
      <DndProvider backend={HTML5Backend}>
        <PhononWallet />

        <div className="grid grid-cols-2">
          <CardDeck />
          <CardDeck />
        </div>
      </DndProvider>
    </main>
  );
};
