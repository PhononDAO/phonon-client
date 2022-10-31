/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { caretUpCircle } from 'ionicons/icons';
import { Button, IconButton } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { CardManagementContext } from '../../assets/contexts/CardManagementContext';
import { PhononCard } from '../../interfaces/interfaces';

export const CloseCard: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { addPhononCardsToState, isCardsMini } = useContext(
    CardManagementContext
  );

  const closeCard = () => {
    card.TrayId = false;
    addPhononCardsToState([card]);
  };

  return isCardsMini && !card.TrayId ? (
    <IconButton
      bg="darkGray.100"
      aria-label={t('Close Card')}
      size="xs"
      icon={<IonIcon icon={caretUpCircle} />}
      onClick={closeCard}
    />
  ) : (
    <Button
      bg="darkGray.100"
      size="xs"
      leftIcon={<IonIcon icon={caretUpCircle} />}
      onClick={closeCard}
    >
      {t('Close Card')}
    </Button>
  );
};
