import { useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';

export const CardShadow: React.FC<{ isOver }> = ({ isOver }) => {
  const { isCardsMini } = useContext(CardManagementContext);
  return (
    <div
      className={
        'absolute rounded-lg border overflow-hidden flex items-center justify-center text-xl border-4' +
        (isOver
          ? ' border-green-500 bg-green-200'
          : ' border-dashed border-gray-700 bg-phonon-card-gray bg-cover bg-no-repeat') +
        (isCardsMini ? ' w-56 h-36' : ' w-80 h-52')
      }
    ></div>
  );
};
