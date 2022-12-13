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
} from '@chakra-ui/react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { notifySuccess } from '../utils/notify';

type CreatePhononFormData = {
  tokenAddress: string;
  denomination: string;
};

export const ModalCreatePhonon: React.FC<{ card; isOpen; onClose }> = ({
  card,
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreatePhononFormData>();

  // event when you create a phonon
  const onSubmit = (data: CreatePhononFormData, event) => {
    event.preventDefault();

    console.log(data);

    onClose();
    reset();

    notifySuccess(t('Phonon created!'));
  };

  return (
    <Modal size="lg" isOpen={isOpen} onClose={onClose}>
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent>
        <ModalHeader>
          <div className="font-noto-sans-mono">
            <div className="text-sm">{t('Create Phonon for')}</div>
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
            <div className="grid grid-cols-1 gap-y-6">
              <FormControl>
                <FormLabel>{t('Token Address')}</FormLabel>
                <Input
                  bg="gray.700"
                  color="white"
                  type="text"
                  maxLength={20}
                  placeholder="0x..."
                  {...register('tokenAddress', {
                    required: t('Token Address is required.'),
                  })}
                />
                {errors.tokenAddress && (
                  <span className="text-red-600">
                    {errors.tokenAddress.message}
                  </span>
                )}
                <FormHelperText>
                  {t(
                    'This is the token contract address that you would like to create.'
                  )}
                </FormHelperText>
              </FormControl>
              <FormControl>
                <FormLabel>{t('Denomination')}</FormLabel>
                <Input
                  bg="gray.700"
                  color="white"
                  type="number"
                  maxLength={20}
                  placeholder="0.00"
                  {...register('denomination', {
                    required: t('Denomination is required.'),
                  })}
                />
                {errors.denomination && (
                  <span className="text-red-600">
                    {errors.denomination.message}
                  </span>
                )}
                <FormHelperText>
                  {t(
                    'This is the token contract address that you would like to create.'
                  )}
                </FormHelperText>
              </FormControl>
            </div>
          </ModalBody>

          <ModalFooter>
            <Button colorScheme="green" type="submit" mr={3}>
              {t('Create New Phonon')}
            </Button>
            <Button onClick={onClose}>{t('Cancel')}</Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
