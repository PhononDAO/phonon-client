/**
 * `useFeature` is a React hook for feature flags that allow you to enable/disable features.
 *
 * To add a feature, add a SNAKE_CASE key to the `features` variable.
 */
export const useFeature = (): { [feature: string]: boolean } => {
  const features = {
    CAN_MINE_PHONONS: true,
    ENABLE_MOCK_CARDS: true,
  };

  return Object.fromEntries(Object.entries(features));
};
