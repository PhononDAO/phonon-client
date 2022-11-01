import { Tooltip } from '@chakra-ui/react';
import { IonIcon } from '@ionic/react';
import { helpCircle } from 'ionicons/icons';

export const HelpTooltip: React.FC<{
  text: string;
  theme?: 'error' | 'normal';
  children;
}> = ({ text, theme = 'normal', children }) => {
  const colors = {
    normal: {
      text: 'text-white',
      tooltip: 'bg-white',
    },
    error: {
      text: 'text-red-600',
      tooltip: 'bg-red-600',
    },
  };

  return (
    <Tooltip hasArrow label={children} bg="gray.300" color="black">
      <button
        className={'inline flex space-x-1 cursor-pointer ' + colors[theme].text}
      >
        <IonIcon slot="start" icon={helpCircle} />
        <p className="text-sm">{text}</p>
      </button>
    </Tooltip>
  );
};
