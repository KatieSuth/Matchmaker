"use client";

// Loading placeholder mirroring EventCard's layout, shown while the active bucket loads.
//
// A "skeleton" card is a content-agnostic stand-in for a real EventCard: instead of a spinner or
// blank space, it renders gray blocks shaped like the real card's layout (title, badge, detail
// grid) so the page's structure doesn't jump/reflow once data arrives. This is purely presentational
// — it takes no props and holds no state — so usage is just "render N of these while loading":
//
//   {isInitialLoading ? (
//     <>
//       <SkeletonCard />
//       <SkeletonCard />
//       <SkeletonCard />
//     </>
//   ) : (
//     activeBucket.events.map((event) => <EventCard key={event.id} event={event} ... />)
//   )}
//
// Callers are responsible for deciding *how many* to show (my_events/page.tsx hardcodes 3 to
// approximate a typical first page) and for swapping them out for the real list once the fetch
// settles; this component only owns the "what does one loading row look like" concern.
//
// Keep this in sync with EventCard's structure: if EventCard's layout changes (e.g. a new row is
// added to the details grid, or the header gains an extra badge), update the placeholder blocks
// here too so the loading state still visually matches the loaded state.
export function SkeletonCard() {
  return (
    <div className="card rounded-xl p-4 flex flex-col gap-3 animate-pulse">
      {/* Mirrors EventCard's header row: title/subtitle on the left, status badge on the right. */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1.5 flex-1">
          <div className="h-4 w-2/5 rounded bg-white/8" />
          <div className="h-3 w-1/4 rounded bg-white/5" />
        </div>
        <div className="h-5 w-14 rounded-full bg-white/5" />
      </div>
      {/* Mirrors EventCard's divider between the header and the details grid. */}
      <div className="h-px bg-white/5" />
      {/* Mirrors EventCard's 3-column details grid (Date / Host / Players). */}
      <div className="grid grid-cols-3 gap-3">
        <div className="h-3 rounded bg-white/5" />
        <div className="h-3 rounded bg-white/5" />
        <div className="h-3 rounded bg-white/5" />
      </div>
    </div>
  );
}
