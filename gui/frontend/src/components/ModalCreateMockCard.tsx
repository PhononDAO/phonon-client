import {
  Button,
  ButtonGroup,
  UnorderedList,
  ListItem,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
} from '@chakra-ui/react';
import { PhononCard as Card } from '../classes/PhononCard';
import { useTranslation } from 'react-i18next';
import { notifySuccess } from '../utils/notify';
import { useContext } from 'react';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

export const ModalCreateMockCard: React.FC<{
  isOpen;
  onClose;
}> = ({ isOpen, onClose }) => {
  const { t } = useTranslation();
  const { addPhononCardsToState } = useContext(CardManagementContext);

  const createMockCard = () => {
    const aCard = new Card();
    aCard.CardId = '04e0d5eb884ae' + String(Math.floor(Math.random() * 999));
    aCard.IsMock = true;

    addPhononCardsToState([aCard]);

    notifySuccess(t('New mock card created!'));

    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>{t('Create Mock Card')}</ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>
          <div className="mb-4">
            {t('A mock card is a temporary card to test the platform.')}
          </div>
          <UnorderedList>
            <ListItem>
              Mock cards are deleted, including all phonons, when this app is
              closed.
            </ListItem>
            <ListItem>
              Mock cards have a different certificate than alpha and testnet
              phonon cards and therefore cannot communicate with them.
            </ListItem>
          </UnorderedList>
        </ModalBody>

        <ModalFooter>
          <ButtonGroup spacing={2}>
            <Button size="sm" variant="ghost" onClick={onClose}>
              {t('Cancel')}
            </Button>
            <Button size="sm" colorScheme="green" onClick={createMockCard}>
              {t('I understand, create mock card')}
            </Button>
          </ButtonGroup>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
