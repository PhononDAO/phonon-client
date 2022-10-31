/* eslint-disable @typescript-eslint/no-unsafe-return */
import { PhononCard as Card } from '../../classes/PhononCard';
import { IonIcon } from '@ionic/react';
import { lockClosed } from 'ionicons/icons';
import { ModalUnlockCard } from '../ModalUnlockCard';
import { useDisclosure } from '@chakra-ui/react';
import { useContext } from 'react';
import { CardManagementContext } from '../../assets/contexts/CardManagementContext';

export const UnlockCard: React.FC<{
  card: Card;
}> = ({ card }) => {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { isCardsMini } = useContext(CardManagementContext);

  return (
    <>
      <button
        onClick={onOpen}
        className="w-full z-50 text-amber-400 hover:text-amber-300 text-center"
      >
        <IonIcon
          className={
            'p-16 duration-150 ' +
            (isCardsMini && !card.TrayId
              ? 'text-4xl hover:text-5xl'
              : 'text-6xl hover:text-7xl')
          }
          icon={lockClosed}
        />
      </button>
      <ModalUnlockCard isOpen={isOpen} onClose={onClose} card={card} />
    </>
  );
};
