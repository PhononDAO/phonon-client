import {
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
} from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';

export const ModalOutgoingTransferProposal: React.FC<{
  card;
  isOpen;
  onClose;
}> = ({ card, isOpen, onClose }) => {
  const { t } = useTranslation();

  return (
    <Modal isOpen={isOpen} size="3xl" onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          {t('Transfer Phonons')}: {card.CardId}
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>TRANSFER OUTGOING PHONONS HERE</ModalBody>

        <ModalFooter>
          <Button colorScheme="green" mr={3}>
            {t('Save')}
          </Button>
          <Button onClick={onClose}>{t('Cancel')}</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
