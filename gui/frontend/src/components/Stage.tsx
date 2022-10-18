import { PhononCard as Card } from '../classes/PhononCard';
import { AddMockCardButton } from './AddMockCardButton';
import { PhononCard } from './PhononCard';

export const Stage = () => {
  const aCard = new Card();
  aCard.CardId = '04e0d5eb884a73cf';

  const bCard = new Card();
  bCard.CardId = '048766bd1944fb16';
  bCard.IsMock = true;

  return (
    <main className="bg-zinc-900 shadow-top shadow-gray-600 font-bandeins-sans text-lg text-white px-6 py-4 flex-grow relative">
      <div className="relative flex space-x-10 overflow-x-auto pb-4">
        <PhononCard card={aCard} />
        <PhononCard card={bCard} />
        <AddMockCardButton />
      </div>
    </main>
  );
};
