import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { send, shieldCheckmark, warning } from 'ionicons/icons';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { Card } from './Card';
import { IncomingTransferActions } from './IncomingTransferActions';

import { PhononValidator } from './PhononValidator';

export const ModalIncomingTransferProposal: React.FC<{
  sourceCard: PhononCard;
  destinationCard: PhononCard;
  isOpen;
  onClose;
}> = ({ sourceCard, destinationCard, isOpen, onClose }) => {
  const { t } = useTranslation();
  const { resetPhononsOnCardTransferState } = useContext(CardManagementContext);

  const closeTransfer = () => {
    onClose();

    // let's clear the incoming transfer proposal
    resetPhononsOnCardTransferState(
      destinationCard,
      'IncomingTransferProposal'
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      size="4xl"
      onClose={closeTransfer}
      closeOnOverlayClick={['unvalidated', 'completed'].includes(
        destinationCard.IncomingTransferProposal.Status
      )}
    >
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent>
        <ModalHeader>
          <span className="text-5xl font-bandeins-sans-light">
            {t('Incoming Phonons')}
          </span>
        </ModalHeader>
        {['unvalidated', 'completed'].includes(
          destinationCard.IncomingTransferProposal.Status
        ) && <ModalCloseButton />}
        <ModalBody pb={6}>
          <div className="relative">
            <div className="absolute flex justify-center w-full z-10">
              <div className="relative grid grid-row-1 content-center text-green-700 w-2/3 \ h-36">
                {destinationCard.IncomingTransferProposal.Status ===
                  'transferred' && (
                  <>
                    <IonIcon
                      icon={send}
                      className="mx-auto -rotate-30 text-5xl"
                    />
                    <div className="mt-4 px-28 text-sm text-center">
                      {t('Phonons Transferred Successfully!')}
                    </div>
                  </>
                )}
                {destinationCard.IncomingTransferProposal.Status ===
                  'transferring' && (
                  <>
                    <div className="flex justify-content align-items">
                      <span className="animate-incoming">
                        <IonIcon icon={send} className="rotate-180" />
                      </span>
                      <span className="animate-incoming animation-delay-1">
                        <IonIcon icon={send} className="rotate-180" />
                      </span>
                      <span className="animate-incoming animation-delay-2">
                        <IonIcon icon={send} className="rotate-180" />
                      </span>
                      <span className="animate-incoming animation-delay-3">
                        <IonIcon icon={send} className="rotate-180" />
                      </span>
                    </div>
                    <div className="mt-4 px-28 text-sm text-center">
                      {t('Receiving Phonons...')}
                    </div>
                  </>
                )}
                {destinationCard.IncomingTransferProposal.Status ===
                  'unvalidated' && (
                  <>
                    <IonIcon
                      icon={send}
                      className="mx-auto -rotate-30 text-4xl text-black"
                    />
                    <div className="mt-4 px-28 text-sm text-center text-black">
                      {t('The remote card is attempting to transfer Phonons.')}
                    </div>
                  </>
                )}
                {destinationCard.IncomingTransferProposal.Status ===
                  'validating' && (
                  <>
                    <IonIcon
                      icon={shieldCheckmark}
                      className="mx-auto text-4xl text-blue-600 animate-ping"
                    />
                    <div className="mt-4 px-28 text-sm text-center text-blue-600">
                      {t('The remote card is validating Phonons to transfer.')}
                    </div>
                  </>
                )}
                {destinationCard.IncomingTransferProposal.Status ===
                  'validated' && (
                  <>
                    <IonIcon
                      icon={shieldCheckmark}
                      className="mx-auto text-5xl text-green-500"
                    />
                    <div className="mt-4 px-28 text-sm text-center text-green-600">
                      {t('The incoming Phonons have been validated.')}
                    </div>
                  </>
                )}
                {destinationCard.IncomingTransferProposal.Status ===
                  'has_errors' && (
                  <>
                    <IonIcon
                      icon={warning}
                      className="mx-auto text-5xl text-yellow-500"
                    />
                    <div className="mt-4 px-28 text-sm text-center text-yellow-600">
                      {t(
                        'Error validating phonons. Hover over the validation errors below.'
                      )}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
          <div className="relative flex justify-between z-30">
            <div className="relative w-56 h-36">
              <Card card={destinationCard} isMini={true} showActions={false} />
            </div>
            <div className="relative w-56 h-36">
              <Card card={sourceCard} isMini={true} showActions={false} />
            </div>
          </div>

          <h3 className="mt-8 mb-2 text-xl text-gray-500">
            {destinationCard.IncomingTransferProposal.Status ===
              'unvalidated' &&
              t('The following Phonons are waiting to be transferred:')}
            {destinationCard.IncomingTransferProposal.Status === 'validating' &&
              t('The following Phonons are being validated:')}
            {destinationCard.IncomingTransferProposal.Status === 'validated' &&
              t('The following Phonons have been validated:')}
            {destinationCard.IncomingTransferProposal.Status ===
              'transferring' &&
              t('The following Phonons are being transferred:')}
            {destinationCard.IncomingTransferProposal.Status ===
              'transferred' && t('The following Phonons were transferred:')}
          </h3>
          <div
            className={
              'overflow-scroll gap-2 grid w-full' +
              (destinationCard?.IncomingTransferProposal.Status ===
              'transferring'
                ? ' animate-pulse opacity-60'
                : '')
            }
          >
            {destinationCard?.IncomingTransferProposal?.Phonons?.map(
              (phonon, key) => (
                <PhononValidator
                  key={key}
                  phonon={phonon}
                  isProposed={true}
                  showAction={false}
                  isTransferred={
                    destinationCard.IncomingTransferProposal.Status ===
                    'transferred'
                  }
                />
              )
            )}
          </div>
        </ModalBody>

        <ModalFooter>
          <IncomingTransferActions
            destinationCard={destinationCard}
            sourceCard={sourceCard}
            closeTransfer={closeTransfer}
          />
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
