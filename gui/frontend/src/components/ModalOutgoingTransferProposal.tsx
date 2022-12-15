import {
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { send } from 'ionicons/icons';
import { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { notifySuccess } from '../utils/notify';
import { Card } from './Card';
import { Phonon } from './Phonon';

export const ModalOutgoingTransferProposal: React.FC<{
  destinationCard: PhononCard;
  isOpen;
  onClose;
}> = ({ destinationCard, isOpen, onClose }) => {
  const { t } = useTranslation();
  const {
    getCardById,
    resetPhononsOnCardTransferState,
    updateCardTransferStatusState,
  } = useContext(CardManagementContext);
  const [isTransferred, setIsTransferred] = useState(false);

  const sourceCard = getCardById(
    destinationCard?.OutgoingTransferProposal.Phonons[0].SourceCardId
  );

  const closeTransfer = () => {
    updateCardTransferStatusState(
      destinationCard,
      'OutgoingTransferProposal',
      'unvalidated'
    );
    onClose();

    // let's clear the incoming transfer proposal
    resetPhononsOnCardTransferState(
      destinationCard,
      'OutgoingTransferProposal'
    );
  };

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const promise = new Promise((resolve) => {
    setTimeout(() => {
      resolve('paired');
    }, 15 * 1000);
  }).then(() => {
    updateCardTransferStatusState(
      destinationCard,
      'OutgoingTransferProposal',
      'transferred'
    );

    setIsTransferred(true);
  });

  useEffect(() => {
    if (isTransferred) {
      notifySuccess(
        t(
          'Successfully transferred {{phononCount}} phonons from {{sourceCardId}} → {{destinationCardId}}',
          {
            phononCount:
              destinationCard.OutgoingTransferProposal.Phonons.length,
            sourceCardId: sourceCard.CardId,
            destinationCardId: destinationCard.CardId,
          }
        )
      );
    }
  }, [isTransferred]);

  return (
    <Modal
      isOpen={isOpen}
      size="4xl"
      onClose={closeTransfer}
      closeOnOverlayClick={
        destinationCard.OutgoingTransferProposal.Status === 'transferred'
      }
    >
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent>
        <ModalHeader>
          <span className="text-5xl font-bandeins-sans-light">
            {t('Outgoing Phonons')}
          </span>
        </ModalHeader>
        {destinationCard.OutgoingTransferProposal.Status === 'transferred' && (
          <ModalCloseButton />
        )}
        <ModalBody pb={6}>
          <div className="relative">
            <div className="absolute flex justify-center w-full z-10">
              <div className="relative grid grid-row-1 content-center text-green-700 w-2/3 h-36">
                {destinationCard.OutgoingTransferProposal.Status ===
                'transferred' ? (
                  <>
                    <IonIcon
                      icon={send}
                      className="mx-auto -rotate-30 text-5xl"
                    />
                    <div className="mt-4 text-sm text-center">
                      {t('Phonons Transferred Successfully!')}
                    </div>
                  </>
                ) : (
                  <>
                    <div className="flex justify-content align-items">
                      <IonIcon icon={send} className="animate-outgoing" />
                      <IonIcon
                        icon={send}
                        className="animate-outgoing animation-delay-1"
                      />
                      <IonIcon
                        icon={send}
                        className="animate-outgoing animation-delay-2"
                      />
                      <IonIcon
                        icon={send}
                        className="animate-outgoing animation-delay-3"
                      />
                    </div>
                    <div className="mt-4 text-sm text-center">
                      {t('Sending Phonons')}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
          <div className="relative flex justify-between z-30">
            <div className="relative w-56 h-36">
              <Card card={sourceCard} isMini={true} showActions={false} />
            </div>
            <div className="relative w-56 h-36">
              <Card card={destinationCard} isMini={true} showActions={false} />
            </div>
          </div>

          <h3 className="mt-8 mb-2 text-xl text-gray-500">
            {destinationCard.OutgoingTransferProposal.Status === 'transferred'
              ? t('The following Phonons were transferred:')
              : t('The following Phonons are being transferred:')}
          </h3>
          <div
            className={
              'overflow-scroll gap-2 grid w-full' +
              (destinationCard.OutgoingTransferProposal.Status === 'transferred'
                ? ''
                : ' animate-pulse opacity-60')
            }
          >
            {destinationCard.OutgoingTransferProposal?.Phonons?.map(
              (phonon, key) => (
                <Phonon
                  key={key}
                  phonon={phonon}
                  isProposed={true}
                  showAction={false}
                />
              )
            )}
          </div>
        </ModalBody>

        <ModalFooter>
          {destinationCard.OutgoingTransferProposal.Status === 'transferred' ? (
            <Button onClick={closeTransfer}>{t('Close')}</Button>
          ) : (
            <Button colorScheme="red" onClick={closeTransfer}>
              {t('Cancel Transfer')}
            </Button>
          )}
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
