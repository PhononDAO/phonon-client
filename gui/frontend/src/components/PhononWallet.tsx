import { useState, useContext, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Phonon, PhononCard } from '../interfaces/interfaces';
import { AddMockCardButton } from './AddMockCardButton';
import { Button } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { removeCircle, addCircle } from 'ionicons/icons';
import { WalletSlot } from './WalletSlot';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { Pluralize } from 'pluralize-react';

export const PhononWallet = () => {
  const { t } = useTranslation();
  const { phononCards, addCardsToState, isCardsMini } = useContext(
    CardManagementContext
  );

  const aPhonon = {
    PubKey:
      '04351ed0872482a41bd005d886b7151f40dd691f1efc8b03d9f5f24d9bee80afb01ef84c3d1515b9ed1c48cb86c14290ee0659233899d4387ad73bbff7bac7326d',
    Address: '',
    AddressType: 0,
    SchemaVersion: 0,
    ExtendedSchemaVersion: 0,
    CurveType: 0,
    ChainID: 3,
    Denomination: '40000000000000000',
    CurrencyType: 2,
    SourceCardId: '04e0d5eb884a73cf',
    ValidationStatus: 'unvalidated',
  } as Phonon;

  const bPhonon = {
    PubKey:
      '04fc53a5e843e76cac55e7ce43d7592fb9523a749832b1f65702783108e858fe6cfdd459f144ccb7739f947c1f317e9cfaa1c40bd138358e155afffdd626d0303e',
    Address: '',
    AddressType: 0,
    SchemaVersion: 0,
    ExtendedSchemaVersion: 0,
    CurveType: 0,
    ChainID: 137,
    Denomination: '50600000000000000',
    CurrencyType: 2,
    SourceCardId: '04e0d5eb884a73cf',
    ValidationStatus: 'unvalidated',
  } as Phonon;

  const cPhonon = {
    PubKey:
      '0406fae3f294d33c5b67c6d66199e9eb6b4c735efbb7a1e1677dfd681165be3fa533a38a0e301f9f0022ff6cc2620075d3e3e751f5f969caf7cf90d8f3484ff072',
    Address: '',
    AddressType: 0,
    SchemaVersion: 0,
    ExtendedSchemaVersion: 0,
    CurveType: 0,
    ChainID: 43114,
    Denomination: '3100000000000000000',
    CurrencyType: 3,
    SourceCardId: '04e0d5eb884a73cf',
    ValidationStatus: 'unvalidated',
  } as Phonon;

  const dPhonon = {
    PubKey:
      '046d623e328c1b2e618f9131767dc92f33651e71538cccc7f0c8efe879cc292e663d2bf566fe99c1065406e158c83b710eadc733a802a54053242097a637916862',
    Address: '',
    AddressType: 0,
    SchemaVersion: 0,
    ExtendedSchemaVersion: 0,
    CurveType: 0,
    ChainID: 43114,
    Denomination: '5008000000000000000',
    CurrencyType: 3,
    SourceCardId: '04e0d5eb884a73ce',
    ValidationStatus: 'unvalidated',
  } as Phonon;

  const aCard = {
    CardId: '04e0d5eb884a73cf',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;
  aCard.Phonons.push(aPhonon);
  aCard.Phonons.push(bPhonon);
  aCard.Phonons.push(cPhonon);

  const bCard = {
    CardId: '04e0d5eb884a73ce',
    VanityName: 'my favorite card',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;
  bCard.Phonons.push(dPhonon);

  const cCard = {
    CardId: '04e0d5eb884a73c0',
    IsLocked: true,
    ShowActions: true,
    Phonons: [],
  } as PhononCard;

  useEffect(() => {
    addCardsToState([aCard, bCard, cCard]);
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
          <Pluralize
            count={phononCards.filter((card) => !card.IsRemote).length}
            singular={t('card')}
            plural={t('cards')}
            className="font-bandeins-sans-bold text-xl"
          />{' '}
          {t('in your wallet')}
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
