import { useTranslation } from 'react-i18next';
import { JsonViewer } from '@textea/json-viewer';
import { useEthers } from '@usedapp/core';
import { useWindowDimensions } from 'window-dimensions-hooks';
import { Button, useClipboard } from '@chakra-ui/react';
import { useEffect } from 'react';

export const GlobalSettingsDiagnostics: React.FC = () => {
  const { t } = useTranslation();
  return (
    <>
      <div className="mb-4">
        {t(
          'The following device details and console history can be shared with the development team to help diagnose issues with Phonon Manager.'
        )}
      </div>

      <DeviceDetails />
      <ConsoleHistory />
    </>
  );
};

const DeviceDetails: React.FC = () => {
  const { t } = useTranslation();
  const { account, library } = useEthers();
  const { height, width } = useWindowDimensions();
  const { onCopy, value, setValue, hasCopied } = useClipboard(
    JSON.stringify({})
  );

  useEffect(() => {
    setValue(
      JSON.stringify({
        browser: navigator.userAgent,
        metamask: {
          installed: typeof window.ethereum !== 'undefined',
          walletConnected: !!account,
        },
        browserDimensions: {
          width: width,
          height: height,
        },
        network: library?.network
          ? {
              name: library.network.name,
              chainId: library.network.chainId,
              ensAddress: library.network.ensAddress,
            }
          : false,
      })
    );
  }, [account, height, library, setValue, width]);

  return (
    <>
      <h2 className="text-2xl mb-2 flex items-center gap-x-2">
        {t('Device Details')}{' '}
        <Button size="sm" colorScheme="blue" onClick={onCopy}>
          {hasCopied ? t('Copied!') : t('Copy')}
        </Button>
      </h2>
      <div className="h-80 w-full overflow-scroll rounded mb-6">
        <JsonViewer
          theme="dark"
          rootName={false}
          displayDataTypes={false}
          displayObjectSize={false}
          enableClipboard={false}
          className="w-full min-h-full rounded p-2 "
          value={JSON.parse(value)}
        />
      </div>
    </>
  );
};

const ConsoleHistory: React.FC = () => {
  const { t } = useTranslation();
  const { onCopy, value, setValue, hasCopied } = useClipboard(
    JSON.stringify({})
  );

  useEffect(() => {
    setValue(JSON.stringify(window.console['history']));
  }, [setValue]);

  return (
    <>
      <h2 className="text-2xl mb-2 flex items-center gap-x-2">
        {t('Console History')}
        <Button size="sm" colorScheme="blue" onClick={onCopy}>
          {hasCopied ? t('Copied!') : t('Copy')}
        </Button>
      </h2>
      <div className="h-80 w-full overflow-scroll rounded">
        <JsonViewer
          theme="dark"
          rootName={false}
          displayDataTypes={false}
          displayObjectSize={false}
          enableClipboard={false}
          className="w-full min-h-full rounded p-2"
          value={JSON.parse(value)}
        />
      </div>
    </>
  );
};
