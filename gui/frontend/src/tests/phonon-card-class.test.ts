/**
 * The following tests the attributes and functions of the Phonon class
 */
import { Phonon } from '../classes/Phonon';
import { PhononCard } from '../classes/PhononCard';

describe('Phonon Card Class', () => {
  test('Correct attribute types', () => {
    // an initialized PhononCard
    const aPhononCard = new PhononCard();

    expect(aPhononCard.CardId).toBeNull();
    expect(aPhononCard.VanityName).toBeNull();
    expect(aPhononCard.IsLocked).toBe(true);
    expect(aPhononCard.IsMock).toBe(false);
    expect(aPhononCard.InTray).toBe(false);
    expect(aPhononCard.IsRemote).toBe(false);
    expect(aPhononCard.ShowActions).toBe(true);
    expect(aPhononCard.Phonons).toStrictEqual([]);

    // a data filled Phonon Card
    const bPhononCard = new PhononCard();
    bPhononCard.CardId = 'some card id';
    bPhononCard.VanityName = 'some vanity name';

    const aPhonon = new Phonon();
    bPhononCard.Phonons = [aPhonon];

    expect(typeof bPhononCard.CardId).toBe('string');
    expect(typeof bPhononCard.VanityName).toBe('string');
    expect(typeof bPhononCard.IsLocked).toBe('boolean');
    expect(typeof bPhononCard.IsMock).toBe('boolean');
    expect(typeof bPhononCard.InTray).toBe('boolean');
    expect(typeof bPhononCard.IsRemote).toBe('boolean');
    expect(typeof bPhononCard.ShowActions).toBe('boolean');
    expect(typeof bPhononCard.Phonons).toBe('object');
    expect(bPhononCard.Phonons[0] instanceof Phonon).toBe(true);
  });
});
