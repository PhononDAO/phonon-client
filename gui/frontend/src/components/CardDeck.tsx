import { useTranslation } from 'react-i18next';
import { PhononCard as Card } from '../classes/PhononCard';
import { CardTray } from './CardTray';

export const CardDeck: React.FC<{
  card?: Card;
}> = ({ card = null }) => {
  const { t } = useTranslation();

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div className="text-white">
      <CardTray />
    </div>
  );
};
