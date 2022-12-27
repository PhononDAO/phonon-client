import { IonIcon } from '@ionic/react';
import { addCircle } from 'ionicons/icons';
import { useTranslation } from 'react-i18next';
import { useDisclosure } from '@chakra-ui/react';
import { ModalCreateMockCard } from './ModalCreateMockCard';
import { useFeature } from '../hooks/useFeature';

export const AddMockCardButton: React.FC = () => {
  const { t } = useTranslation();
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { ENABLE_MOCK_CARDS } = useFeature();

  return ENABLE_MOCK_CARDS ? (
    <>
      <button
        onClick={onOpen}
        title={t('Create Mock Card')}
        className="text-zinc-800 hover:text-zinc-500 relative flex items-center justify-center"
      >
        <div className="absolute m-auto rounded-full w-8 h-8 bg-red-600"></div>
        <IonIcon className="rounded-full text-6xl" icon={addCircle} />
      </button>
      <ModalCreateMockCard isOpen={isOpen} onClose={onClose} />
    </>
  ) : (
    <></>
  );
};
