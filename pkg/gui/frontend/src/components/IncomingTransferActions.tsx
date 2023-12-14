import { Button } from '@chakra-ui/react';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';

export const IncomingTransferActions: React.FC<{
  destinationCard: PhononCard;
  closeTransfer;
}> = ({ destinationCard, closeTransfer }) => {
  const { t } = useTranslation();
  const { addPhononsToCardTransferState, updateCardTransferStatusState } =
    useContext(CardManagementContext);

  const startValidation = () => {
    updateCardTransferStatusState(
      destinationCard,
      'IncomingTransferProposal',
      'validating'
    );

    // loop through all phonons and mark as validating
    destinationCard.IncomingTransferProposal?.Phonons?.map((phonon) => {
      phonon.ValidationStatus = 'validating';

      addPhononsToCardTransferState(
        destinationCard,
        [phonon],
        'IncomingTransferProposal'
      );

      // simulate validation
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const promise = new Promise((resolve) => {
        setTimeout(() => {
          resolve('validating');
        }, 3000);
      }).then(() => {
        updateCardTransferStatusState(
          destinationCard,
          'IncomingTransferProposal',
          'validated'
        );

        phonon.ValidationStatus = 'valid';

        addPhononsToCardTransferState(
          destinationCard,
          [phonon],
          'IncomingTransferProposal'
        );
      });
    });
  };

  const startTransfer = () => {
    updateCardTransferStatusState(
      destinationCard,
      'IncomingTransferProposal',
      'transferring'
    );

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const promise = new Promise((resolve) => {
      setTimeout(() => {
        resolve('paired');
      }, 8000);
    }).then(() => {
      updateCardTransferStatusState(
        destinationCard,
        'IncomingTransferProposal',
        'transferred'
      );
    });
  };

  return (
    <>
      {destinationCard.IncomingTransferProposal.Status === 'unvalidated' && (
        <Button className="mr-3" colorScheme="green" onClick={startValidation}>
          {t('Validate Assets')}
        </Button>
      )}
      {destinationCard.IncomingTransferProposal.Status === 'validated' && (
        <Button className="mr-3" colorScheme="green" onClick={startTransfer}>
          {t('Accept Transfer')}
        </Button>
      )}
      {!['transferring', 'transferred'].includes(
        destinationCard.IncomingTransferProposal.Status
      ) && (
        <Button className="mr-3" colorScheme="red" onClick={closeTransfer}>
          {t('Decline Transfer')}
        </Button>
      )}
      <Button onClick={closeTransfer}>
        {t(
          destinationCard.IncomingTransferProposal.Status === 'transferred'
            ? 'Close'
            : 'Cancel'
        )}
      </Button>
    </>
  );
};
