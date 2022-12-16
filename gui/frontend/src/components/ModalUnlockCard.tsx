/* eslint-disable @typescript-eslint/no-unsafe-return */
import { PhononCard } from '../interfaces/interfaces';
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
import React, { useState, useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';

type PINFormData = {
  cardPin: string;
};

export const ModalUnlockCard: React.FC<{
  isOpen;
  onClose;
  card: PhononCard;
}> = ({ isOpen, onClose, card }) => {
  const { t } = useTranslation();
  const [isError, setIsError] = useState(false);
  const { addCardsToState } = useContext(CardManagementContext);
  const pinLength = 6;

  const {
    control,
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

      addCardsToState([card]);
      setValue('cardPin', '');

      onClose();
      notifySuccess(
        t('Card "{{cardId}}" is unlocked!', {
          cardId: card.CardId,
        })
      );
    }
  };

  return (
    <Modal
      size="sm"
      isOpen={isOpen}
      onClose={() => {
        onClose();
        card.AttemptUnlock = false;
        addCardsToState([card]);
      }}
    >
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent
        className={'overflow-hidden ' + (isError ? 'animate-errorShake' : '')}
      >
        <ModalHeader>
          <div className="font-noto-sans-mono">
            <div className="text-sm">{t('Unlocking')}</div>
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
              name="cardPin"
              rules={{
                required: t('Card PIN Required'),
                minLength: {
                  value: pinLength,
                  message: t('Card PIN too short'),
                },
              }}
              render={({ field: { ...restField } }) => (
                <HStack>
                  <PinInput
                    onChange={restField.onChange}
                    value={restField.value}
                    mask
                    autoFocus
                  >
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
                  addCardsToState([card]);
                }}
              >
                {t('Cancel')}
              </Button>
              <Button
                size="sm"
                className="uppercase"
                colorScheme="green"
                type="submit"
              >
                {t('unlock')}
              </Button>
            </ButtonGroup>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
