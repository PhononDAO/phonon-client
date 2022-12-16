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
import { Phonon as iPhonon, PhononCard } from '../interfaces/interfaces';
import { useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { CURRENCIES } from '../constants/Currencies';

type CreatePhononFormData = {
  tokenAddress: string;
  denomination: string;
};

export const ModalCreatePhonon: React.FC<{
  card: PhononCard;
  isOpen;
  onClose;
}> = ({ card, isOpen, onClose }) => {
  const { t } = useTranslation();
  const { addPhononsToCardState } = useContext(CardManagementContext);
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

    const nPhonon = {
      PubKey:
        '04fc53a5e843e76cac55e7ce43d7592fb5002a749832b1f65708e84108e958fe6cfdd459f144ccb7739f947c1f317e9cfaa1c40bd138358e155afffdd626d43ff9',
      Address: '',
      AddressType: 0,
      SchemaVersion: 0,
      ExtendedSchemaVersion: 0,
      CurveType: 0,
      ChainID: 137,
      Denomination: '7903000000000000000',
      CurrencyType: 2,
      SourceCardId: card.CardId,
      ValidationStatus: 'unvalidated',
    } as iPhonon;

    addPhononsToCardState(card, [nPhonon]);

    onClose();
    reset();

    notifySuccess(
      t(
        'Phonon "{{phononPubKey}}" in the amount of {{amount}}{{ticker}} was created!',
        {
          phononPubKey: abbreviateHash(nPhonon.PubKey),
          amount: fromDecimals(
            nPhonon.Denomination,
            CURRENCIES[nPhonon.CurrencyType].decimals
          ),
          ticker: CURRENCIES[nPhonon.CurrencyType].ticker,
        }
      )
    );
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
