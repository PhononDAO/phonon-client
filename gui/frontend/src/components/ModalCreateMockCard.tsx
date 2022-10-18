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

export const ModalCreateMockCard: React.FC<{ isOpen; onClose }> = ({
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>{t('Create Mock Card')}</ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>CREATE MOCK CARD HERE</ModalBody>

        <ModalFooter>
          <Button colorScheme="green" mr={3}>
            {t('Create')}
          </Button>
          <Button onClick={onClose}>{t('Cancel')}</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
