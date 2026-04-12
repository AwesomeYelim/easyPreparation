"use client";

export default function GlobalError({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <html>
      <body className="flex flex-col items-center justify-center min-h-screen gap-4">
        <h2 className="text-lg font-semibold text-red-600">오류가 발생했습니다</h2>
        <p className="text-sm text-slate-500">{error.message}</p>
        <button onClick={reset} className="px-4 py-2 text-sm rounded-lg bg-blue-500 text-white hover:bg-blue-600">
          다시 시도
        </button>
      </body>
    </html>
  );
}
