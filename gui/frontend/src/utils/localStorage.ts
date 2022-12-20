/* eslint-disable @typescript-eslint/restrict-template-expressions */
/* eslint-disable @typescript-eslint/no-unsafe-return */
import omit from 'lodash/omit';
import { defaultConfigurableSettings } from '../constants/Constants';

const ROOT_STORE = process.env.REACT_APP_ROOT_STORE || 'phonon';
const ACTIVITY_HISTORY_STORAGE_KEY = 'activity_history';
const CONFIGURABLE_SETTINGS_STORAGE_KEY = 'configurable_settings';
const PHONON_CARDS_STORAGE_KEY = 'phonon_cards';

// #region -- Generic Local Storage Functions

const getItem = (key) => {
  const value: string = window.localStorage.getItem(key);
  try {
    return JSON.parse(value);
  } catch (e) {
    return JSON.parse(JSON.stringify(value));
  }
};
const setItem = (key, value) =>
  window.localStorage.setItem(key, JSON.stringify(value));
const removeItem = (key) => window.localStorage.removeItem(key);

const getRootStore = () => getItem(ROOT_STORE) ?? {};
const setRootStore = (value) =>
  window.localStorage.setItem(
    ROOT_STORE,
    JSON.stringify({ ...getRootStore(), ...value })
  );
const removeRootStore = () => removeItem(ROOT_STORE);

const getRootStoreItem = (key) => getItem(ROOT_STORE)?.[key] ?? {};
const setRootStoreItem = (key, value) =>
  window.localStorage.setItem(
    ROOT_STORE,
    JSON.stringify({ ...getRootStore(), [`${key}`]: value })
  );
const removeRootStoreItem = (key) =>
  window.localStorage.setItem(
    ROOT_STORE,
    JSON.stringify(omit(getRootStore(), key))
  );

// #endregion

// #region -- Activity History Functions

const getActivityHistory = () => getItem(ACTIVITY_HISTORY_STORAGE_KEY) ?? [];
const setActivityHistory = (value) =>
  setItem(ACTIVITY_HISTORY_STORAGE_KEY, value);

// #endregion

// #region -- Activity History Functions

const getConfigurableSettings = () =>
  getItem(CONFIGURABLE_SETTINGS_STORAGE_KEY) ?? defaultConfigurableSettings;
const setConfigurableSettings = (value) =>
  setItem(CONFIGURABLE_SETTINGS_STORAGE_KEY, value);

// #endregion

// #region -- Activity History Functions

const getPhononCards = () => getItem(PHONON_CARDS_STORAGE_KEY) ?? [];
const setPhononCards = (value) => setItem(PHONON_CARDS_STORAGE_KEY, value);
const clearPhononCards = () => removeItem(PHONON_CARDS_STORAGE_KEY);

// #endregion

const exports = {
  getItem,
  setItem,
  removeItem,
  getRootStore,
  setRootStore,
  removeRootStore,
  getRootStoreItem,
  setRootStoreItem,
  removeRootStoreItem,
  getActivityHistory,
  setActivityHistory,
  getConfigurableSettings,
  setConfigurableSettings,
  getPhononCards,
  setPhononCards,
  clearPhononCards,
};

export default exports;
