/**
 * Capitalizes the first letter of the string
 * @param text
 * @returns string
 */
export const capitalize = (text: string) => {
  return text[0].toUpperCase() + text.slice(1);
};
