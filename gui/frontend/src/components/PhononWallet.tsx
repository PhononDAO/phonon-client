import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard as Card } from '../classes/PhononCard';
import { Phonon as PhononObj } from '../classes/Phonon';
import { AddMockCardButton } from './AddMockCardButton';
import { Button } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { removeCircle, addCircle } from 'ionicons/icons';
import { WalletSlot } from './WalletSlot';

export const PhononWallet = () => {
  const { t } = useTranslation();

  const aPhonon = new PhononObj();
  aPhonon.Address = '0x7Ab7050217C76d729fa542161ca59Cb28654bf80';
  aPhonon.ChainID = '3';
  aPhonon.Denomination = '40000000000000000';
  aPhonon.CurrencyType = 2;

  const bPhonon = new PhononObj();
  bPhonon.Address = '0x7Ab7050217C76d729fa542161ca59Cb28484bf8e';
  bPhonon.ChainID = '137';
  bPhonon.Denomination = '50600000000000000';
  bPhonon.CurrencyType = 2;

  const aCard = new Card();
  aCard.CardId = '04e0d5eb884a73cf';
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(bPhonon);

  const bCard = new Card();
  bCard.CardId = '04e0d5eb884a73ce';
  bCard.VanityName = 'my favorite card';
  const cCard = new Card();
  cCard.CardId = '04e0d5eb884a73c0';
  bCard.Phonons.push(aPhonon);
  bCard.Phonons.push(bPhonon);

  const [phononWallet] = useState<Array<Card>>([aCard, bCard, cCard]);

  const [hideCards, setHideCards] = useState<boolean>(false);
  const [isCardsMini, setIsCardsMini] = useState<boolean>(false);

  const toggleCardVisibility = () => {
    setHideCards((prev) => !prev);
  };

  return (
    <div className="">
      <div className="flex gap-x-2 text-xl">
        <span className="text-white">
          {phononWallet.length}{' '}
          {t('card' + (phononWallet.length === 1 ? '' : 's') + ' connected.')}
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
          (hideCards ? 'h-0 mb-0' : 'mb-2 ' + (isCardsMini ? 'h-44' : 'h-60'))
        }
      >
        {phononWallet.length > 0 &&
          phononWallet?.map((card, key) => (
            <WalletSlot key={key} card={card} isMini={isCardsMini} />
          ))}

        <AddMockCardButton />
      </div>
    </div>
  );
};
