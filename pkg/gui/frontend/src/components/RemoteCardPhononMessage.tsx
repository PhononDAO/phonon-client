import { IonIcon } from '@ionic/react';
import { cloud } from 'ionicons/icons';
import { useTranslation } from 'react-i18next';

export const RemoteCardPhononMessage = () => {
  const { t } = useTranslation();

  return (
    <div className="text-xl px-6 py-8 flex flex-col gap-y-2 items-center bg-black rounded-md text-white text-center font-bandeins-sans-light">
      <IonIcon icon={cloud} className="text-white text-7xl" />
      <span>
        {t(
          'This card is paired remotely. Phonons on remote cards do not show.'
        )}
      </span>
    </div>
  );
};
