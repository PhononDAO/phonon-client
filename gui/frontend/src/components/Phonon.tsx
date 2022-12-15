import { Phonon as iPhonon, PhononCard } from '../interfaces/interfaces';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { ChainIDTag } from './ChainIDTag';
import { CURRENCIES } from '../constants/Currencies';
import { useDrag } from 'react-dnd';
import { IonIcon } from '@ionic/react';
import { closeCircle } from 'ionicons/icons';
import { useContext, useEffect } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { getEmptyImage } from 'react-dnd-html5-backend';
import { ModalPhononDetails } from './ModalPhononDetails';
import { useDisclosure } from '@chakra-ui/react';

interface DropResult {
  name: string;
  type: string;
}

export const Phonon: React.FC<{
  phonon: iPhonon;
  destinationCard?: PhononCard;
  layoutType?: string;
  isProposed?: boolean;
  showAction?: boolean;
  isCustomDragLayer?: boolean;
}> = ({
  phonon,
  destinationCard = null,
  layoutType = 'list',
  isProposed = false,
  showAction = false,
  isCustomDragLayer = false,
}) => {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { getCardById, removePhononsFromCardTransferState } = useContext(
    CardManagementContext
  );

  const card: PhononCard = destinationCard
    ? destinationCard
    : getCardById(phonon.SourceCardId);

  const [{ isDragging }, drag, preview] = useDrag(() => ({
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

  useEffect(() => {
    preview(getEmptyImage(), { captureDraggingState: true });
  }, [preview]);

  const removeFromProposal = () => {
    removePhononsFromCardTransferState(
      card,
      [phonon],
      'OutgoingTransferProposal'
    );
  };

  // require a double-click to open the phonon modal details
  const openPhononModal = (e) => {
    if (e.detail === 2) {
      onOpen();
    }
  };

  return (
    <>
      <div
        ref={phonon.ProposedForTransfer ? null : drag}
        className={
          'transition-all duration-300 rounded-full overflow-hidden' +
          (phonon.ProposedForTransfer
            ? ''
            : ' hover:shadow-md hover:shadow-zinc-800/80') +
          (layoutType === 'grid' ? ' inline-block relative w-1/4' : ' w-full') +
          (isCustomDragLayer ? ' -rotate-3 w-2/5' : '')
        }
        onClick={
          !phonon.ProposedForTransfer || !isProposed ? openPhononModal : null
        }
      >
        {layoutType === 'grid' && <div className="mt-full"></div>}
        <div
          className={
            'transition-all rounded-full px-4 py-2 bg-black' +
            (phonon.ProposedForTransfer && !isProposed ? ' opacity-5' : '') +
            (isDragging ? ' opacity-20' : '') +
            (layoutType === 'grid'
              ? ' absolute top-0 right-1 bottom-0 left-1 pt-12'
              : ' flex items-center gap-x-8')
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
            {abbreviateHash(phonon.PubKey)}
          </div>
          {showAction && (
            <IonIcon
              icon={closeCircle}
              className="text-red-500 bg-white rounded-full cursor-pointer"
              onClick={removeFromProposal}
            />
          )}
        </div>
      </div>
      <ModalPhononDetails phonon={phonon} isOpen={isOpen} onClose={onClose} />
    </>
  );
};
