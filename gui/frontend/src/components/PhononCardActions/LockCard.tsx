/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { lockOpen } from 'ionicons/icons';
import { Button, IconButton } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { CardManagementContext } from '../../assets/contexts/CardManagementContext';
import { PhononCard } from '../../interfaces/interfaces';

export const LockCard: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { addPhononCardsToState, isCardsMini } = useContext(
    CardManagementContext
  );

  const lockCard = () => {
    card.IsLocked = true;
    card.TrayId = false;
    addPhononCardsToState([card]);
  };

  return isCardsMini && !card.TrayId ? (
    <IconButton
      colorScheme="red"
      aria-label={t('Lock')}
      size="xs"
      icon={<IonIcon icon={lockOpen} />}
      onClick={lockCard}
    />
  ) : (
    <Button
      colorScheme="red"
      size="xs"
      leftIcon={<IonIcon icon={lockOpen} />}
      onClick={lockCard}
    >
      {t('Lock')}
    </Button>
  );
};
