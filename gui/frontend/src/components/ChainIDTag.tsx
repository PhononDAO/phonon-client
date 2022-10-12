import { CHAINS } from '../constants/ChainID';

export const ChainIDTag: React.FC<{
  id: string;
}> = ({ id }) => {
  const chain = CHAINS[id];
  return (
    <div
      className={
        'inline-block px-6 py-2 rounded-full font-bandeins-sans-bold uppercase whitespace-nowrap ' +
        chain.bgColor +
        ' ' +
        chain.textColor
      }
    >
      {chain.name}
    </div>
  );
};
