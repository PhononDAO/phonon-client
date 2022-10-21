/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { ellipsisHorizontalCircle } from 'ionicons/icons';
import { Button, IconButton } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';

export const ViewPhonons: React.FC<{
  setThisCard;
  isMini?: boolean;
}> = ({ setThisCard, isMini = false }) => {
  const { t } = useTranslation();

  const viewPhonons = () => {
    setThisCard((prevState) => ({
      ...prevState,
      IsActive: true,
    }));
  };

  return isMini ? (
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
