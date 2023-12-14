/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { ellipsisHorizontalCircle } from 'ionicons/icons';
import { Button, IconButton } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { CardManagementContext } from '../../contexts/CardManagementContext';
import { PhononCard } from '../../interfaces/interfaces';

export const ViewPhonons: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { addCardsToState, isCardsMini } = useContext(CardManagementContext);

  const viewPhonons = () => {
    card.InTray = true;
    addCardsToState([card]);
  };

  return isCardsMini && !card.InTray ? (
    <IconButton
      bg="darkGray.100"
      aria-label={t('View Phonons')}
      size="xs"
      icon={<IonIcon icon={ellipsisHorizontalCircle} />}
      onClick={viewPhonons}
    />
  ) : (
    <Button
      bg="darkGray.100"
      size="xs"
      leftIcon={<IonIcon icon={ellipsisHorizontalCircle} />}
      onClick={viewPhonons}
    >
      {t('View Phonons')}
    </Button>
  );
};
