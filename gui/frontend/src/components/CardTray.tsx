import { useTranslation } from 'react-i18next';
import { useDrop } from 'react-dnd';

import { PhononCard as Card } from '../classes/PhononCard';

export const CardTray: React.FC<{
  card?: Card;
}> = ({ card = null }) => {
  const { t } = useTranslation();

  const [{ canDrop, isOver }, drop] = useDrop(() => ({
    accept: 'PhononCard',
    drop: (item, monitor) => {
      console.log(item);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div
      ref={drop}
      className="w-80 h-52 rounded-lg border border-dashed border-white border-4 bg-phonon-card bg-cover bg-no-repeat overflow-hidden flex items-center justify-center text-xl"
    >
      Drop a card here
    </div>
  );
};
