import { Phonon } from './Phonon';

/**
 * This is the object to be used for all Phonon Cards.
 * Interacting with Phonon Cards should be done through this class.
 */
export class PhononCard {
  CardId: string | null;
  VanityName: string | null;
  IsLocked: boolean;
  IsMock: boolean;
  IsActive: boolean;
  Phonons: Array<Phonon>;

  constructor() {
    this.CardId = null;
    this.VanityName = null;
    this.IsLocked = true;
    this.IsMock = false;
    this.IsActive = false;
    this.Phonons = [];
  }
}
