import {
  Box,
  Button,
  ButtonGroup,
  Divider,
  FormControl,
  FormHelperText,
  FormLabel,
  IconButton,
  Select,
  Slider,
  SliderFilledTrack,
  SliderMark,
  SliderThumb,
  SliderTrack,
  Stack,
  Switch,
  usePrevious,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { apps, menu, reorderFour } from 'ionicons/icons';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useFeature } from '../hooks/useFeature';
import { useTranslation } from 'react-i18next';
import { maxMiningDifficulty } from '../constants/Constants';
import localStorage from '../utils/localStorage';
import { notifySuccess } from '../utils/notify';
import { GlobalSettings } from '../interfaces/interfaces';
import { DateTime } from 'luxon';

type GlobalSettingsFormData = GlobalSettings;

export const GlobalSettingsSettingsForm: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { CAN_MINE_PHONONS } = useFeature();
  const { register, handleSubmit, getValues, setValue } =
    useForm<GlobalSettingsFormData>({
      defaultValues: localStorage.getConfigurableSettings(),
    });
  const [defaultMiningDifficulty, setDefaultMiningDifficulty] = useState(
    getValues('defaultMiningDifficulty')
  );
  const [defaultPhononLayout, setDefaultPhononLayout] = useState(
    getValues('defaultPhononLayout')
  );
  const languageOptions = {
    'en-US': t('English'),
    'es-MX': t('Spanish'),
  };

  const prevLanguage = usePrevious(getValues('defaultLanguage'));

  const changeLanguage = async (language) => {
    return await i18n.changeLanguage(language);
  };

  const labelStyles = {
    mt: '4',
    ml: '-2.5',
    fontSize: 'sm',
  };

  const setDifficultyValue = (value: number) => {
    setValue('defaultMiningDifficulty', value);
    setDefaultMiningDifficulty(value);
  };

  const setPhononLayout = (value: string) => {
    setValue('defaultPhononLayout', value);
    setDefaultPhononLayout(value);
  };

  const onSubmit = (data: GlobalSettingsFormData, event) => {
    event.preventDefault();

    localStorage.setConfigurableSettings(data);

    // if the language changes, then we update the manager
    if (prevLanguage !== data.defaultLanguage) {
      // update the language
      void changeLanguage(data.defaultLanguage).then(() => {
        document.title = t('PHONON MANAGER');
      });

      const activityHistory = localStorage.getActivityHistory();

      activityHistory.push({
        type: 'seperator',
        datetime: DateTime.now(),
        message: 'Language changed from {{prevLanguage}} to {{curLanguage}}',
        data: {
          prevLanguage: prevLanguage,
          curLanguage: data.defaultLanguage,
        },
      });

      localStorage.setActivityHistory(activityHistory);
    }

    notifySuccess(t('Phonon Manager settings saved!'));
  };

  return (
    <>
      <div className="mb-4">
        {t(
          'The following global settings allows you to configure Phonon Manager to best suite you.'
        )}
      </div>

      <form
        // eslint-disable-next-line @typescript-eslint/no-misused-promises
        onSubmit={handleSubmit(onSubmit)}
      >
        <div className="grid grid-cols-1 gap-y-3 h-144 overflow-scroll">
          <Divider />
          <FormControl>
            <FormLabel>{t('Language')}</FormLabel>
            <div className="w-48">
              <Select
                className="border rounded flex"
                {...register('defaultLanguage')}
                defaultValue={getValues('defaultLanguage')}
              >
                {Object.entries(languageOptions).map(([key, value]) => {
                  return (
                    <option key={key} value={key}>
                      {value}
                    </option>
                  );
                })}
              </Select>
            </div>
            <FormHelperText>
              {t('This is the language used with Phonon Manager.')}
            </FormHelperText>
          </FormControl>
          <Divider />
          <FormControl>
            <FormLabel>{t('Default Phonon Sort')}</FormLabel>
            <div className="w-48">
              <Select
                className="border rounded flex"
                {...register('defaultPhononSortBy')}
                defaultValue={getValues('defaultPhononSortBy')}
              >
                <option value="Key">{t('Key')}</option>
                <option value="ChainId">{t('Network Chain')}</option>
                <option value="Denomination">{t('Denomination')}</option>
                <option value="CurrencyType">{t('Currency Type')}</option>
              </Select>
            </div>
            <FormHelperText>
              {t('This is the default sort order to show phonons on a card.')}
            </FormHelperText>
          </FormControl>
          <Divider />
          <FormControl>
            <FormLabel>{t('Default Phonon Layout')}</FormLabel>
            <div className="rounded flex">
              <ButtonGroup className="border rounded-md" isAttached>
                <IconButton
                  bgColor={defaultPhononLayout === 'list' ? 'black' : 'white'}
                  textColor={defaultPhononLayout === 'list' ? 'white' : 'black'}
                  aria-label={t('List View')}
                  icon={<IonIcon icon={reorderFour} />}
                  onClick={() => {
                    setPhononLayout('list');
                  }}
                />
                <IconButton
                  bgColor={defaultPhononLayout === 'grid' ? 'black' : 'white'}
                  textColor={defaultPhononLayout === 'grid' ? 'white' : 'black'}
                  aria-label={t('Grid View')}
                  icon={<IonIcon icon={apps} />}
                  onClick={() => {
                    setPhononLayout('grid');
                  }}
                />
              </ButtonGroup>
            </div>
            <FormHelperText>
              {t('This is the default layout to show phonons on a card.')}
            </FormHelperText>
          </FormControl>
          <Divider />
          <FormControl>
            <FormLabel>{t('Auto-Validate Incoming Phonons')}</FormLabel>
            <Stack direction="row">
              <Switch
                colorScheme="green"
                size="lg"
                defaultChecked={getValues('autoValidateIncomingPhononRequests')}
                {...register('autoValidateIncomingPhononRequests')}
              />
            </Stack>
            <FormHelperText>
              {t(
                'Should we auto-validate phonons in an incoming transfer request.'
              )}
            </FormHelperText>
          </FormControl>
          <Divider />
          {CAN_MINE_PHONONS && (
            <>
              <FormControl>
                <FormLabel>{t('Default Mining Difficulty')}</FormLabel>
                <div className="w-96">
                  <Box pt={12} pb={4}>
                    <Slider
                      {...register('defaultMiningDifficulty')}
                      aria-label={t('mining difficulty')}
                      defaultValue={getValues('defaultMiningDifficulty')}
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
                </div>
                <FormHelperText>
                  {t('Set the default mining difficulty.')}
                </FormHelperText>
              </FormControl>
              <Divider />
            </>
          )}
          <Button colorScheme="green" type="submit">
            {t('Save Settings')}
          </Button>
        </div>
      </form>
    </>
  );
};
