import toast from 'react-hot-toast';
import { capitalize } from './formatting';

/**
 * Generates a success notification
 * @param text
 */
export const notifySuccess = (text: string) => {
  toast.success(text);
};

/**
 * Generate an error notification
 * @param text
 */
export const notifyError = (text: string) => {
  toast.error(text);
};

/**
 * Generates a promise notification with a loading state
 * @param promise
 * @param purpose
 */
export const notifyPromise = (promise, purpose: string) => {
  void toast.promise(
    promise,
    {
      loading: 'Waiting for ' + purpose,
      success: capitalize(purpose) + ' was successful!',
      error: 'Error with ' + purpose,
    },
    {
      duration: 8000,
    }
  );
};
