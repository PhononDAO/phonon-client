import Tippy from '@tippyjs/react';
import { IonIcon } from '@ionic/react';
import { helpCircle } from 'ionicons/icons';

export const HelpTooltip: React.FC<{
  text: string;
  tooltip: string | JSX.Element;
  theme?: 'error' | 'normal';
}> = ({ text, tooltip = false, theme = 'normal' }) => {
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
    <Tippy theme="light" content={tooltip}>
      <button
        className={'inline flex space-x-1 cursor-pointer ' + colors[theme].text}
      >
        <IonIcon slot="start" icon={helpCircle} />
        <p className="text-sm">{text}</p>
      </button>
    </Tippy>
  );
};
