import { useTranslation } from 'react-i18next';
import { useDrop } from 'react-dnd';
import { Button } from '@chakra-ui/react';
import { Card } from './Card';
import { PhononCard } from '../classes/PhononCard';
import { useContext } from 'react';
import { IonIcon } from '@ionic/react';
import { cloudDownload } from 'ionicons/icons';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

export const CardTray: React.FC<{
  card: PhononCard;
  canHaveRemote?: boolean;
  setDeckCard;
}> = ({ card = null, canHaveRemote = false, setDeckCard }) => {
  const { t } = useTranslation();
  const { addPhononCardsToState, setIsCardsMini } = useContext(
    CardManagementContext
  );

  const [{ isOver }, drop] = useDrop(() => ({
    accept: 'PhononCard',
    drop: (item: PhononCard, monitor) => {
      monitor.getItem().TrayId = true;
      addPhononCardsToState([item]);
      setDeckCard(item);
      setIsCardsMini(true);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return card?.TrayId ? (
    <div className="w-80 h-52">
      <Card card={card} />
    </div>
  ) : (
    <div
      ref={drop}
      className={
        'w-80 h-52 rounded-lg border border-4 overflow-hidden flex flex-col gap-y-2 items-center justify-center text-xl transition-all ' +
        (isOver
          ? 'border-green-500 bg-green-200'
          : 'border-dashed border-white bg-phonon-card bg-cover bg-no-repeat')
      }
    >
      <div className="text-white ">Drop a card here</div>
      {canHaveRemote && (
        <>
          <div>
            <span className="block text-center text-white ">OR</span>
          </div>
          <Button
            leftIcon={<IonIcon icon={cloudDownload} />}
            size="md"
            className="uppercase"
            onClick={() => {
              alert('TODO: Show pairing next steps.');
            }}
          >
            {t('Pair Remote Card')}
          </Button>
        </>
      )}
    </div>
  );
};
