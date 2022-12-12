import {
  Box,
  Button,
  ButtonGroup,
  Divider,
  FormControl,
  FormHelperText,
  FormLabel,
  IconButton,
  Input,
  Select,
  Slider,
  SliderFilledTrack,
  SliderMark,
  SliderThumb,
  SliderTrack,
  Stack,
  Switch,
} from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { apps, menu, reorderFour } from 'ionicons/icons';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';

type GlobalSettingsFormData = {
  defaultMiningDifficulty: number;
  autoValidateIncomingPhononRequests: boolean;
  defaultPhononSortBy: string;
  defaultPhononLayout: string;
};

export const GlobalSettingsSettingsForm: React.FC = () => {
  const { t } = useTranslation();
  const defaultDifficulty = 5;
  const maxDifficulty = 30;
  const [sliderValue, setSliderValue] = useState(defaultDifficulty);
  const [defaultPhononLayout, setDefaultPhononLayout] = useState('');

  const {
    control,
    register,
    handleSubmit,
    watch,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<GlobalSettingsFormData>();

  const labelStyles = {
    mt: '4',
    ml: '-2.5',
    fontSize: 'sm',
  };

  const setDifficultyValue = (value: number) => {
    setValue('defaultMiningDifficulty', value);
    setSliderValue(value);
  };

  return (
    <>
      <div className="mb-4">
        {t(
          'The following global settings allows you to configure Phonon Manager to best suite you.'
        )}
      </div>

      <form>
        <div className="grid grid-cols-1 gap-y-6">
          <Divider />
          <FormControl>
            <FormLabel>{t('Default Phonon Sort')}</FormLabel>
            <div className="w-48">
              <Select
                className="border rounded flex"
                placeholder="Select order"
                {...register('defaultPhononSortBy')}
              >
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
                    setValue('defaultPhononLayout', 'list');
                    setDefaultPhononLayout('list');
                  }}
                />
                <IconButton
                  bgColor={defaultPhononLayout === 'grid' ? 'black' : 'white'}
                  textColor={defaultPhononLayout === 'grid' ? 'white' : 'black'}
                  aria-label={t('Grid View')}
                  icon={<IonIcon icon={apps} />}
                  onClick={() => {
                    setValue('defaultPhononLayout', 'grid');
                    setDefaultPhononLayout('grid');
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
          <FormControl>
            <FormLabel>{t('Default Mining Difficulty')}</FormLabel>
            <div className="w-96">
              <Box pt={12} pb={4}>
                <Slider
                  {...register('defaultMiningDifficulty')}
                  aria-label="mining-difficulty"
                  defaultValue={defaultDifficulty}
                  min={1}
                  max={maxDifficulty}
                  value={sliderValue}
                  onChange={(val) => {
                    setDifficultyValue(val);
                  }}
                >
                  <SliderMark
                    value={Math.ceil(maxDifficulty * 0.25)}
                    {...labelStyles}
                  >
                    {Math.ceil(maxDifficulty * 0.25)}
                  </SliderMark>
                  <SliderMark
                    value={Math.ceil(maxDifficulty * 0.5)}
                    {...labelStyles}
                  >
                    {Math.ceil(maxDifficulty * 0.5)}
                  </SliderMark>
                  <SliderMark
                    value={Math.ceil(maxDifficulty * 0.75)}
                    {...labelStyles}
                  >
                    {Math.ceil(maxDifficulty * 0.75)}
                  </SliderMark>
                  <SliderMark
                    value={sliderValue}
                    textAlign="center"
                    bg="blue.500"
                    color="white"
                    mt="-14"
                    ml="-5"
                    w="10"
                    fontSize="2xl"
                    className="rounded"
                  >
                    {sliderValue}
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
          <Button colorScheme="green" type="submit">
            {t('Save Settings')}
          </Button>
        </div>
      </form>
    </>
  );
};
