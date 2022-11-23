import { useState, useContext, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Phonon, PhononCard } from '../interfaces/interfaces';
import { AddMockCardButton } from './AddMockCardButton';
import { Button } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { removeCircle, addCircle } from 'ionicons/icons';
import { WalletSlot } from './WalletSlot';
import { CardManagementContext } from '../contexts/CardManagementContext';

export const PhononWallet = () => {
  const { t } = useTranslation();
  const {
    phononCards,
    addPhononCardsToState,
    addCardPhononsToState,
    isCardsMini,
  } = useContext(CardManagementContext);

  const aPhonon = {
    Address: '0x7Ab7050217C76d729fa542161ca59Cb28654bf80',
    ChainID: 3,
    Denomination: '40000000000000000',
    CurrencyType: 2,
  } as Phonon;

  const bPhonon = {
    Address: '0x7Ab7050217C76d729fa542161ca59Cb28484bf8e',
    ChainID: 137,
    Denomination: '50600000000000000',
    CurrencyType: 2,
  } as Phonon;

  const cPhonon = {
    Address: '0x7Ab7050217C76d729fa542161ca59Cb28484ee04',
    ChainID: 43114,
    Denomination: '3100000000000000000',
    CurrencyType: 3,
  } as Phonon;

  const aCard = {
    CardId: '04e0d5eb884a73cf',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(cPhonon);
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(cPhonon);

  const bCard = {
    CardId: '04e0d5eb884a73ce',
    VanityName: 'my favorite card',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;
  bCard.Phonons.push(aPhonon);
  bCard.Phonons.push(bPhonon);

  const cCard = {
    CardId: '04e0d5eb884a73c0',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;

  useEffect(() => {
    addPhononCardsToState([aCard, bCard, cCard]);

    addCardPhononsToState(aCard.CardId, aCard.Phonons);
    addCardPhononsToState(bCard.CardId, bCard.Phonons);
    addCardPhononsToState(cCard.CardId, cCard.Phonons);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [hideCards, setHideCards] = useState<boolean>(false);

  const toggleCardVisibility = () => {
    setHideCards((prev) => !prev);
  };

  return (
    <div className="">
      <div className="flex gap-x-2 text-xl">
        <span className="text-white">
          {phononCards.filter((card) => !card.IsRemote).length}{' '}
          {t(
            'card' +
              (phononCards.filter((card) => !card.IsRemote).length === 1
                ? ''
                : 's') +
              ' connected.'
          )}
        </span>
        <Button
          leftIcon={<IonIcon icon={hideCards ? addCircle : removeCircle} />}
          size="xs"
          colorScheme="gray"
          className="uppercase"
          onClick={toggleCardVisibility}
        >
          {hideCards ? t('Show Cards') : t('Hide Cards')}
        </Button>
      </div>

      <div
        className={
          'relative py-4 flex space-x-10 overflow-x-auto transition-all duration-300 ease-out overflow-hidden ' +
          (hideCards
            ? 'h-0 mb-0 py-0'
            : 'mb-2 ' + (isCardsMini ? 'h-44' : 'h-60'))
        }
      >
        {phononCards.length > 0 &&
          phononCards
            ?.filter((card) => !card.IsRemote)
            .map((card, key) => <WalletSlot key={key} card={card} />)}

        <AddMockCardButton />
      </div>
    </div>
  );
};
