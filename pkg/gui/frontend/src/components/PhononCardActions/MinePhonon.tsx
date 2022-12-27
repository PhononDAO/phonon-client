/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { hammerSharp } from 'ionicons/icons';
import { Button, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../../interfaces/interfaces';
import { useFeature } from '../../hooks/useFeature';
import { ModalMinePhonon } from '../ModalMinePhonon';

export const MinePhonon: React.FC<{ card: PhononCard }> = ({ card }) => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { CAN_MINE_PHONONS } = useFeature();

  return CAN_MINE_PHONONS ? (
    <>
      <Button
        bg="black"
        _hover={{ bg: 'gray.900' }}
        textColor="white"
        leftIcon={<IonIcon icon={hammerSharp} />}
        onClick={onOpen}
      >
        {t('Mine Phonon')}
      </Button>
      <ModalMinePhonon card={card} isOpen={isOpen} onClose={onClose} />
    </>
  ) : (
    <></>
  );
};
