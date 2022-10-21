import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard as Card } from '../classes/PhononCard';
import { AddMockCardButton } from './AddMockCardButton';
import { Button } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { removeCircle, addCircle } from 'ionicons/icons';
import { PhononCard } from './PhononCard';

export const PhononWallet = () => {
  const { t } = useTranslation();

  const aCard = new Card();
  aCard.CardId = '04e0d5eb884a73cf';
  const bCard = new Card();
  bCard.CardId = '04e0d5eb884a73ce';
  bCard.VanityName = 'my favorite card';
  const cCard = new Card();
  cCard.CardId = '04e0d5eb884a73c0';

  const [phononWallet] = useState<Array<Card>>([aCard, bCard, cCard]);

  const [hideCards, setHideCards] = useState<boolean>(false);
  const [isCardsMini, setIsCardsMini] = useState<boolean>(false);

  const toggleCardVisibility = () => {
    setHideCards((prev) => !prev);
  };

  return (
    <>
      <div className="flex gap-x-2 mb-4 text-xl">
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
          'relative flex space-x-10 overflow-x-auto pb-4 transition-transform duration-300 ease-out overflow-hidden origin-top transform ' +
          (hideCards ? 'scale-y-0' : '')
        }
      >
        {phononWallet.length > 0 &&
          phononWallet?.map((card, key) => (
            <PhononCard key={key} card={card} isMini={isCardsMini} />
          ))}

        <AddMockCardButton />
      </div>
    </>
  );
};
