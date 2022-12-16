import { useDragLayer, XYCoord } from 'react-dnd';
import { Phonon } from './Phonon';

export const PhononDragLayer: React.FC = () => {
  const { isDragging, phonon, currentOffset } = useDragLayer((monitor) => ({
    phonon: monitor.getItem(),
    currentOffset: monitor.getSourceClientOffset(),
    isDragging: monitor.isDragging(),
  }));

  const getDragLayerStyles = (currentOffset: XYCoord | null) => {
    const { x, y } = currentOffset;

    const transform = `translate(${x}px, ${y}px)`;
    return {
      transform,
      WebkitTransform: transform,
    };
  };

  if (!isDragging || !phonon.PubKey) {
    return null;
  }

  // show this while dragging
  return (
    <div className="fixed w-full h-full top-0 left-0 z-50 pointer-events-none">
      <div style={getDragLayerStyles(currentOffset)}>
        <Phonon phonon={phonon} isCustomDragLayer={true} />
      </div>
    </div>
  );
};
