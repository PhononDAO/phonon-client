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
import { useContext, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../interfaces/interfaces';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { notifySuccess } from '../utils/notify';
import { CardRemote } from './PhononCardStates/CardRemote';

type RemotePairingFormData = {
  remotePairingCode: string;
};

export const CardPairing: React.FC<{ setShowPairingOptions }> = ({
  setShowPairingOptions = false,
}) => {
  const { t } = useTranslation();
  const {
    phononCards,
    addCardsToState,
    removeCardsFromState,
    getCardPairingCode,
    resetPhononsOnCardTransferState,
  } = useContext(CardManagementContext);
  const loadedCards = phononCards.filter((card: PhononCard) => card.InTray);
  const { onCopy, value, hasCopied } = useClipboard(
    getCardPairingCode(loadedCards[0].CardId)
  );
  const [currentStep, setCurrentStep] = useState('share');
  const [isPaired, setIsPaired] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RemotePairingFormData>();

  // event when you submit remote pairing form data
  const onSubmit = (data: RemotePairingFormData, event) => {
    event.preventDefault();

    console.log(JSON.parse(atob(data.remotePairingCode)));

    setCurrentStep('pairing');

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const promise = new Promise((resolve) => {
      setTimeout(() => {
        resolve('paired');
      }, 3000);
    })
      .then(() => {
        setCurrentStep('success');
        notifySuccess(
          t('Successfully paired to remote card: {{cardId}}!', {
            cardId: '04e0d5eb884a73c0',
          })
        );

        return new Promise((resolve) => {
          setTimeout(() => {
            resolve('success');
          }, 1000);
        });
      })
      .then(() => {
        setIsPaired(true);
        const remoteCard = {
          CardId: '04e0d5eb884a73e9',
          IsRemote: true,
          InTray: true,
        } as PhononCard;

        addCardsToState([remoteCard]);
      });
  };

  const unpair = () => {
    setShowPairingOptions(false);
    setIsPaired(false);
    setCurrentStep('share');

    resetPhononsOnCardTransferState(
      phononCards.filter((card: PhononCard) => card.InTray)[0],
      'IncomingTransferProposal'
    );

    removeCardsFromState(
      phononCards.filter((card: PhononCard) => card.IsRemote)
    );
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
          <div className="flex flex-col gap-y-2 px-1 items-center justify-center text-xl">
            <IonIcon icon={cloudDownload} className="text-white" />
            <div>
              <span className="block text-center text-white text-sm">
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
          <div className="flex flex-col gap-y-2 px-1 items-center justify-center text-xl">
            <IonIcon icon={cloudDownload} className="text-white" />
            <div>
              <span className="block text-center text-white text-sm">
                {t(
                  "Input the person's pairing code below to initiate pairing."
                )}
              </span>
            </div>
            <form
              className="w-full flex flex-col gap-y-2 px-2 items-center justify-center"
              // eslint-disable-next-line @typescript-eslint/no-misused-promises
              onSubmit={handleSubmit(onSubmit)}
            >
              <InputGroup size="md">
                <Input
                  type="text"
                  bgColor="white"
                  className="w-full"
                  {...register('remotePairingCode', {
                    required: 'Please provide a remote pairing code.',
                  })}
                />
              </InputGroup>
              {errors.remotePairingCode && (
                <span className="absolute -mt-5 text-red-600 text-xs">
                  {errors.remotePairingCode.message}
                </span>
              )}
              <Button
                leftIcon={<IonIcon icon={repeatOutline} />}
                size="sm"
                className="uppercase"
                type="submit"
              >
                {t('Initiate Pairing')}
              </Button>
            </form>
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
            <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card-blue bg-cover bg-no-repeat overflow-hidden">
              <CardRemote unpair={unpair} />
            </div>
            <div className="flip-card-front w-full h-full absolute rounded-lg border border-4 overflow-hidden transition-all border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat">
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
