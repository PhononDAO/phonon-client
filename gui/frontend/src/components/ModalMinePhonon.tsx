import {
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  FormHelperText,
  Slider,
  SliderMark,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  Box,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { calendarOutline, menu } from 'ionicons/icons';
import { useContext, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import {
  GlobalSettings,
  PhononCard,
  Phonon as iPhonon,
} from '../interfaces/interfaces';
import { abbreviateHash } from '../utils/formatting';
import { notifySuccess } from '../utils/notify';
import { maxMiningDifficulty } from '../constants/Constants';
import localStorage from '../utils/localStorage';
import { CardManagementContext } from '../contexts/CardManagementContext';

type MiningFormData = {
  difficulty: number;
};

export const ModalMinePhonon: React.FC<{
  card: PhononCard;
  isOpen;
  onClose;
}> = ({ card, isOpen, onClose }) => {
  const { t } = useTranslation();
  const { addPhononsToCardState } = useContext(CardManagementContext);
  const defaultSettings: GlobalSettings =
    localStorage.getConfigurableSettings();
  const [defaultMiningDifficulty, setDefaultMiningDifficulty] = useState(
    defaultSettings.defaultMiningDifficulty
  );
  const [currentState, setCurrentState] = useState('settings');

  const labelStyles = {
    mt: '4',
    ml: '-2.5',
    fontSize: 'sm',
  };
  const { register, setValue, handleSubmit } = useForm<MiningFormData>();

  const setDifficultyValue = (value: number) => {
    setValue('difficulty', value);
    setDefaultMiningDifficulty(value);
  };

  // event when you start mining a phonon
  const onSubmit = (data: MiningFormData, event) => {
    event.preventDefault();

    console.log(data);

    setCurrentState('mining');

    notifySuccess(
      t(
        'Starting mining with a difficulty of {{difficulty}} on the card {{cardId}}',
        {
          difficulty: data.difficulty,
          cardId: card.CardId,
        }
      )
    );

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const promise = new Promise((resolve) => {
      setTimeout(() => {
        resolve('mined');
      }, 5000);
    }).then(() => {
      setCurrentState('result');

      const nPhonon = {
        PubKey:
          '04fc53a5e843e76cac55e7ce43d7592fb5002a749832b1f65708e84108e958fe6cfdd459f144ccb7739f947c1f317e9cfaa1c40bd138358e155afffdd626de003ab',
        Address: '',
        AddressType: 0,
        SchemaVersion: 0,
        ExtendedSchemaVersion: 0,
        CurveType: 0,
        ChainID: 0,
        Denomination: '88320000000000000',
        CurrencyType: 0,
        SourceCardId: card.CardId,
        ValidationStatus: 'unvalidated',
      } as iPhonon;

      addPhononsToCardState(card, [nPhonon]);

      notifySuccess(
        t('Phonon "{{phononPubKey}}" was mined on the card {{cardId}}!', {
          phononPubKey: abbreviateHash(nPhonon.PubKey),
          cardId: card.CardId,
        })
      );
    });
  };

  const cancelMining = () => {
    setCurrentState('settings');

    if (['settings', 'result'].includes(currentState)) {
      setDifficultyValue(defaultSettings.defaultMiningDifficulty);
      onClose();
    }
  };

  return (
    <Modal
      size="xl"
      isOpen={isOpen}
      onClose={cancelMining}
      closeOnOverlayClick={currentState !== 'mining'}
    >
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px)" />
      <ModalContent className="bg-black overflow-hidden">
        <ModalHeader className="text-white bg-black">
          <div className="font-noto-sans-mono">
            <div className="text-sm">{t('Mining a Phonon on')}</div>
            <div className="text-2xl">
              {card.VanityName ? card.VanityName : card.CardId}
            </div>
            {card.VanityName && (
              <div className="text-sm text-gray-400">{card.CardId}</div>
            )}
          </div>
        </ModalHeader>
        {currentState !== 'mining' && (
          <ModalCloseButton onClick={cancelMining} />
        )}
        <form
          // eslint-disable-next-line @typescript-eslint/no-misused-promises
          onSubmit={handleSubmit(onSubmit)}
        >
          <ModalBody pb={6} className="text-white bg-black">
            {currentState === 'settings' && (
              <FormControl>
                <FormLabel>{t('Mining Difficulty')}</FormLabel>
                <Box pt={12} pb={4}>
                  <Slider
                    {...register('difficulty')}
                    aria-label={t('mining difficulty')}
                    defaultValue={defaultSettings.defaultMiningDifficulty}
                    min={1}
                    max={maxMiningDifficulty}
                    value={defaultMiningDifficulty}
                    onChange={(val) => {
                      setDifficultyValue(val);
                    }}
                  >
                    <SliderMark
                      value={Math.ceil(maxMiningDifficulty * 0.25)}
                      {...labelStyles}
                    >
                      {Math.ceil(maxMiningDifficulty * 0.25)}
                    </SliderMark>
                    <SliderMark
                      value={Math.ceil(maxMiningDifficulty * 0.5)}
                      {...labelStyles}
                    >
                      {Math.ceil(maxMiningDifficulty * 0.5)}
                    </SliderMark>
                    <SliderMark
                      value={Math.ceil(maxMiningDifficulty * 0.75)}
                      {...labelStyles}
                    >
                      {Math.ceil(maxMiningDifficulty * 0.75)}
                    </SliderMark>
                    <SliderMark
                      value={defaultMiningDifficulty}
                      textAlign="center"
                      bg="blue.500"
                      color="white"
                      mt="-14"
                      ml="-5"
                      w="10"
                      fontSize="2xl"
                      className="rounded"
                    >
                      {defaultMiningDifficulty}
                    </SliderMark>
                    <SliderTrack>
                      <SliderFilledTrack />
                    </SliderTrack>
                    <SliderThumb
                      boxSize={6}
                      textColor="blue.500"
                      _hover={{ bg: 'blue.500', textColor: 'white' }}
                      _active={{ bg: 'blue.500', textColor: 'white' }}
                    >
                      <IonIcon icon={menu} />
                    </SliderThumb>
                  </Slider>
                </Box>
                <FormHelperText>
                  {t(
                    'The higher the difficulty, the longer it will take to mine a phonon.'
                  )}
                </FormHelperText>
              </FormControl>
            )}
            {['mining', 'result'].includes(currentState) && (
              <>
                {currentState === 'mining' && (
                  <img
                    className="w-32 h-32 mx-auto mb-8"
                    src="/assets/images/mining-phonon.gif"
                  />
                )}
                {currentState === 'result' && (
                  <>
                    <div className="w-24 h-24 mx-auto my-4 relative">
                      <img
                        className="w-full h-full z-50 absolute"
                        src="/assets/images/phonon-logo.png"
                      />
                      <img
                        className="w-full h-full animate-ping absolute"
                        src="/assets/images/phonon-logo.png"
                      />
                    </div>

                    <h3 className="text-xl text-white font-bold text-center">
                      {t('New Phonon Mined!')}
                    </h3>
                    <h4 className="text-sm text-gray-300 mb-8 text-center">
                      Hash:{' '}
                      {abbreviateHash(
                        '04fc53a5e843e76cac55e7ce43d7592fb5002a749832b1f65708e84108e958fe6cfdd459f144ccb7739f947c1f317e9cfaa1c40bd138358e155afffdd626de003ab'
                      )}
                    </h4>
                  </>
                )}
                <div>
                  <h3 className="text-lg font-medium text-white">
                    {t('Mining stats')}
                  </h3>
                  <dl className="mt-2 md:flex justify-between rounded-lg bg-gray-800 overflow-hidden divide-y divide-gray-700 md:divide-y-0 md:divide-x">
                    <div className="px-4 py-3">
                      <dt className="text-base font-normal text-gray-400">
                        {t('Attempts')}
                      </dt>
                      <dd className="mt-1 items-baseline">
                        <div className="flex items-baseline text-2xl font-semibold text-white">
                          30
                        </div>
                        <div className="text-xs text-gray-400"></div>
                      </dd>
                    </div>
                    {currentState === 'mining' && (
                      <div className="px-4 py-3">
                        <dt className="text-base font-normal text-gray-400">
                          {t('Time Elapsed')}
                        </dt>
                        <dd className="mt-1 items-baseline">
                          <div className="flex items-baseline text-2xl font-semibold text-white">
                            4.86
                          </div>
                          <div className="text-xs text-gray-400">
                            {t('seconds')}
                          </div>
                        </dd>
                      </div>
                    )}
                    {currentState === 'result' && (
                      <>
                        <div className="px-4 py-3">
                          <dt className="text-base font-normal text-gray-400">
                            {t('Time Elapsed')}
                          </dt>
                          <dd className="mt-1 items-baseline">
                            <div className="flex items-baseline text-2xl font-semibold text-white">
                              13.978
                            </div>
                            <div className="text-xs text-gray-400">
                              {t('seconds')}
                            </div>
                          </dd>
                        </div>
                        <div className="px-4 py-3">
                          <dt className="text-base font-normal text-gray-400">
                            {t('Avg. Time')}
                          </dt>
                          <dd className="mt-1 items-baseline">
                            <div className="flex items-baseline text-2xl font-semibold text-white">
                              0.465
                            </div>
                            <div className="text-xs text-gray-400">
                              {t('seconds')}
                            </div>
                          </dd>
                        </div>
                      </>
                    )}
                  </dl>

                  <dl className="mt-2 md:grid md:grid-cols-2 rounded-lg bg-gray-800 overflow-hidden divide-y divide-gray-700 md:divide-y-0 md:divide-x">
                    <div className="px-4 py-3">
                      <dt className="text-base font-normal text-gray-400">
                        {t('Started mining on:')}
                      </dt>
                      <dd className="mt-1 flex text-white">
                        <IonIcon slot="end" icon={calendarOutline} />
                        <div className="text-xs inline ml-2">
                          Dec 7, 2022 1:13 PM
                        </div>
                      </dd>
                    </div>

                    <div className="px-4 py-3">
                      <dt className="text-base font-normal text-gray-400">
                        {t('Completed mining on:')}
                      </dt>
                      <dd className="mt-1 flex text-white">
                        {currentState === 'mining' && (
                          <div className="text-xs inline">
                            {t('still mining...')}
                          </div>
                        )}
                        {currentState === 'result' && (
                          <>
                            <IonIcon slot="end" icon={calendarOutline} />
                            <div className="text-xs inline ml-2">
                              Dec 7, 2022 1:15 PM
                            </div>
                          </>
                        )}
                      </dd>
                    </div>
                  </dl>
                </div>
              </>
            )}
          </ModalBody>

          <ModalFooter className="bg-black">
            {currentState === 'settings' && (
              <Button colorScheme="green" type="submit" mr={3}>
                {t('Start Mining')}
              </Button>
            )}
            <Button onClick={cancelMining}>
              {currentState === 'result' ? t('Close') : t('Cancel')}
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
