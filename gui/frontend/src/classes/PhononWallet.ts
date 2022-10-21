import { PhononCard } from './PhononCard';

/**
 * This is the object to be used for holding phonon cards.
 */
export class PhononWallet {
  PhononCards: Array<PhononCard>;

  constructor() {
    this.PhononCards = [];
  }
}
