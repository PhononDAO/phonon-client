import BigNumber from 'bignumber.js';

/**
 * Capitalizes the first letter of the string
 * @param text
 * @returns string
 */
export const capitalize = (text: string) => {
  return text[0].toUpperCase() + text.slice(1);
};

/**
 * `abbreviateHash` takes a string and returns the first 10 and last 8 characters with an ellipsis between.
 * @param {string} hash - the hash string to shorten.
 */
export const abbreviateHash = (hash: string) =>
  hash && hash.length > 24
    ? `${hash.slice(0, 10)}...${hash.slice(hash.length - 8, hash.length)}`
    : hash ?? '';

/**
 * `fromDecimals` takes a denomination and decimal count and returns the denomination from decimals.
 * @param {string} denomination - the denomination amount
 * @param {number} decimals - the decimals
 */
export const fromDecimals = (denomination: string, decimals: number) => {
  const amount = new BigNumber(denomination);

  return amount.div(10 ** decimals).toFixed(3);
};
