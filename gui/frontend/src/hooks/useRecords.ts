import differenceBy from 'lodash/differenceBy';
import unionBy from 'lodash/unionBy';
import { useState } from 'react';
import { PhononCard } from '../interfaces/interfaces';

/**
 * `useRecords` is a React hook that builds off of `useState` to add setter functions for
 * interacting with a list of objects:
 *  - `addRecords` - Combines passed in array of records and records in state by comparing ids
 *  - `removeRecords` - Removes passed in array of records from records in state by comparing ids
 * @param defaultValue - any array to set the default value
 */
export const useRecords = <T extends PhononCard>(
  defaultValue: T[],
  CardId = 'CardId'
): [T[], (toAdd: T[]) => void, (toRemove: T[]) => void, () => void] => {
  const [records, setRecords] = useState<T[]>(defaultValue);

  const addRecords = (recordsToAdd: T[]) =>
    setRecords((recordsInState) =>
      unionBy(recordsInState, recordsToAdd, CardId)
    );

  const removeRecords = (recordsToRemove: T[]) =>
    setRecords((recordsInState) =>
      differenceBy(recordsInState, recordsToRemove, CardId)
    );

  const resetRecords = () => setRecords([]);

  return [records, addRecords, removeRecords, resetRecords];
};
