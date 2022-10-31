import { useContext } from 'react';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

export const CardShadow: React.FC = () => {
  const { isCardsMini } = useContext(CardManagementContext);
  return (
    <div
      className={
        'rounded-lg border border-dashed border-gray-700 border-4 bg-phonon-card bg-cover bg-no-repeat overflow-hidden flex items-center justify-center text-xl ' +
        (isCardsMini ? 'w-56 h-36 ' : 'w-80 h-52')
      }
    ></div>
  );
};
