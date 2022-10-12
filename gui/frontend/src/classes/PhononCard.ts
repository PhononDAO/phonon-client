import { Phonon } from './Phonon';

/**
 * This is the object to be used for all Phonon Cards.
 * Interacting with Phonon Cards should be done through this class.
 */
export class PhononCard {
  CardId: string | null;
  VanityName: string | null;
  IsLocked: boolean | null;
  Phonons: Array<Phonon>;

  constructor() {
    this.CardId = null;
    this.VanityName = null;
    this.IsLocked = null;
    this.Phonons = [];
  }
}
