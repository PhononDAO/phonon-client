/**
 * This is the object to be used for all Phonons.
 * Interacting with Phonons should be done through this class.
 */
export class Phonon {
  Address: string | null;
  AddressType: number | null;
  ChainID: number | null;
  CurveType: number | null;
  Denomination: string | null;
  ExtendedSchemaVersion: number | null;
  KeyIndex: number | null;
  PubKey: string | null;
  SchemaVersion: number | null;
  IsStaged: boolean | null;

  constructor() {
    this.Address = null;
    this.AddressType = null;
    this.ChainID = null;
    this.CurveType = null;
    this.Denomination = null;
    this.ExtendedSchemaVersion = null;
    this.KeyIndex = null;
    this.PubKey = null;
    this.SchemaVersion = null;
    this.IsStaged = null;
  }
}
