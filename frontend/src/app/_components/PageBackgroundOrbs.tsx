// Fixed decorative glow orbs behind page content (login + app shell).
export default function PageBackgroundOrbs() {
  return (
    <>
      <div
        className="animate-drift glow-orb-blue pointer-events-none fixed rounded-full w-[600px] h-[600px] -top-[200px] -left-[200px]"
        aria-hidden
      />
      <div
        className="animate-drift-reverse glow-orb-orange pointer-events-none fixed rounded-full w-[600px] h-[600px] -bottom-[200px] -right-[200px]"
        aria-hidden
      />
    </>
  );
}
