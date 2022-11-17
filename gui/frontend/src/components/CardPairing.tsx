import { Button } from '@chakra-ui/button';
import { Input, InputGroup, InputRightElement } from '@chakra-ui/input';
import { useClipboard } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import {
  cloudDownload,
  cloudUpload,
  cloudDone,
  arrowForward,
  repeatOutline,
} from 'ionicons/icons';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { notifySuccess } from '../utils/notify';
import { CardRemote } from './PhononCardStates/CardRemote';

export const CardPairing: React.FC<{ setShowPairingOptions }> = ({
  setShowPairingOptions = false,
}) => {
  const { t } = useTranslation();
  const pairingCode =
    '6UqNxx9DGCCWrXt+36HuLY6Bmkzf99Xz9bq02HVadg3hZ3mgGsyorvDKyBY6WkkkpFgszXu9E+Uol0gnD3TnPw==';
  const { onCopy, value, hasCopied } = useClipboard(pairingCode);
  const [currentStep, setCurrentStep] = useState('share');
  const [isPaired, setIsPaired] = useState(false);

  useEffect(() => {
    if (currentStep === 'pairing') {
      setTimeout(() => {
        setCurrentStep('success');

        notifySuccess(
          t('Successfully paired to remote card: ' + '04e0d5eb884a73c0' + '!')
        );
      }, 3000);
    }

    if (currentStep === 'success') {
      setTimeout(() => {
        setIsPaired(true);
      }, 1000);
    }
  }, [currentStep, t]);

  const unpair = () => {
    setShowPairingOptions(false);
    setIsPaired(false);
  };

  return (
    <>
      {currentStep === 'share' && (
        <div className="w-80 h-52 rounded-lg border border-4 overflow-hidden transition-all border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat">
          <Button
            size="xs"
            colorScheme="red"
            variant="ghost"
            className="absolute top-0 left-1"
            onClick={() => {
              setShowPairingOptions(false);
            }}
          >
            {t('Cancel')}
          </Button>
          <div className="flex flex-col gap-y-2 px-2 items-center justify-center text-xl">
            <IonIcon icon={cloudDownload} className="text-white" />
            <div>
              <span className="block text-center text-white text-base">
                {t(
                  "To pair, share this code with the person you'd like to pair with."
                )}
              </span>
            </div>
            <InputGroup size="md">
              <Input
                pr="4.5rem"
                type="text"
                bgColor="white"
                disabled={true}
                value={value}
              />
              <InputRightElement width="4.5rem">
                <Button h="1.75rem" size="sm" onClick={onCopy}>
                  {hasCopied ? t('Copied!') : t('Copy')}
                </Button>
              </InputRightElement>
            </InputGroup>
            <Button
              rightIcon={<IonIcon icon={arrowForward} />}
              size="sm"
              className="uppercase"
              onClick={() => {
                setCurrentStep('request');
              }}
            >
              {t('Next')}
            </Button>
          </div>
        </div>
      )}

      {currentStep === 'request' && (
        <div className="w-80 h-52 rounded-lg border border-4 overflow-hidden transition-all border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat">
          <Button
            size="xs"
            colorScheme="red"
            variant="ghost"
            className="absolute top-0 left-1"
            onClick={() => {
              setShowPairingOptions(false);
            }}
          >
            {t('Cancel')}
          </Button>
          <div className="flex flex-col gap-y-2 px-2 items-center justify-center text-xl">
            <IonIcon icon={cloudDownload} className="text-white" />
            <div>
              <span className="block text-center text-white text-base">
                {t(
                  "Input the person's pairing code below to initiate pairing."
                )}
              </span>
            </div>
            <InputGroup size="md">
              <Input type="text" bgColor="white" />
            </InputGroup>
            <Button
              leftIcon={<IonIcon icon={repeatOutline} />}
              size="sm"
              className="uppercase"
              onClick={() => {
                setCurrentStep('pairing');
              }}
            >
              {t('Initiate Pairing')}
            </Button>
          </div>
        </div>
      )}

      {currentStep === 'pairing' && (
        <div className="w-80 h-52 rounded-lg border border-4 overflow-hidden transition-all border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat">
          <Button
            size="xs"
            colorScheme="red"
            variant="ghost"
            className="absolute top-0 left-1"
            onClick={() => {
              setShowPairingOptions(false);
            }}
          >
            {t('Cancel')}
          </Button>
          <div className="flex flex-col gap-y-2 py-4 px-2 items-center justify-center text-xl">
            <IonIcon
              icon={cloudUpload}
              className="text-white animate-pulse text-5xl"
            />
            <div>
              <span className="block text-center text-white text-base">
                {t('Awaiting other person to establish pairing...')}
              </span>
            </div>
          </div>
        </div>
      )}

      {currentStep === 'success' && (
        <div
          className={
            'w-80 h-52 opacity-100 absolute transition-all flip-card duration-150 bg-transparent ' +
            (isPaired ? '' : 'flip-card-locked')
          }
        >
          <div className="flip-card-inner relative w-full h-full">
            <div className="flip-card-front w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card-blue bg-cover bg-no-repeat overflow-hidden">
              <CardRemote unpair={unpair} />
            </div>
            <div className="flip-card-back w-full h-full absolute rounded-lg border border-4 overflow-hidden transition-all border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat">
              <div className="flex flex-col gap-y-2 py-12 px-2 items-center justify-center text-xl">
                <IonIcon
                  icon={cloudDone}
                  className="text-white animate-success text-6xl"
                />
                <div>
                  <span className="block text-center text-white text-base">
                    {t('Successfully paired!')}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
