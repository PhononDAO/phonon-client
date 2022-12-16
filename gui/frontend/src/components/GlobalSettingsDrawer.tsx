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
import { useTranslation } from 'react-i18next';
import { GlobalSettingsActivityHistory } from './GlobalSettingsActivityHistory';
import { GlobalSettingsDiagnostics } from './GlobalSettingsDiagnostics';
import { GlobalSettingsSettingsForm } from './GlobalSettingsSettingsForm';

export const GlobalSettingsDrawer: React.FC<{
  isOpen;
  onClose;
}> = ({ isOpen, onClose }) => {
  const { t } = useTranslation();

  return (
    <Drawer onClose={onClose} isOpen={isOpen} size="xl">
      <DrawerOverlay bg="blackAlpha.300" backdropFilter="blur(10px)" />
      <DrawerContent>
        <DrawerCloseButton />
        <Tabs>
          <DrawerHeader>
            <TabList overflowX="scroll">
              <Tab>{t('Settings')}</Tab>
              <Tab>{t('Activity History')}</Tab>
              <Tab>{t('Diagnostics')}</Tab>
            </TabList>
          </DrawerHeader>
          <DrawerBody>
            <TabPanels>
              <TabPanel>
                <GlobalSettingsSettingsForm />
              </TabPanel>
              <TabPanel>
                <GlobalSettingsActivityHistory />
              </TabPanel>
              <TabPanel>
                <GlobalSettingsDiagnostics />
              </TabPanel>
            </TabPanels>
          </DrawerBody>
        </Tabs>
      </DrawerContent>
    </Drawer>
  );
};
