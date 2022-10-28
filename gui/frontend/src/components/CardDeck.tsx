import { useTranslation } from 'react-i18next';
import { ButtonGroup, IconButton, Select } from '@chakra-ui/react';
import { PhononCard } from '../classes/PhononCard';
import { Phonon as PhononObj } from '../classes/Phonon';
import { CardTray } from './CardTray';
import { Phonon } from './Phonon';
import { IonIcon } from '@ionic/react';
import { reorderFour, apps } from 'ionicons/icons';
import { useState } from 'react';

export const CardDeck: React.FC<{
  card?: PhononCard;
}> = ({ card = null }) => {
  const { t } = useTranslation();
  const [layoutType, setLayoutType] = useState<string>('list');

  const aCard = new PhononCard();
  aCard.CardId = '04e0d5eb884a73cf';

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

  const cPhonon = new PhononObj();
  cPhonon.Address = '0x7Ab7050217C76d729fa542161ca59Cb28484ee04';
  cPhonon.ChainID = '43114';
  cPhonon.Denomination = '3100000000000000000';
  cPhonon.CurrencyType = 2;

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

  const [thisCard, setThisCard] = useState(aCard);

  const sortPhononsBy = (key: string) => {
    if (key === 'ChainId') {
      thisCard.Phonons.sort((a, b) => a.ChainID.localeCompare(b.ChainID));
    } else if (key === 'Denomination') {
      thisCard.Phonons.sort((a, b) =>
        a.Denomination.localeCompare(b.Denomination)
      );
    } else if (key === 'CurrencyType') {
      thisCard.Phonons.sort((a, b) => a.CurrencyType - b.CurrencyType);
    }

    setThisCard(thisCard);
  };

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div className="relative w-full p-4 rounded-sm mt-40 pt-24 bg-gray-300">
      <div className="absolute -mt-60">
        <CardTray />
      </div>
      <div className="absolute top-0 right-0 p-4 flex gap-x-4">
        <div className="flex items-center">
          <div className="whitespace-nowrap mr-2 text-lg text-gray-600">
            Sort by:
          </div>
          <Select
            placeholder="Select order"
            onChange={(evt) => {
              sortPhononsBy(evt.target.value);
            }}
          >
            <option value="ChainId">Network Chain</option>
            <option value="Denomination">Denomination</option>
            <option value="CurrencyType">CurrencyType</option>
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
        {thisCard.Phonons.length > 0 ? (
          thisCard.Phonons?.map((phonon, key) => (
            <Phonon key={key} phonon={phonon} layoutType={layoutType} />
          ))
        ) : (
          <div className="text-2xl text-center my-12 italic text-gray-500">
            This card has no phonons yet.
          </div>
        )}
      </div>
    </div>
  );
};
