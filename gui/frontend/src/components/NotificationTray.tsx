import { useToaster } from 'react-hot-toast/headless';

export const NotificationTray = () => {
  const { toasts, handlers } = useToaster();
  const { startPause, endPause, calculateOffset, updateHeight } = handlers;

  return (
    <div
      className="fixed top-2 left-2 w-48 z-100"
      onMouseEnter={startPause}
      onMouseLeave={endPause}
    >
      {toasts.map((toast) => {
        const offset = calculateOffset(toast, {
          reverseOrder: false,
          gutter: 8,
        });

        const ref = (el) => {
          if (el && typeof toast.height !== 'number') {
            const height = el.getBoundingClientRect().height;
            updateHeight(toast.id, height);
          }
        };
        return (
          <div
            key={toast.id}
            ref={ref}
            className="absolute w-48 bg-gray-200 rounded border-gray-400"
            style={{
              transition: 'all 0.5s ease-out',
              //   opacity: toast.visible ? 1 : 0,
              transform: `translateY(${offset}px)`,
            }}
            {...toast.ariaProps}
          >
            {toast.message}
          </div>
        );
      })}
    </div>
  );
};
