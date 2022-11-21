/* eslint-disable @typescript-eslint/no-unsafe-return */
import { PhononCard as Card } from '../classes/PhononCard';
import { useForm, Controller } from 'react-hook-form';
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
import { notifySuccess } from '../utils/notify';
import { useState, useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';

type PINFormData = {
  cardPin: string;
};

export const ModalUnlockCard: React.FC<{
  isOpen;
  onClose;
  card: Card;
}> = ({ isOpen, onClose, card }) => {
  const { t } = useTranslation();
  const [isError, setIsError] = useState(false);
  const { addPhononCardsToState } = useContext(CardManagementContext);
  const pinLength = 6;

  const {
    control,
    register,
    handleSubmit,
    setError,
    setValue,
    formState: { errors },
  } = useForm<PINFormData>();

  // event when you start mining a phonon
  const onSubmit = (data: PINFormData, event) => {
    event.preventDefault();

    if (data.cardPin !== '111111') {
      setError('cardPin', {
        type: 'custom',
        message: 'Incorrect PIN, please try again.',
      });
      setValue('cardPin', '');
      setIsError(true);

      setInterval(() => {
        setIsError(false);
      }, 1000);
    } else {
      card.IsLocked = false;
      if (card.FutureAction) {
        card[card.FutureAction] = true;
        card.FutureAction = null;
      }
      card.AttemptUnlock = false;

      addPhononCardsToState([card]);

      onClose();
      notifySuccess(t('Card "' + String(card.CardId) + '" is unlocked!'));
    }
  };

  return (
    <Modal
      size={'sm'}
      isOpen={isOpen}
      onClose={() => {
        onClose();
        card.AttemptUnlock = false;
        addPhononCardsToState([card]);
      }}
    >
      <ModalOverlay />
      <ModalContent
        className={'overflow-hidden ' + (isError ? 'animate-errorShake' : '')}
      >
        <ModalHeader>
          <div className="font-noto-sans-mono">
            <div className="text-sm">Unlocking</div>
            <div className="text-2xl">
              {card.VanityName ? card.VanityName : card.CardId}
            </div>
            {card.VanityName && (
              <div className="text-sm text-gray-400">{card.CardId}</div>
            )}
          </div>
        </ModalHeader>
        <ModalCloseButton />
        <form
          // eslint-disable-next-line @typescript-eslint/no-misused-promises
          onSubmit={handleSubmit(onSubmit)}
        >
          <ModalBody pb={6}>
            <div className="mb-2">{t('Enter PIN to unlock card:')}</div>

            <Controller
              control={control}
              {...register('cardPin', {
                required: 'Card PIN Required',
                minLength: { value: pinLength, message: 'Card PIN too short' },
              })}
              render={({ field: { ...restField } }) => (
                <HStack>
                  <PinInput {...restField} mask>
                    {Array(pinLength)
                      .fill(null)
                      .map((val, key) => (
                        <PinInputField bg="gray.700" color="white" key={key} />
                      ))}
                  </PinInput>
                </HStack>
              )}
            />
            {errors.cardPin && (
              <span className="text-red-600">{errors.cardPin.message}</span>
            )}
          </ModalBody>

          <ModalFooter>
            <ButtonGroup spacing={2}>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  onClose();
                  card.AttemptUnlock = false;
                  addPhononCardsToState([card]);
                }}
              >
                {t('Cancel')}
              </Button>
              <Button size="sm" colorScheme="green" type="submit">
                {t('UNLOCK')}
              </Button>
            </ButtonGroup>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
