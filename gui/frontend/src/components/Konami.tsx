import { useState, useRef, useEffect } from 'react';
import { DateTime } from 'luxon';

export const Konami = () => {
  const keyTimeDelay = 500;
  const [konamiCode] = useState([
    'ArrowUp',
    'ArrowDown',
    'ArrowDown',
    'ArrowLeft',
    'ArrowRight',
    'ArrowLeft',
    'ArrowRight',
    'b',
    'a',
  ]);

  const [lastKeyPress, _setLastKeyPress] = useState(DateTime.now().toMillis());
  const [sequenceBuffer, setSequenceBuffer] = useState<Array<string>>([]);

  // we have to use a reference for the doc event listener, in coop with the state update
  const lastKeyPressStateRef = useRef(lastKeyPress);
  const setLastKeyPress = (data) => {
    lastKeyPressStateRef.current = data;
    _setLastKeyPress(data);
  };

  // now we refresh the last key time and check if the sequence is correct
  useEffect(() => {
    setLastKeyPress(DateTime.now().toMillis());

    // lets see if the sequence is correct
    if (
      konamiCode.join('').toLowerCase() ===
      sequenceBuffer.join('').toLowerCase()
    ) {
      alert('FUTURE EASTER EGG ENABLED!');
    }
  }, [konamiCode, sequenceBuffer]);

  // let's handle the key press and add it to a sequence buffer, if within the permissable key time delay
  useEffect(() => {
    const keyPressHandler = (event) => {
      if (
        DateTime.now().toMillis() - lastKeyPressStateRef.current <=
        keyTimeDelay
      ) {
        setSequenceBuffer((sequenceBuffer) => [
          ...sequenceBuffer,
          String(event.key),
        ]);
      } else {
        setSequenceBuffer([]);
      }
    };

    document.addEventListener('keydown', keyPressHandler, false);

    return () => {
      document.removeEventListener('keydown', keyPressHandler, false);
    };
  }, []);

  return <></>;
};
