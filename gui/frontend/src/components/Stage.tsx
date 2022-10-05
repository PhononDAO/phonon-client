import toast, { useToaster } from 'react-hot-toast';

export const Stage = () => {
  const notify = (type) => {
    if (type === 'success') {
      toast.success('Some successful notification message for everyone!');
    } else {
      toast.error('Some notification message');
    }
  };

  return (
    <div className="flex gap-x-2">
      <button
        className="block px-4 py-2 bg-gray-200 border border-gray-400 rounded"
        onClick={() => {
          notify('success');
        }}
      >
        Send success
      </button>
      <button
        className="block px-4 py-2 bg-gray-200 border border-gray-400 rounded"
        onClick={() => {
          notify('error');
        }}
      >
        Send error
      </button>
    </div>
  );
};
