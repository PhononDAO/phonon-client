import { Chain } from '../interfaces/interfaces';

export const CHAINS: { [key: string]: Chain } = {
  '0': {
    name: 'nPhonon',
    textColor: 'text-black',
    bgColor: 'bg-white',
  },
  '1': {
    name: 'Mainnet',
    textColor: 'text-indigo-300',
    bgColor: 'bg-indigo-800',
  },
  '3': {
    name: 'Ropsten',
    textColor: 'text-pink-300',
    bgColor: 'bg-pink-800',
  },
  '42': {
    name: 'Kovan',
    textColor: 'text-purple-300',
    bgColor: 'bg-purple-800',
  },
  '4': {
    name: 'Rinkeby',
    textColor: 'text-yellow-300',
    bgColor: 'bg-yellow-800',
  },
  '5': {
    name: 'Goerli',
    textColor: 'text-blue-400',
    bgColor: 'bg-blue-800',
  },
  '56': {
    name: 'Binance',
    textColor: 'text-yellow-300',
    bgColor: 'bg-yellow-800',
  },
  '97': {
    name: 'Binance-testnet',
    textColor: 'text-yellow-300',
    bgColor: 'bg-yellow-800',
  },
  '137': {
    name: 'polygon',
    textColor: 'text-purple-300',
    bgColor: 'bg-purple-800',
  },
  '4002': {
    name: 'fantom-testnet',
    textColor: 'text-blue-300',
    bgColor: 'bg-blue-800',
  },
  '43113': {
    name: 'avalanche-testnet',
    textColor: 'text-red-300',
    bgColor: 'bg-red-800',
  },
  '43114': {
    name: 'avalanche',
    textColor: 'text-red-300',
    bgColor: 'bg-red-800',
  },
  '80001': {
    name: 'polygon-testnet',
    textColor: 'text-purple-300',
    bgColor: 'bg-purple-800',
  },
};
