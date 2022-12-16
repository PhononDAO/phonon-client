/**
 * Error interface to be thrown in errors
 */
export interface ErrorObject {
  code: number;
  message: string;
}

/**
 * Chain interface
 */
export interface Chain {
  name: string;
  textColor: string;
  bgColor: string;
}

/**
 * Chain interface
 */
export interface Currency {
  name: string;
  ticker: string;
  decimals: number;
}

/**
 * Phonon interface
 */
export interface Phonon {
  Address: string;
  AddressType: number;
  ChainID: number;
  CurveType: number;
  CurrencyType: number;
  Denomination: string;
  ExtendedSchemaVersion: number;
  KeyIndex: number;
  PubKey: string;
  SchemaVersion: number;
  ProposedForTransfer: boolean;
  SourceCardId: string;
  ValidationStatus: 'unvalidated' | 'validating' | 'valid' | 'error';
}

/**
 * PhononCard interface
 */
export interface PhononCard {
  CardId: string;
  VanityName: string;
  IsLocked: boolean;
  IsMock: boolean;
  InTray: boolean;
  IsRemote: boolean;
  AttemptUnlock: boolean;
  FutureAction: string | null;
  ShowActions: boolean;
  Phonons: Array<Phonon>;
  IncomingTransferProposal: PhononTransferProposal;
  OutgoingTransferProposal: PhononTransferProposal;
}

/**
 * PhononWallet interface
 */
export interface PhononWallet {
  PhononCards: Array<PhononCard>;
}

/**
 * PhononTransferProposal interface
 */
export interface PhononTransferProposal {
  Status:
    | 'unvalidated'
    | 'validating'
    | 'validated'
    | 'has_errors'
    | 'transferring'
    | 'transferred';
  Phonons: Array<Phonon>;
}

/**
 * GlobalSettings interface
 */
export interface GlobalSettings {
  defaultLanguage: string;
  defaultMiningDifficulty: number;
  autoValidateIncomingPhononRequests: boolean;
  defaultPhononSortBy: string;
  defaultPhononLayout: string;
}
