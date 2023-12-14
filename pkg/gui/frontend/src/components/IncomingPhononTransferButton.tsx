/* eslint-disable @typescript-eslint/no-unsafe-return */
import { Button, useDisclosure } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../interfaces/interfaces';
import { ModalIncomingTransferProposal } from './ModalIncomingTransferProposal';

export const IncomingPhononTransferButton: React.FC<{
  sourceCard: PhononCard;
  destinationCard: PhononCard;
}> = ({ sourceCard, destinationCard }) => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      <Button
        bg="blue"
        textColor="white"
        size="sm"
        _hover={{ bg: 'blue.400' }}
        className="place-self-end"
        onClick={onOpen}
      >
        {t('See Incoming Transfer Request')}
      </Button>
      <ModalIncomingTransferProposal
        sourceCard={sourceCard}
        destinationCard={destinationCard}
        isOpen={isOpen}
        onClose={onClose}
      />
    </>
  );
};
