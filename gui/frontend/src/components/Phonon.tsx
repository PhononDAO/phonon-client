import { useTranslation } from 'react-i18next';
import { Phonon as aPhonon } from '../classes/Phonon';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { ChainIDTag } from './ChainIDTag';
import { CURRENCIES } from '../constants/Currencies';

export const Phonon: React.FC<{
  phonon: aPhonon;
  layoutType?: string;
}> = ({ phonon = null, layoutType = 'list' }) => {
  const { t } = useTranslation();

  return (
    phonon && (
      <div
        className={
          'transition-all duration-300 ' +
          (layoutType === 'grid' ? 'inline-block relative w-1/4' : 'w-full')
        }
      >
        {layoutType === 'grid' && <div className="mt-full"></div>}
        <div
          className={
            'px-4 py-2 rounded-full bg-black ' +
            (layoutType === 'grid'
              ? 'absolute top-0 right-1 bottom-0 left-1 pt-12'
              : 'flex items-center gap-x-8')
          }
        >
          <div
            className={
              'flex ' +
              (layoutType === 'grid' ? 'justify-center mb-2' : 'w-32 ')
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
    )
  );
};
