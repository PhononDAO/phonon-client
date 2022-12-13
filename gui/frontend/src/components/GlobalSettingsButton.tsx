/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { cog } from 'ionicons/icons';
import { IconButton, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { GlobalSettingsDrawer } from './GlobalSettingsDrawer';

export const GlobalSettingsButton: React.FC = () => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      <IconButton
        aria-label={t('settings')}
        colorScheme="whiteAlpha"
        icon={<IonIcon icon={cog} className="w-6 h-6" />}
        onClick={onOpen}
      />
      <GlobalSettingsDrawer isOpen={isOpen} onClose={onClose} />
    </>
  );
};
