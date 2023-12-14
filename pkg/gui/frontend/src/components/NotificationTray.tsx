import { toast, Toaster, ToastBar } from 'react-hot-toast';
import { IonIcon } from '@ionic/react';
import { closeOutline } from 'ionicons/icons';

export const NotificationTray: React.FC = () => {
  return (
    <Toaster
      position="top-right"
      toastOptions={{
        style: {
          width: '400px',
          padding: '0px',
        },
        duration: 8000,
        success: {
          className: 'border border-green-400',
        },
        error: {
          className: 'border border-red-400',
        },
      }}
    >
      {(t) => (
        <>
          <ToastBar toast={t}>
            {({ icon, message }) => (
              <div className="rounded w-full overflow-hidden">
                <div className="flex w-full justify-between px-4 py-2">
                  <div className="flex">
                    {icon}
                    {message}
                  </div>
                  {t.type !== 'loading' && (
                    <button
                      className="flex items-center justify-self-end"
                      onClick={() => toast.dismiss(t.id)}
                    >
                      <IonIcon icon={closeOutline} />
                    </button>
                  )}
                </div>
                {t.type !== 'loading' && (
                  <div
                    className={
                      'border-b-3 animate-dismissIndicator w-full ' +
                      String(t.className)
                    }
                  ></div>
                )}
              </div>
            )}
          </ToastBar>
        </>
      )}
    </Toaster>
  );
};
