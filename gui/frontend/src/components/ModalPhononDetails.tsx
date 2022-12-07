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
  Input,
  FormHelperText,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  TableContainer,
  Table,
  Tr,
  Tbody,
  Td,
} from '@chakra-ui/react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { CURRENCIES } from '../constants/Currencies';
import { Phonon } from '../interfaces/interfaces';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { notifySuccess } from '../utils/notify';
import { ChainIDTag } from './ChainIDTag';

type PhononFormData = {
  address: string;
};

export const ModalPhononDetails: React.FC<{
  phonon: Phonon;
  isOpen;
  onClose;
}> = ({ phonon, isOpen, onClose }) => {
  const { t } = useTranslation();
  const [tabIndex, setTabIndex] = useState(0);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PhononFormData>();

  // event when you redeem a phonon
  const onSubmit = (data: PhononFormData, event) => {
    event.preventDefault();

    onClose();
    notifySuccess(
      t(
        'Phonon ' +
          abbreviateHash(phonon.Address) +
          ' in the amount of ' +
          fromDecimals(
            phonon.Denomination,
            CURRENCIES[phonon.CurrencyType].decimals
          ) +
          CURRENCIES[phonon.CurrencyType].ticker +
          ' was redeemed!'
      )
    );
  };

  return (
    <Modal size="2xl" isOpen={isOpen} onClose={onClose}>
      <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px) " />
      <ModalContent className="overflow-hidden">
        <ModalHeader className="bg-black text-white text-center">
          <ChainIDTag id={phonon.ChainID} />
          <div className="text-3xl text-white font-bandeins-sans-bold">
            <>
              {fromDecimals(
                phonon.Denomination,
                CURRENCIES[phonon.CurrencyType].decimals
              )}
              <span className="text-base font-bandeins-sans-light ml-2">
                {CURRENCIES[phonon.CurrencyType].ticker}
              </span>
            </>
          </div>
          <div className="text-gray-400 ml-auto">
            {abbreviateHash(phonon.Address)}
          </div>
        </ModalHeader>
        <ModalCloseButton />
        <form
          // eslint-disable-next-line @typescript-eslint/no-misused-promises
          onSubmit={handleSubmit(onSubmit)}
        >
          <ModalBody pb={6}>
            <Tabs
              onChange={(index) => setTabIndex(index)}
              colorScheme="blackAlpha"
            >
              <TabList>
                <Tab>Details</Tab>
                <Tab>Redeem</Tab>
              </TabList>

              <TabPanels>
                <TabPanel padding={0} className="mt-4">
                  <TableContainer>
                    <Table size="sm">
                      <Tbody>
                        <Tr>
                          <Td>{t('Address')}</Td>
                          <Td isNumeric>{phonon.Address}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Address Type')}</Td>
                          <Td isNumeric>{phonon.AddressType}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Chain ID')}</Td>
                          <Td isNumeric>{phonon.ChainID}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Curve Type')}</Td>
                          <Td isNumeric>{phonon.CurveType}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Currency Type')}</Td>
                          <Td isNumeric>{phonon.CurrencyType}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Denomination')}</Td>
                          <Td isNumeric>{phonon.Denomination}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Public Key')}</Td>
                          <Td isNumeric>{phonon.PubKey}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Schema Version')}</Td>
                          <Td isNumeric>{phonon.SchemaVersion}</Td>
                        </Tr>
                        <Tr>
                          <Td>{t('Extended Schema Version')}</Td>
                          <Td isNumeric>{phonon.ExtendedSchemaVersion}</Td>
                        </Tr>
                      </Tbody>
                    </Table>
                  </TableContainer>
                </TabPanel>
                <TabPanel padding={0} className="grid grid-cols-1 gap-y-6 mt-4">
                  <FormControl>
                    <FormLabel>{t('Address to Redeem')}</FormLabel>
                    <Input
                      bg="gray.700"
                      color="white"
                      type="text"
                      placeholder="0x..."
                      {...register('address', {
                        required: 'Address to redeem is required.',
                      })}
                    />
                    {errors.address && (
                      <span className="text-red-600">
                        {errors.address.message}
                      </span>
                    )}
                    <FormHelperText>
                      {t(
                        'The redeemed Phonon will be sent to this address. Confirm the address belongs to the network above. Lost Phonons are lost forever.'
                      )}
                    </FormHelperText>
                  </FormControl>
                </TabPanel>
              </TabPanels>
            </Tabs>
          </ModalBody>

          <ModalFooter>
            {tabIndex === 1 && (
              <Button colorScheme="green" type="submit" mr={3}>
                {t('Redeem')}
              </Button>
            )}
            <Button onClick={onClose}>
              {tabIndex === 1 ? t('Cancel') : t('Close')}
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
