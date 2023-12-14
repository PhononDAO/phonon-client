/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { addCircleSharp } from 'ionicons/icons';
import { Button, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../../interfaces/interfaces';
import { ModalCreatePhonon } from '../ModalCreatePhonon';

export const CreatePhonon: React.FC<{ card: PhononCard }> = ({ card }) => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      <Button
        bg="white"
        textColor="black"
        leftIcon={<IonIcon icon={addCircleSharp} />}
        onClick={onOpen}
      >
        {t('Create Phonon')}
      </Button>
      <ModalCreatePhonon card={card} isOpen={isOpen} onClose={onClose} />
    </>
  );
};
