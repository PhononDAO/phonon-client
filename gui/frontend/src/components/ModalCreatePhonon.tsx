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

export const ModalCreatePhonon: React.FC<{ card; isOpen; onClose }> = ({
  card,
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent>
        <ModalHeader>
          {t('Create Phonon')}: {card.CardId}
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>CREATE PHONON HERE</ModalBody>

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
