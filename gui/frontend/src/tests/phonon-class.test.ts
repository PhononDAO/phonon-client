/**
 * The following tests the attributes and functions of the Phonon class
 */
import { Phonon } from './../classes/Phonon';

describe('Phonon Class', () => {
  test('Correct attribute types', () => {
    // an initialized Phonon
    const aPhonon = new Phonon();

    expect(aPhonon.Address).toBeNull();
    expect(aPhonon.AddressType).toBeNull();
    expect(aPhonon.ChainID).toBeNull();
    expect(aPhonon.CurveType).toBeNull();
    expect(aPhonon.CurrencyType).toBeNull();
    expect(aPhonon.Denomination).toBeNull();
    expect(aPhonon.ExtendedSchemaVersion).toBeNull();
    expect(aPhonon.KeyIndex).toBeNull();
    expect(aPhonon.PubKey).toBeNull();
    expect(aPhonon.SchemaVersion).toBeNull();
    expect(aPhonon.IsStaged).toBeNull();

    // a data filled Phonon
    const bPhonon = new Phonon();
    bPhonon.Address = 'some address';
    bPhonon.AddressType = 1;
    bPhonon.ChainID = '1';
    bPhonon.CurveType = 1;
    bPhonon.CurrencyType = 2;
    bPhonon.Denomination = 'some deonomination';
    bPhonon.ExtendedSchemaVersion = 1;
    bPhonon.KeyIndex = 0;
    bPhonon.PubKey = 'some pubkey';
    bPhonon.SchemaVersion = 1;
    bPhonon.IsStaged = true;

    expect(typeof bPhonon.Address).toBe('string');
    expect(typeof bPhonon.AddressType).toBe('number');
    expect(typeof bPhonon.ChainID).toBe('string');
    expect(typeof bPhonon.CurveType).toBe('number');
    expect(typeof bPhonon.CurrencyType).toBe('number');
    expect(typeof bPhonon.Denomination).toBe('string');
    expect(typeof bPhonon.ExtendedSchemaVersion).toBe('number');
    expect(typeof bPhonon.KeyIndex).toBe('number');
    expect(typeof bPhonon.PubKey).toBe('string');
    expect(typeof bPhonon.SchemaVersion).toBe('number');
    expect(typeof bPhonon.IsStaged).toBe('boolean');
  });
});
