import { useTranslation } from 'react-i18next';
import { ButtonGroup, IconButton, Select } from '@chakra-ui/react';
import { CardTray } from './CardTray';
import { Phonon } from './Phonon';
import { IonIcon } from '@ionic/react';
import { reorderFour, apps, bulb } from 'ionicons/icons';
import { useContext, useEffect, useState } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import {
  GlobalSettings,
  Phonon as iPhonon,
  PhononCard,
} from '../interfaces/interfaces';
import { MinePhonon } from './PhononCardActions/MinePhonon';
import { CreatePhonon } from './PhononCardActions/CreatePhonon';
import { RemoteCardPhononMessage } from './RemoteCardPhononMessage';
import { PhononTransferProposal } from './PhononTransferProposal';
import { IncomingTransferNotice } from './IncomingTransferNotice';
import localStorage from '../utils/localStorage';

export const CardDeck: React.FC<{
  card: PhononCard;
  canHaveRemote?: boolean;
}> = ({ card, canHaveRemote }) => {
  const { t } = useTranslation();
  const [layoutType, setLayoutType] = useState<string>('list');
  const { phononCards, addCardsToState, addPhononsToCardTransferState } =
    useContext(CardManagementContext);
  const defaultSettings: GlobalSettings =
    localStorage.getConfigurableSettings();

  // let's poll for updates on this card
  if (false) {
    const simulateIncomingRequest = setInterval(() => {
      // let's fake an incoming proposal
      if (
        phononCards.filter((card: PhononCard) => card.InTray).length > 1 &&
        card?.CardId === '04e0d5eb884a73cf'
      ) {
        const aPhonon = {
          PubKey:
            '04351ed0872482a41bd005d886b7151f40dd691f1efc8b03d9f5f24e9bee95afb01ef84c3d1515b9ed1c48cb86c14290ee0659233899d4387ad73bbff7bac7326d',
          Address: '',
          AddressType: 0,
          SchemaVersion: 0,
          ExtendedSchemaVersion: 0,
          CurveType: 0,
          ChainID: 3,
          Denomination: '40000000000000000',
          CurrencyType: 2,
          SourceCardId: '04e0d5eb884a73e9',
          ValidationStatus: 'unvalidated',
        } as iPhonon;

        const bPhonon = {
          PubKey:
            '04fc53a5e843e76cac55e7ce43d7592fb9523a749832b1f65708e84108e958fe6cfdd459f144ccb7739f947c1f317e9cfaa1c40bd138358e155afffdd626d0303e',
          Address: '',
          AddressType: 0,
          SchemaVersion: 0,
          ExtendedSchemaVersion: 0,
          CurveType: 0,
          ChainID: 137,
          Denomination: '50600000000000000',
          CurrencyType: 2,
          SourceCardId: '04e0d5eb884a73e9',
          ValidationStatus: 'unvalidated',
        } as iPhonon;

        addPhononsToCardTransferState(
          card,
          [aPhonon, bPhonon],
          'IncomingTransferProposal'
        );

        clearInterval(simulateIncomingRequest);
      }
    }, 15 * 1000);
  }

  const sortPhononsBy = (key: string) => {
    if (card?.Phonons) {
      if (key === 'ChainId') {
        card.Phonons.sort((a, b) => a.ChainID - b.ChainID);
      } else if (key === 'Denomination') {
        card.Phonons.sort((a, b) =>
          a.Denomination.localeCompare(b.Denomination)
        );
      } else if (key === 'CurrencyType') {
        card.Phonons.sort((a, b) => a.CurrencyType - b.CurrencyType);
      }
      addCardsToState([card]);
    }
  };

  useEffect(() => {
    sortPhononsBy(defaultSettings.defaultPhononSortBy);

    setLayoutType(defaultSettings.defaultPhononLayout);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div
      className={
        'relative w-full p-4 rounded-sm mt-40 pt-24 ' +
        (card ? 'bg-gray-300' : '')
      }
    >
      <div className="absolute -mt-60">
        <CardTray card={card} canHaveRemote={canHaveRemote} />
      </div>

      {card && (
        <>
          {!card.IsRemote && (
            <>
              <div className="absolute -top-16 right-0 flex gap-x-4">
                <MinePhonon card={card} />
                <CreatePhonon card={card} />
              </div>

              <div className="absolute top-0 right-0 p-4 flex gap-x-4">
                <div className="flex items-center">
                  <div className="whitespace-nowrap mr-2 text-lg text-gray-600">
                    {t('Sort by')}:
                  </div>
                  <Select
                    onChange={(evt) => {
                      sortPhononsBy(evt.target.value);
                    }}
                    defaultValue={defaultSettings.defaultPhononSortBy}
                  >
                    <option value="Key">{t('Key')}</option>
                    <option value="ChainId">{t('Network Chain')}</option>
                    <option value="Denomination">{t('Denomination')}</option>
                    <option value="CurrencyType">{t('Currency Type')}</option>
                  </Select>
                </div>
                <div className="rounded flex">
                  <ButtonGroup isAttached>
                    <IconButton
                      bgColor={layoutType === 'list' ? 'black' : 'white'}
                      textColor={layoutType === 'list' ? 'white' : 'black'}
                      aria-label={t('List View')}
                      icon={<IonIcon icon={reorderFour} />}
                      onClick={() => {
                        setLayoutType('list');
                      }}
                    />
                    <IconButton
                      bgColor={layoutType === 'grid' ? 'black' : 'white'}
                      textColor={layoutType === 'grid' ? 'white' : 'black'}
                      aria-label={t('Grid View')}
                      icon={<IonIcon icon={apps} />}
                      onClick={() => {
                        setLayoutType('grid');
                      }}
                    />
                  </ButtonGroup>
                </div>
              </div>
            </>
          )}
          {card?.IncomingTransferProposal?.Phonons?.length > 0 && (
            <IncomingTransferNotice card={card} />
          )}
          {(phononCards.filter(
            (card: PhononCard) => card.InTray && !card.IsRemote
          ).length > 1 ||
            card.IsRemote) && <PhononTransferProposal card={card} />}
          {!card.IsRemote && card?.Phonons.length > 0 && (
            <div className="flex justify-end text-gray-500 items-center mb-2">
              <>
                <IonIcon icon={bulb} />
                {t('Double-click Phonons to see details.')}
              </>
            </div>
          )}
          {!card.IsRemote ? (
            <div
              className={
                'overflow-visible gap-2 ' +
                (layoutType === 'grid' ? 'relative' : 'grid')
              }
            >
              {card.Phonons.length > 0 ? (
                card.Phonons?.map((phonon, key) => (
                  <Phonon key={key} phonon={phonon} layoutType={layoutType} />
                ))
              ) : (
                <div className="text-2xl text-center my-12 italic text-gray-500">
                  {t('This card has no phonons yet.')}
                </div>
              )}
            </div>
          ) : (
            <RemoteCardPhononMessage />
          )}
        </>
      )}
    </div>
  );
};
