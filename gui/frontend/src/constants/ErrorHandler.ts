export const ErrorHandler = (
  error: Error,
  info: { componentStack: string }
) => {
  // Do something with the error
  console.error(error);
  // E.g. log to an error logging client here
  console.error(info);
};
