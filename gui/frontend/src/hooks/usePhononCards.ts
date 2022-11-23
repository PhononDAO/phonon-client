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
  Address = 'Address'
): [
  T[],
  (toAdd: T[]) => void,
  (toRemove: T[]) => void,
  () => void,
  (cardId: string) => T,
  (card: T, toAdd: Phonon[]) => void,
  (card: T, toRemove: Phonon[]) => void,
  (card: T) => void,
  (card: T, toAdd: Phonon[]) => void,
  (card: T, toRemove: Phonon[]) => void,
  (card: T) => void
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

  const addPhononsToCard = (card: T, phononsToAdd: Phonon[]) => {
    card.Phonons = unionBy(card.Phonons, phononsToAdd, Address);

    addPhononCards([card]);
  };

  const removePhononsFromCard = (card: T, phononsToRemove: Phonon[]) => {
    card.Phonons = differenceBy(card.Phonons, phononsToRemove, Address);

    addPhononCards([card]);
  };

  const resetPhononsOnCard = (card: T) => {
    card.Phonons = [];

    addPhononCards([card]);
  };

  const addPhononsForTransferToCard = (
    destinationCard: T,
    phononsToAdd: Phonon[]
  ) => {
    // let's we update the transfer proposal for this card
    destinationCard.IncomingTransferProposal = unionBy(
      destinationCard.IncomingTransferProposal,
      phononsToAdd,
      Address
    );
    addPhononCards([destinationCard]);

    // now update the phonons on the source cards
    phononsToAdd.map((phonon) => {
      phonon.ProposedForTransfer = true;
    });
  };

  const removePhononsForTransferFromCard = (
    destinationCard: T,
    phononsToRemove: Phonon[]
  ) => {
    // let's we update the transfer proposal for this card
    destinationCard.IncomingTransferProposal = differenceBy(
      destinationCard.IncomingTransferProposal,
      phononsToRemove,
      Address
    );
    addPhononCards([destinationCard]);

    // now update the phonons on the source cards
    phononsToRemove.map((phonon) => {
      phonon.ProposedForTransfer = false;
    });
  };

  const resetPhononsForTransferOnCard = (card: T) => {
    removePhononsForTransferFromCard(card, card.IncomingTransferProposal);
  };

  return [
    records,
    addPhononCards,
    removePhononCards,
    resetPhononCards,
    getCardById,
    addPhononsToCard,
    removePhononsFromCard,
    resetPhononsOnCard,
    addPhononsForTransferToCard,
    removePhononsForTransferFromCard,
    resetPhononsForTransferOnCard,
  ];
};
