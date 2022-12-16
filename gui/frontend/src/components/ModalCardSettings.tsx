import {
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  FormHelperText,
  HStack,
  PinInput,
  PinInputField,
} from '@chakra-ui/react';
import { useContext } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { notifySuccess } from '../utils/notify';

type SettingsFormData = {
  vanityName: string;
  cardPin: string;
};

export const ModalCardSettings: React.FC<{
  card: PhononCard;
  isOpen;
  onClose;
}> = ({ card, isOpen, onClose }) => {
  const { t } = useTranslation();
  const {
    control,
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<SettingsFormData>();
  const watchVanityName = watch('vanityName', card?.VanityName);
  const { addCardsToState } = useContext(CardManagementContext);

  const pinLength = 6;

  // event when you start mining a phonon
  const onSubmit = (data: SettingsFormData, event) => {
    event.preventDefault();

    card.VanityName = data.vanityName;

    addCardsToState([card]);

    onClose();
    notifySuccess(
      t('Card "{{cardId}}" settings saved!', { cardId: card.CardId })
    );
  };

  return (
    <Modal size="sm" isOpen={isOpen} onClose={onClose}>
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent>
        <ModalHeader>
          <div className="font-noto-sans-mono">
            <div className="text-sm">{t('Card settings for')}</div>
            <div className="text-2xl">
              {watchVanityName ? watchVanityName : card.CardId}
            </div>
            {watchVanityName && (
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
            <div className="grid grid-cols-1 gap-y-6">
              <FormControl>
                <FormLabel>{t('Vanity Name')}</FormLabel>
                <Input
                  bg="gray.700"
                  color="white"
                  type="text"
                  maxLength={20}
                  placeholder={card.CardId}
                  {...register('vanityName')}
                />
                <FormHelperText>
                  {t(
                    'This name is for internal use only. When pairing with remote cards, your card ID will show.'
                  )}
                </FormHelperText>
              </FormControl>

              <FormControl>
                <FormLabel>{t('Change Card PIN')}</FormLabel>
                <Controller
                  control={control}
                  name="cardPin"
                  rules={{
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
                      >
                        {Array(pinLength)
                          .fill(null)
                          .map((val, key) => (
                            <PinInputField
                              bg="gray.700"
                              color="white"
                              key={key}
                            />
                          ))}
                      </PinInput>
                    </HStack>
                  )}
                />
                {errors.cardPin && (
                  <span className="text-red-600">{errors.cardPin.message}</span>
                )}
                <FormHelperText>
                  {t('Card PIN will only change if you provide a new PIN.')}
                </FormHelperText>
              </FormControl>
            </div>
          </ModalBody>

          <ModalFooter>
            <Button colorScheme="green" type="submit" mr={3}>
              {t('Save')}
            </Button>
            <Button onClick={onClose}>{t('Cancel')}</Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
