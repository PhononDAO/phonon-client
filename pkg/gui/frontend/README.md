# Getting Started with Phonon Manager Development

This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app). This project was set up with a few standard tools for development. See below.

## Available Scripts

In the project directory, you can run:

### `npm run start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in your browser.

The page will reload when you make changes.\
You may also see any lint errors in the console.

### `npm run start:tailwind`

Starts a watcher that keeps Tailwind CSS styles up to date.

### `npm run version`

Generates the latest version as an exportable variable.

### `npm run test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `npm run lint`

This runs lint checks with ESLint with Prettier formmatting.

### `npm run i18n-extract`

Extracts all strings within the application to be translated. Locale languages are set in the **i18next-parser.config.mjs**

# Tools for Development

## Error Handling

Most errors are handled by an `ErrorBoundary` wrapper via [react-error-boundary](https://github.com/bvaughn/react-error-boundary#readme). The wrapper has a fallback component which is structured in `ErrorFallback`. The wrapper handles errors in the constant `ErrorHandler`.

### Logical Errors

All logical code should be wrapped in a `try/catch` which throws an error like so:

```js
throw new Error({
  code: null, // only API errors return an error code
  message: 'Something went wrong',
});
```

### API Errors

API Error Keys and Messages can be found here: [API_ERROR_CODES](./../API_ERROR_CODES.md).

### Form Errors

Form errors are handled inline using field validation via [react-hook-form](https://react-hook-form.com/). Learn more [here](https://react-hook-form.com/get-started#Handleerrors)

Example:

```jsx
import React, { useState, useEffect } from "react";
import { useForm } from "react-hook-form";

export type FormData = {
  name: string;
};

export default function ExampleFormComponent() {
    const [ errorMessage, setErrorMessage ] = useState<string|null>(null);

    const {
        register,
        handleSubmit,
        reset,
        trigger,
        formState: { errors },
    } = useForm<FormData>();

    useEffect(() => {
        if(errors?.name) {
            setErrorMessage('Name must be 3 characters');
        }
    }, [errors]);

    return (
        <>
            <div>{errorMessage}</div>
            <input
                {...register("name", {
                    required: true,
                    onChange: () => trigger(),
                    validate: async (value) => {
                        return value.length > 3;
                    },
                })}
            />
        </>
    );

}
```

## I18N Locale Language Support

I18N Language support is handled with [react-i18next](https://react.i18next.com/). Language string extractions are handled by [i18next-parser](https://github.com/i18next/i18next-parser). Language is detected by the browser handled by [i18next-browser-languagedetector](https://github.com/i18next/i18next-browser-languageDetector).

All text strings in components should be wrapped in `t('STRING TEXT HERE')`.

To generate text strings for translation run the command `npm run i18n-extract`.

To add new language locales update the `locales` in **i18next-parser.config.mjs**.

To change the language within the app use the `useTranslation` hook:

```js
const { t, i18n } = useTranslation();

const changeLanguage = async (language) => {
  return await i18n.changeLanguage(language);
};

useEffect(() => {
  changeLanguage('fr-FR').catch((err) => {
    console.log(err);
  });
}, []);
```

## Notifications

The application has a notification system built in. Notifications are handled by [react-hot-toast](https://react-hot-toast.com/).

To send a success notification:

```js
notifySuccess('This was a success!');
```

To send an error notification:

```js
notifyError('There was an error.');
```

To send a promise notification:

```js
const somePromise = new Promise((resolve) => setTimeout(resolve, 4000));

notifyPromise(somePromise, 'blockchain transaction');
```

## Testing

**TODO**

## Feature Version Support

To make a feature work for a specific version add a flag to the hook `useFeature`.

# Deployment

This section has moved here: [https://facebook.github.io/create-react-app/docs/deployment](https://facebook.github.io/create-react-app/docs/deployment)

### `npm run build` fails to minify

This section has moved here: [https://facebook.github.io/create-react-app/docs/troubleshooting#npm-run-build-fails-to-minify](https://facebook.github.io/create-react-app/docs/troubleshooting#npm-run-build-fails-to-minify)

# Learn More

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).
