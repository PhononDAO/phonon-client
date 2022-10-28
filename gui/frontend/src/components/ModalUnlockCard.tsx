/* eslint-disable @typescript-eslint/no-unsafe-return */
import { PhononCard as Card } from '../classes/PhononCard';
import {
  Button,
  ButtonGroup,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  HStack,
  PinInput,
  PinInputField,
} from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { notifyError, notifySuccess } from '../utils/notify';
import { useRef, useState } from 'react';

export const ModalUnlockCard: React.FC<{
  isOpen;
  onClose;
  card: Card;
  setThisCard;
}> = ({ isOpen, onClose, card, setThisCard }) => {
  const { t } = useTranslation();
  const [isError, setIsError] = useState(false);
  const initialRef = useRef(null);
  const pinLength = 6;

  const unlockCard = () => {
    if (false) {
      setIsError(true);
      notifyError(t('Wrong PIN for "' + String(card.CardId) + '"!'));

      setInterval(() => {
        setIsError(false);
      }, 1000);
    } else {
      setThisCard((prevState) => ({
        ...prevState,
        IsLocked: false,
      }));
      onClose();
      notifySuccess(t('Card "' + String(card.CardId) + '" is unlocked!'));
    }
  };

  return (
    <Modal
      size={'sm'}
      isOpen={isOpen}
      onClose={onClose}
      initialFocusRef={initialRef}
    >
      <ModalOverlay />
      <ModalContent
        className={'overflow-hidden ' + (isError ? 'animate-errorShake' : '')}
      >
        <ModalHeader>
          <div className="font-noto-sans-mono">
            <div className="text-2xl">
              {card.VanityName ? card.VanityName : card.CardId}
            </div>
            {card.VanityName && (
              <div className="text-sm text-gray-400">{card.CardId}</div>
            )}
          </div>
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>
          <div className="mb-2">{t('Enter PIN to unlock card:')}</div>
          <form>
            <HStack>
              <PinInput mask>
                {Array(pinLength)
                  .fill(null)
                  .map((val, key) => (
                    <PinInputField
                      bg="gray.700"
                      color="white"
                      key={key}
                      ref={key === 0 ? initialRef : null}
                    />
                  ))}
              </PinInput>
            </HStack>
          </form>
        </ModalBody>

        <ModalFooter>
          <ButtonGroup spacing={2}>
            <Button size="sm" variant="ghost" onClick={onClose}>
              {t('Cancel')}
            </Button>
            <Button size="sm" colorScheme="green" onClick={unlockCard}>
              {t('UNLOCK')}
            </Button>
          </ButtonGroup>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
