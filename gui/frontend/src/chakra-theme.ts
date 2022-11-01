// theme.ts (tsx file with usage of StyleFunctions, see 4.)
import {
  extendTheme,
  withDefaultSize,
  withDefaultVariant,
} from '@chakra-ui/react';

// TODO: How do we set the default bgColor = gray.700?

const theme = extendTheme(
  withDefaultSize({
    size: 'lg',
    components: ['Button', 'Input', 'NumberInput', 'PinInput', 'Select'],
  }),
  withDefaultVariant({
    variant: 'solid',
    components: ['Button', 'Input', 'NumberInput', 'PinInput', 'Select'],
  }),
  {
    colors: {
      darkGray: {
        100: '#3D3D3D',
      },
    },
  }
);

export default theme;
