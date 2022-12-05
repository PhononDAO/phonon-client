import {
  Button,
  ButtonGroup,
  IconButton,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { caretDown } from 'ionicons/icons';
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
        // let's randomw determine the validation result
        phonon.ValidationStatus = Math.round(Math.random()) ? 'valid' : 'error';

        addPhononsToCardTransferState(
          destinationCard,
          [phonon],
          'IncomingTransferProposal'
        );

        // let's determine the card's status state
        if (destinationCard.IncomingTransferProposal.Status !== 'has_errors') {
          updateCardTransferStatusState(
            destinationCard,
            'IncomingTransferProposal',
            phonon.ValidationStatus === 'valid' ? 'validated' : 'has_errors'
          );
        }
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
      {['unvalidated', 'has_errors'].includes(
        destinationCard.IncomingTransferProposal.Status
      ) && (
        <Menu offset={[-248, 4]}>
          <ButtonGroup className="mr-3" isAttached>
            <Button colorScheme="green" onClick={startValidation}>
              {t('Validate Assets')}
            </Button>
            <MenuButton
              as={IconButton}
              colorScheme="green"
              aria-label="More options"
              icon={<IonIcon icon={caretDown} />}
              className="border-l border-green-500"
            />
          </ButtonGroup>
          <MenuList>
            <MenuItem _hover={{ bg: 'green.300' }} onClick={startTransfer}>
              {t('Ignore Validation and Accept Transfer')}
            </MenuItem>
          </MenuList>
        </Menu>
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
