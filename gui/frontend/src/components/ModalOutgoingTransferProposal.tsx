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
  const { getCardById, resetPhononsOnCardTransferState } = useContext(
    CardManagementContext
  );
  const [transferComplete, setTransferComplete] = useState(false);

  const sourceCard = getCardById(
    destinationCard?.OutgoingTransferProposal[0].SourceCardId
  );

  const closeTransfer = () => {
    setTransferComplete(false);
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
    }, 8000);
  }).then(() => {
    setTransferComplete(true);
  });

  useEffect(() => {
    if (transferComplete) {
      notifySuccess(
        t(
          'Successfully transfered ' +
            String(destinationCard.OutgoingTransferProposal.length) +
            ' phonons from ' +
            String(sourceCard.CardId) +
            ' → ' +
            destinationCard.CardId
        )
      );
    }
  }, [destinationCard, sourceCard, t, transferComplete]);

  return (
    <Modal
      isOpen={isOpen}
      size="4xl"
      onClose={closeTransfer}
      closeOnOverlayClick={transferComplete}
    >
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <span className="text-5xl font-bandeins-sans-light">
            Outgoing Phonons
          </span>
        </ModalHeader>
        {transferComplete && <ModalCloseButton />}
        <ModalBody pb={6}>
          <div className="relative">
            <div className="absolute flex justify-center w-full z-10">
              <div className="relative grid grid-row-1 content-center text-green-700 w-2/3 h-36">
                {transferComplete ? (
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
            {transferComplete
              ? t('The following Phonons were transfered:')
              : t('The following Phonons are being transfered:')}
          </h3>
          <div
            className={
              'overflow-scroll gap-2 grid w-full' +
              (transferComplete ? '' : ' animate-pulse opacity-60')
            }
          >
            {destinationCard.OutgoingTransferProposal?.map((phonon, key) => (
              <Phonon
                key={key}
                phonon={phonon}
                card={destinationCard}
                isProposed={true}
                showAction={false}
              />
            ))}
          </div>
        </ModalBody>

        <ModalFooter>
          {transferComplete ? (
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
