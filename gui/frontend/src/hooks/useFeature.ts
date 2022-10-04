import { useContext } from 'react';
import { version } from '../constants/version';

/**
 * `useFeature` is a React hook for feature flags that makes it easy to know when a particular
 * feature is active for a version of the Lattice firmware (or other external data).
 *
 * To add a feature, add a SNAKE_CASE key to the `features` variable with an array that specifies
 * the required version of app as [fix, minor, major].
 */
export const useFeature = (): { [feature: string]: boolean } => {
  const [major, minor, fix] = version.split('.').map(Number);

  const features = {
    CAN_MINE_PHONONS: [0, 1, 0],
  };

  return Object.fromEntries(
    Object.entries(features).map(([key, [_fix, _minor, _major]]) => [
      key,
      fix >= _fix && minor >= _minor && major >= _major,
    ])
  );
};
