import { useState } from 'react';
import { submitJob } from '../services/api';

type JobType = 'add_numbers' | 'reverse_string' | 'resize_image' | 'large_array_sum';

interface JobFormData {
  type: JobType;
  priority: number;
  thread_demand: number;
  payload: any;
}

export default function JobSubmitForm({ onSubmit }: { onSubmit?: () => void }) {
  const [isOpen, setIsOpen] = useState(false);
  const [jobType, setJobType] = useState<JobType>('add_numbers');
  const [priority, setPriority] = useState(1);
  const [threadDemand, setThreadDemand] = useState(1);
  const [payload, setPayload] = useState<any>({
    add_numbers: { x: 0, y: 0 },
    reverse_string: { text: '' },
    resize_image: { url: '', width: 800, height: 600 },
    large_array_sum: { array: [1, 2, 3, 4, 5] }
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await submitJob({
        type: jobType,
        priority,
        thread_demand: threadDemand,
        payload: payload[jobType]
      });
      setIsOpen(false);
      if (onSubmit) onSubmit();
    } catch (error) {
      console.error('Failed to submit job:', error);
      alert('Failed to submit job. Please try again.');
    }
  };

  const renderPayloadFields = () => {
    switch (jobType) {
      case 'add_numbers':
        return (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700">X Value</label>
              <input
                type="number"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                value={payload.add_numbers.x}
                onChange={(e) => setPayload({
                  ...payload,
                  add_numbers: { ...payload.add_numbers, x: parseInt(e.target.value) }
                })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Y Value</label>
              <input
                type="number"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                value={payload.add_numbers.y}
                onChange={(e) => setPayload({
                  ...payload,
                  add_numbers: { ...payload.add_numbers, y: parseInt(e.target.value) }
                })}
              />
            </div>
          </>
        );

      case 'reverse_string':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700">Text to Reverse</label>
            <input
              type="text"
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              value={payload.reverse_string.text}
              onChange={(e) => setPayload({
                ...payload,
                reverse_string: { text: e.target.value }
              })}
            />
          </div>
        );

      case 'resize_image':
        return (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700">Image URL</label>
              <input
                type="text"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                value={payload.resize_image.url}
                onChange={(e) => setPayload({
                  ...payload,
                  resize_image: { ...payload.resize_image, url: e.target.value }
                })}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Width</label>
                <input
                  type="number"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  value={payload.resize_image.width}
                  onChange={(e) => setPayload({
                    ...payload,
                    resize_image: { ...payload.resize_image, width: parseInt(e.target.value) }
                  })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Height</label>
                <input
                  type="number"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  value={payload.resize_image.height}
                  onChange={(e) => setPayload({
                    ...payload,
                    resize_image: { ...payload.resize_image, height: parseInt(e.target.value) }
                  })}
                />
              </div>
            </div>
          </>
        );

      case 'large_array_sum':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700">Array (comma-separated numbers)</label>
            <input
              type="text"
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              value={payload.large_array_sum.array.join(', ')}
              onChange={(e) => setPayload({
                ...payload,
                large_array_sum: {
                  array: e.target.value.split(',').map((n: string) => parseInt(n.trim())).filter((n: number) => !isNaN(n))
                }
              })}
            />
          </div>
        );
    }
  };

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(true)}
        className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
      >
        Create New Job
      </button>

      {isOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">Create New Job</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Job Type</label>
                <select
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  value={jobType}
                  onChange={(e) => setJobType(e.target.value as JobType)}
                >
                  <option value="add_numbers">Add Numbers</option>
                  <option value="reverse_string">Reverse String</option>
                  <option value="resize_image">Resize Image</option>
                  <option value="large_array_sum">Large Array Sum</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Priority</label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  value={priority}
                  onChange={(e) => setPriority(parseInt(e.target.value))}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Thread Demand</label>
                <input
                  type="number"
                  min="1"
                  max="8"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  value={threadDemand}
                  onChange={(e) => setThreadDemand(parseInt(e.target.value))}
                />
              </div>

              {renderPayloadFields()}

              <div className="flex justify-end gap-4 mt-6">
                <button
                  type="button"
                  onClick={() => setIsOpen(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md"
                >
                  Submit Job
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
