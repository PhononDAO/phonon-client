import { useDragLayer, XYCoord } from 'react-dnd';
import { Card } from './Card';

export const CardDragLayer: React.FC = () => {
  const { isDragging, card, currentOffset } = useDragLayer((monitor) => ({
    card: monitor.getItem(),
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

  if (!isDragging || !card.CardId) {
    return null;
  }

  // show this while dragging
  return (
    <div className="fixed w-full h-full top-0 left-0 z-50 pointer-events-none">
      <div style={getDragLayerStyles(currentOffset)}>
        <Card card={card} isCustomDragLayer={true} />
      </div>
    </div>
  );
};
