// Jump back to the page header from between About / Privacy / FAQ cards.
interface BackToTopProps {
  className?: string;
}

export default function BackToTop({ className = "" }: BackToTopProps) {
  return (
    <div className={`flex justify-center ${className}`.trim()}>
      <a href="#top" className="body-link text-sm">
        Back to top
      </a>
    </div>
  );
}
