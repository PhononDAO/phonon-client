import differenceBy from 'lodash/differenceBy';
import unionBy from 'lodash/unionBy';
import { useState } from 'react';
import { PhononCard, Phonon } from '../interfaces/interfaces';

/**
 * `usePhononCards` is a React hook that builds off of `useState` to add setter functions for
 * interacting with a list of objects:
 *  - `addPhononCards` - Combines passed in array of records and records in state by comparing ids
 *  - `removePhononCards` - Removes passed in array of records from records in state by comparing ids
 * @param defaultValue - any array to set the default value
 */
export const usePhononCards = <T extends PhononCard>(
  defaultValue: T[],
  CardId = 'CardId',
  PubKey = 'PubKey'
): [
  T[],
  (toAdd: T[]) => void,
  (toRemove: T[]) => void,
  () => void,
  (cardId: string) => T,
  (cardId: string) => string,
  (card: T, toAdd: Phonon[]) => void,
  (card: T, toRemove: Phonon[]) => void,
  (card: T) => void,
  (card: T, toAdd: Phonon[], purpose: string) => void,
  (card: T, toRemove: Phonon[], purpose: string) => void,
  (card: T, purpose: string) => void,
  (card: T, purpose: string, status: string) => void
] => {
  const [records, setRecords] = useState<T[]>(defaultValue);

  const addPhononCards = (recordsToAdd: T[]) =>
    setRecords((recordsInState) =>
      unionBy(recordsInState, recordsToAdd, CardId)
    );

  const removePhononCards = (recordsToRemove: T[]) =>
    setRecords((recordsInState) =>
      differenceBy(recordsInState, recordsToRemove, CardId)
    );

  const resetPhononCards = () => setRecords([]);

  const getCardById = (cardId: string) => {
    const foundRecords = records.filter((card) => card.CardId === cardId);
    return foundRecords.length > 0 ? foundRecords[0] : null;
  };

  const getCardPairingCode = (cardId: string) => {
    return btoa(
      JSON.stringify({
        IPAddress: '127.0.0.1', // TODO: NEED TO GET THIS
        Port: '8080', // TODO NEED TO GET THIS
        CardId: cardId,
      })
    );
  };

  const addPhononsToCard = (card: T, phononsToAdd: Phonon[]) => {
    card.Phonons = unionBy(card.Phonons, phononsToAdd, PubKey);

    addPhononCards([card]);
  };

  const removePhononsFromCard = (card: T, phononsToRemove: Phonon[]) => {
    card.Phonons = differenceBy(card.Phonons, phononsToRemove, PubKey);

    addPhononCards([card]);
  };

  const resetPhononsOnCard = (card: T) => {
    card.Phonons = [];

    addPhononCards([card]);
  };

  const addPhononsForTransferToCard = (
    destinationCard: T,
    phononsToAdd: Phonon[],
    proposalPurpose: string
  ) => {
    // if a proposal doesn't exist yet, let's create it
    if (!destinationCard[proposalPurpose]) {
      destinationCard[proposalPurpose] = {
        Status: 'unvalidated',
        Phonons: [],
      };
    }

    // let's update the transfer proposal for this card
    destinationCard[proposalPurpose].Phonons = unionBy(
      destinationCard[proposalPurpose].Phonons,
      phononsToAdd,
      PubKey
    );
    addPhononCards([destinationCard]);

    // now update the phonons on the source cards
    phononsToAdd.map((phonon) => {
      phonon.ProposedForTransfer = true;
    });
  };

  const removePhononsForTransferFromCard = (
    destinationCard: T,
    phononsToRemove: Phonon[],
    proposalPurpose: string
  ) => {
    // let's we update the transfer proposal for this card
    destinationCard[proposalPurpose].Phonons = differenceBy(
      destinationCard[proposalPurpose].Phonons,
      phononsToRemove,
      PubKey
    );
    addPhononCards([destinationCard]);

    // now update the phonons on the source cards
    phononsToRemove.map((phonon) => {
      phonon.ProposedForTransfer = false;
    });
  };

  const resetPhononsForTransferOnCard = (card: T, proposalPurpose: string) => {
    removePhononsForTransferFromCard(
      card,
      card[proposalPurpose].Phonons,
      proposalPurpose
    );
  };

  const updateCardTransferStatus = (
    card: T,
    proposalPurpose: string,
    status: string
  ) => {
    card[proposalPurpose].Status = status;

    addPhononCards([card]);
  };

  return [
    records,
    addPhononCards,
    removePhononCards,
    resetPhononCards,
    getCardById,
    getCardPairingCode,
    addPhononsToCard,
    removePhononsFromCard,
    resetPhononsOnCard,
    addPhononsForTransferToCard,
    removePhononsForTransferFromCard,
    resetPhononsForTransferOnCard,
    updateCardTransferStatus,
  ];
};
