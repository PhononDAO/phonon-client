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
  InTray: boolean;
  AttemptUnlock: boolean;
  FutureAction: string | null;
  ShowActions: boolean;
  Phonons: Array<Phonon>;

  constructor() {
    this.CardId = null;
    this.VanityName = null;
    this.IsLocked = true;
    this.IsMock = false;
    this.InTray = false;
    this.AttemptUnlock = false;
    this.FutureAction = null;
    this.ShowActions = true;
    this.Phonons = [];
  }

  unlock() {
    if (this.IsLocked) {
      this.IsLocked = false;
    }

    return !this.IsLocked;
  }
}
