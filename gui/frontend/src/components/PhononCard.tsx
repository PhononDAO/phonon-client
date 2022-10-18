import { IonIcon } from '@ionic/react';
import { lockClosed } from 'ionicons/icons';
import { useTranslation } from 'react-i18next';
import { PhononCard as Card } from '../classes/PhononCard';
import { HelpTooltip } from './HelpTooltip';
import { useFeature } from '../hooks/useFeature';

export const PhononCard: React.FC<{
  card: Card;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { ENABLE_MOCK_CARDS } = useFeature();

  // only show card if not a mock card or if mock cards are enabled
  return (card.IsMock && ENABLE_MOCK_CARDS) || !card.IsMock ? (
    <div>
      <div className="w-64 h-40 bg-white relative rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
        <div className="absolute w-full h-full bg-black opacity-70"></div>
        <div className="absolute w-full h-full">
          <div className="absolute z-50 w-full h-full p-2 font-noto-sans-mono text-base">
            {card.CardId}
            <img
              className="absolute w-20 right-10 bottom-4"
              src="/assets/images/phonon-logo.png"
            />
          </div>
          <div className="absolute h-full relative flex items-center">
            <div className="absolute text-center -right-14 font-bandeins-sans-bold text-phonon-card uppercase rotate-90">
              PHONON
            </div>
            {card.IsLocked && (
              <div className="w-full text-amber-400 text-center">
                <IonIcon className="text-6xl" icon={lockClosed} />
              </div>
            )}
          </div>
          {card.IsMock && (
            <div className="absolute w-60 rotate-30 top-28 -left-16 font-bandeins-sans-bold text-md text-center bg-red-600 py-px">
              {t('MOCK CARD')}
            </div>
          )}
        </div>
      </div>
      {card.IsMock && (
        <div className="w-64 pt-px flex justify-end">
          <HelpTooltip text={t('What is a mock card?')} theme="error">
            {t(
              'A mock card is a temporary card to test the platform. Mock cards are deleted, including all phonons, when this app is closed and have a different certificate than alpha and testnet phonon cards and therefore cannot communicate with them.'
            )}
          </HelpTooltip>
        </div>
      )}
    </div>
  ) : (
    <></>
  );
};
