"use client";

// Confirmation sheet shown when a new user with a pending post-login event redirect tries to
// navigate away from /my_account before finishing profile setup.
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";

interface LeaveProfileSetupSheetProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirmLeave: () => void;
}

export function LeaveProfileSetupSheet({ isOpen, onClose, onConfirmLeave }: LeaveProfileSetupSheetProps) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title="Leave profile setup?">
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          You haven&apos;t finished setting up your profile yet. If you leave now, you won&apos;t be taken to the event you came here for.
        </p>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onConfirmLeave}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 transition-colors"
          >
            Leave
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors"
          >
            Stay
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
