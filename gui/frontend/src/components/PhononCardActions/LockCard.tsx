/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { lockOpen } from 'ionicons/icons';
import { Button, IconButton } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';

export const LockCard: React.FC<{
  setThisCard;
  isMini?: boolean;
}> = ({ setThisCard, isMini = false }) => {
  const { t } = useTranslation();

  const lockCard = () => {
    setThisCard((prevState) => ({
      ...prevState,
      IsLocked: true,
    }));
  };

  return isMini ? (
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
