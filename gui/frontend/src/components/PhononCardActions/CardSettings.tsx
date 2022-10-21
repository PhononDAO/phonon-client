/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { cog } from 'ionicons/icons';
import { Button, IconButton, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { ModalCardSettings } from '../ModalCardSettings';

export const CardSettings: React.FC<{
  isMini?: boolean;
}> = ({ isMini = false }) => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      {isMini ? (
        <IconButton
          bg="darkGray.100"
          aria-label={t('Settings')}
          size="xs"
          icon={<IonIcon icon={cog} />}
          onClick={onOpen}
        />
      ) : (
        <Button
          bg="darkGray.100"
          size="xs"
          leftIcon={<IonIcon icon={cog} />}
          onClick={onOpen}
        >
          {t('Settings')}
        </Button>
      )}
      <ModalCardSettings isOpen={isOpen} onClose={onClose} />
    </>
  );
};
