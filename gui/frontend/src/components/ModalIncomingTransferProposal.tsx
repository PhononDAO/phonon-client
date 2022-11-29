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
import { send, shieldCheckmark } from 'ionicons/icons';
import { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { notifySuccess } from '../utils/notify';
import { Card } from './Card';
import { PhononValidator } from './PhononValidator';

export const ModalIncomingTransferProposal: React.FC<{
  sourceCard: PhononCard;
  destinationCard: PhononCard;
  isOpen;
  onClose;
}> = ({ sourceCard, destinationCard, isOpen, onClose }) => {
  const { t } = useTranslation();
  const { resetPhononsOnCardTransferState } = useContext(CardManagementContext);
  const [transferState, setTransferState] = useState('waiting');

  const closeTransfer = () => {
    setTransferState('waiting');
    onClose();

    // let's clear the incoming transfer proposal
    resetPhononsOnCardTransferState(destinationCard);
  };

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const promise = new Promise((resolve) => {
    // setTimeout(() => {
    //   resolve('paired');
    // }, 8000);
  }).then(() => {
    setTransferState('completed');
  });

  useEffect(() => {
    if (transferState === 'completed') {
      notifySuccess(
        t(
          'Successfully transfered ' +
            String(sourceCard.IncomingTransferProposal.length) +
            ' phonons from ' +
            String(destinationCard.CardId) +
            ' → ' +
            sourceCard.CardId
        )
      );
    }
  }, [transferState]);

  return (
    <Modal
      isOpen={isOpen}
      size="4xl"
      onClose={closeTransfer}
      closeOnOverlayClick={['waiting', 'completed'].includes(transferState)}
    >
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <span className="text-5xl font-bandeins-sans-light">
            Incoming Phonons
          </span>
        </ModalHeader>
        {['waiting', 'completed'].includes(transferState) && (
          <ModalCloseButton />
        )}
        <ModalBody pb={6}>
          <div className="relative">
            <div className="absolute flex justify-center w-full z-10">
              <div className="relative grid grid-row-1 content-center text-green-700 w-2/3 h-36">
                {transferState === 'completed' && (
                  <>
                    <IonIcon
                      icon={send}
                      className="mx-auto -rotate-30 text-5xl"
                    />
                    <div className="mt-4 text-sm text-center">
                      {t('Phonons Transferred Successfully!')}
                    </div>
                  </>
                )}
                {transferState === 'transferring' && (
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
                    <div className="mt-4 text-sm text-center">
                      {t('Sending Phonons...')}
                    </div>
                  </>
                )}
                {transferState === 'waiting' && (
                  <>
                    <IonIcon
                      icon={send}
                      className="mx-auto -rotate-30 text-4xl text-black"
                    />
                    <div className="mt-4 text-sm text-center text-black">
                      {t('The remote card is attempting to transfer Phonons.')}
                    </div>
                  </>
                )}
                {transferState === 'validating' && (
                  <>
                    <IonIcon
                      icon={shieldCheckmark}
                      className="mx-auto text-4xl text-blue-600 animate-ping"
                    />
                    <div className="mt-4 text-sm text-center text-blue-600">
                      {t('The remote card is validating Phonons to transfer.')}
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
            {transferState === 'waiting' &&
              t('The following Phonons are waiting to be transfered:')}
            {transferState === 'validating' &&
              t('The following Phonons are being validated:')}
            {transferState === 'transferring' &&
              t('The following Phonons are being transfered:')}
            {transferState === 'completed' &&
              t('The following Phonons were transfered:')}
          </h3>
          <div className={'overflow-scroll gap-2 grid w-full'}>
            {sourceCard.IncomingTransferProposal?.map((phonon, key) => (
              <PhononValidator
                key={key}
                phonon={phonon}
                card={sourceCard}
                isProposed={true}
                showAction={false}
              />
            ))}
          </div>
        </ModalBody>

        <ModalFooter>
          {transferState === 'waiting' && (
            <Button
              className="mr-3"
              colorScheme="green"
              onClick={() => {
                setTransferState('validating');
              }}
            >
              {t('Validate Assets')}
            </Button>
          )}
          <Button className="mr-3" colorScheme="red" onClick={closeTransfer}>
            {t('Decline Transfer')}
          </Button>
          <Button onClick={closeTransfer}>{t('Cancel')}</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
