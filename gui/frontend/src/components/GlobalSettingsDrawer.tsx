import {
  Drawer,
  DrawerOverlay,
  DrawerContent,
  DrawerCloseButton,
  DrawerHeader,
  DrawerBody,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
} from '@chakra-ui/react';
import { useContext } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { notifySuccess } from '../utils/notify';

type SettingsFormData = {
  vanityName: string;
  cardPin: string;
};

export const GlobalSettingsDrawer: React.FC<{
  isOpen;
  onClose;
}> = ({ isOpen, onClose }) => {
  const { t } = useTranslation();
  const {
    control,
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<SettingsFormData>();

  return (
    <Drawer onClose={onClose} isOpen={isOpen} size="xl">
      <DrawerOverlay bg="blackAlpha.300" backdropFilter="blur(10px)" />
      <DrawerContent>
        <DrawerCloseButton />
        <Tabs>
          <DrawerHeader>
            <TabList>
              <Tab>Settings</Tab>
              <Tab>Activity History</Tab>
              <Tab>Diagnostics</Tab>
            </TabList>
          </DrawerHeader>
          <DrawerBody>
            <TabPanels>
              <TabPanel>
                <ol>
                  <li>set default mining difficulty</li>
                  <li>auto-validate incoming phonon requests</li>
                  <li>default phonon sort by</li>
                  <li>default phonon layout</li>
                </ol>
              </TabPanel>
              <TabPanel>activity history here</TabPanel>
              <TabPanel>diagnostics here</TabPanel>
            </TabPanels>
          </DrawerBody>
        </Tabs>
      </DrawerContent>
    </Drawer>
  );
};
