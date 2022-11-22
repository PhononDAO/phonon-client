import { Phonon as aPhonon } from '../classes/Phonon';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { ChainIDTag } from './ChainIDTag';
import { CURRENCIES } from '../constants/Currencies';
import { useDrag } from 'react-dnd';
import { PhononCard } from '../classes/PhononCard';

interface DropResult {
  name: string;
  type: string;
}

export const Phonon: React.FC<{
  card: PhononCard;
  phonon: aPhonon;
  layoutType?: string;
}> = ({ phonon, card, layoutType = 'list' }) => {
  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'Phonon-' + card.CardId,
    name: phonon.Address,
    item: phonon,
    end: (item, monitor) => {
      const dropResult = monitor.getDropResult<DropResult>();
      if (item && dropResult) {
        // item.TrayId = true;
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }));

  return (
    <div
      ref={drag}
      className={
        'transition-all duration-300 rounded-full overflow-hidden hover:shadow-md hover:shadow-zinc-800/80' +
        (layoutType === 'grid' ? ' inline-block relative w-1/4' : ' w-full')
      }
    >
      {layoutType === 'grid' && <div className="mt-full"></div>}
      <div
        className={
          'rounded-full px-4 py-2 bg-black' +
          (isDragging ? ' opacity-10' : '') +
          (layoutType === 'grid'
            ? ' absolute top-0 right-1 bottom-0 left-1 pt-12'
            : ' flex items-center gap-x-8')
        }
      >
        <div
          className={
            'flex ' + (layoutType === 'grid' ? 'justify-center mb-2' : 'w-32 ')
          }
        >
          <ChainIDTag id={phonon.ChainID} />
        </div>
        <div
          className={
            'text-3xl text-white font-bandeins-sans-bold ' +
            (layoutType === 'grid' ? 'text-center' : '')
          }
        >
          <>
            {fromDecimals(
              phonon.Denomination,
              CURRENCIES[phonon.CurrencyType].decimals
            )}
            <span className="text-base font-bandeins-sans-light ml-2">
              {CURRENCIES[phonon.CurrencyType].ticker}
            </span>
          </>
        </div>
        <div
          className={
            'text-gray-400 ml-auto ' +
            (layoutType === 'grid' ? 'text-xs text-center' : '')
          }
        >
          {abbreviateHash(phonon.Address)}
        </div>
      </div>
    </div>
  );
};
