import differenceBy from 'lodash/differenceBy';
import unionBy from 'lodash/unionBy';
import { useState } from 'react';
import { Phonon } from '../interfaces/interfaces';

interface CardPhonons {
  [key: string]: Phonon[];
}

/**
 * `usePhonons` is a React hook that builds off of `useState` to add setter functions for
 * interacting with a list of objects:
 *  - `addPhonons` - Combines passed in array of records and records in state by comparing ids
 *  - `removePhonons` - Removes passed in array of records from records in state by comparing ids
 * @param defaultValue - any array to set the default value
 */
export const usePhonons = (
  defaultValue: Record<string, Phonon[]>,
  Address = 'Address'
): [
  CardPhonons,
  (cardId: string, toAdd: Phonon[]) => void,
  (cardId: string, toRemove: Phonon[]) => void,
  (cardId: string) => void
] => {
  const [records, setRecords] = useState<CardPhonons>(defaultValue);

  const addPhonons = (cardId: string, recordsToAdd: Phonon[]) =>
    setRecords((recordsInState: CardPhonons) => {
      recordsInState[cardId] = unionBy(
        cardId in recordsInState ? recordsInState[cardId] : [],
        recordsToAdd,
        Address
      );

      return recordsInState;
    });

  const removePhonons = (cardId: string, recordsToRemove: Phonon[]) =>
    setRecords((recordsInState: CardPhonons) => {
      recordsInState[cardId] = differenceBy(
        cardId in recordsInState ? recordsInState[cardId] : [],
        recordsToRemove,
        Address
      );

      return recordsInState;
    });

  const resetPhonons = (cardId: string) =>
    setRecords((recordsInState: CardPhonons) => {
      recordsInState[cardId] = [];

      return recordsInState;
    });

  return [records, addPhonons, removePhonons, resetPhonons];
};
