export default function AlertCard({ reminder }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <svg className="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
      </svg>
      <span className="text-yellow-800 dark:text-yellow-300 font-medium">{reminder.title}</span>
    </div>
  );
}
