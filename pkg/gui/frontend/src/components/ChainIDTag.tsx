import { CHAINS } from '../constants/ChainID';

export const ChainIDTag: React.FC<{
  id: number;
  className?: string;
}> = ({ id, className = '' }) => {
  const chain = CHAINS[id];
  return (
    <div
      className={
        'inline-block px-4 py-px rounded-full font-bandeins-sans-bold uppercase whitespace-nowrap ' +
        chain.bgColor +
        ' ' +
        chain.textColor +
        ' ' +
        className
      }
    >
      {chain.name}
    </div>
  );
};
