/* eslint-disable @typescript-eslint/no-unsafe-return */
import { IonIcon } from '@ionic/react';
import { send } from 'ionicons/icons';
import { Button, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../interfaces/interfaces';
import { ModalOutgoingTransferProposal } from './ModalOutgoingTransferProposal';
import { Pluralize } from 'pluralize-react';

export const OutgoingPhononTransferButton: React.FC<{
  destinationCard: PhononCard;
}> = ({ destinationCard }) => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      <Button
        bg="blue"
        size="sm"
        _hover={{ bg: 'gray.900' }}
        textColor="white"
        leftIcon={<IonIcon icon={send} />}
        onClick={onOpen}
      >
        {t('Send ')}

        <Pluralize
          className="ml-1"
          count={destinationCard.OutgoingTransferProposal?.Phonons.length}
          singular={t('Proposed Phonon')}
        />
      </Button>
      <ModalOutgoingTransferProposal
        destinationCard={destinationCard}
        isOpen={isOpen}
        onClose={onClose}
      />
    </>
  );
};
