import { PhononWallet } from './PhononWallet';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { CardManagementContextProvider } from '../contexts/CardManagementContext';
import { DeckManager } from './DeckManager';
import { CardDragLayer } from './CardDragLayer';
import { PhononDragLayer } from './PhononDragLayer';
import { useTranslation } from 'react-i18next';
import { useEffect } from 'react';

export const Stage = () => {
  const { i18n } = useTranslation();

  const changeLanguage = async (language) => {
    return await i18n.changeLanguage(language);
  };

  useEffect(() => {
    changeLanguage('en-US').catch((err) => {
      console.log(err);
    });
  }, []);

  return (
    <main className="bg-zinc-900 shadow-top shadow-gray-600 font-bandeins-sans px-6 py-4 flex-grow relative">
      <CardManagementContextProvider>
        <DndProvider backend={HTML5Backend}>
          <PhononWallet />
          <DeckManager />
          <CardDragLayer />
          <PhononDragLayer />
        </DndProvider>
      </CardManagementContextProvider>
    </main>
  );
};
