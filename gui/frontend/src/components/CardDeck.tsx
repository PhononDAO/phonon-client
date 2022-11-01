import { useTranslation } from 'react-i18next';
import { ButtonGroup, IconButton, Select } from '@chakra-ui/react';
import { PhononCard } from '../classes/PhononCard';
import { CardTray } from './CardTray';
import { Phonon } from './Phonon';
import { IonIcon } from '@ionic/react';
import { reorderFour, apps } from 'ionicons/icons';
import { useContext, useState } from 'react';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

export const CardDeck: React.FC<{
  card: PhononCard;
  canHaveRemote?: boolean;
  setDeckCard;
}> = ({ card = null, canHaveRemote = false, setDeckCard }) => {
  const { t } = useTranslation();
  const [layoutType, setLayoutType] = useState<string>('list');
  const { addPhononCardsToState } = useContext(CardManagementContext);

  const sortPhononsBy = (key: string) => {
    if (key === 'ChainId') {
      card.Phonons.sort((a, b) => a.ChainID.localeCompare(b.ChainID));
    } else if (key === 'Denomination') {
      card.Phonons.sort((a, b) => a.Denomination.localeCompare(b.Denomination));
    } else if (key === 'CurrencyType') {
      card.Phonons.sort((a, b) => a.CurrencyType - b.CurrencyType);
    }

    addPhononCardsToState([card]);
  };

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div
      className={
        'relative w-full p-4 rounded-sm mt-40 pt-24 ' +
        (card ? 'bg-gray-300' : '')
      }
    >
      <div className="absolute -mt-60">
        <CardTray
          card={card}
          canHaveRemote={canHaveRemote}
          setDeckCard={setDeckCard}
        />
      </div>

      {card && (
        <>
          <div className="absolute top-0 right-0 p-4 flex gap-x-4">
            <div className="flex items-center">
              <div className="whitespace-nowrap mr-2 text-lg text-gray-600">
                {t('Sort by')}:
              </div>
              <Select
                placeholder="Select order"
                onChange={(evt) => {
                  sortPhononsBy(evt.target.value);
                }}
              >
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
          <div
            className={
              'overflow-scroll gap-2 ' +
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
        </>
      )}
    </div>
  );
};
