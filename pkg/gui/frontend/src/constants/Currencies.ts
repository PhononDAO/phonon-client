import { Currency } from '../interfaces/interfaces';

export const CURRENCIES: { [key: string]: Currency } = {
  2: {
    name: 'Ether',
    ticker: 'ETH',
    decimals: 18,
  },
  3: {
    name: 'Matic',
    ticker: 'MATIC',
    decimals: 18,
  },
};
