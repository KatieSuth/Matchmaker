// Tests for the viewer's own no-show notice: hidden with no record or when the event doesn't
// deprioritize, yellow under the limit, red at it.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, within } from "@/test/render";
import { NoShowNoticeBanner } from "./NoShowNoticeBanner";

const TZ = "UTC";
const TWO_DATES = ["2026-09-04T19:00:00Z", "2026-09-11T19:00:00Z"];

describe("NoShowNoticeBanner", () => {
  it("renders nothing without recent no-shows", () => {
    const { container } = render(
      <NoShowNoticeBanner noShowDates={[]} deprioritizeNoShows maxNoShows={2} timeZone={TZ} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the yellow notice with count and dates when under the limit", () => {
    render(<NoShowNoticeBanner noShowDates={TWO_DATES} deprioritizeNoShows maxNoShows={3} timeZone={TZ} />);

    const notice = screen.getByRole("region", { name: "No-show notice" });
    expect(notice).toHaveClass("border-amber-400/30");
    expect(notice).toHaveTextContent(
      "You have been marked as a no-show 2 times at this host's events in the last 6 months: Sep 4, 2026 and Sep 11, 2026.",
    );
    expect(notice).toHaveTextContent("You are not currently being deprioritized");
    expect(notice).toHaveTextContent("if you no-show again you may be deprioritized for future games");
    expect(notice).toHaveTextContent(
      "If you can't make a game you've registered for, edit your registration and un-toggle the game (remember to save or confirm the delete action!).",
    );
    expect(notice).toHaveTextContent(
      "If you've been placed on a team and can't make it, let the host know as soon as possible so a substitute can be found in time.",
    );
    expect(notice).toHaveTextContent("If you believe an absence was marked by mistake, let the host know.");
    expect(screen.queryByRole("region", { name: "No-show warning" })).not.toBeInTheDocument();
  });

  it("renders nothing when the event does not deprioritize no-shows, whatever the stored limit", () => {
    const { container } = render(
      <NoShowNoticeBanner noShowDates={TWO_DATES} deprioritizeNoShows={false} maxNoShows={1} timeZone={TZ} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("uses singular wording for one no-show", () => {
    render(
      <NoShowNoticeBanner noShowDates={[TWO_DATES[0]]} deprioritizeNoShows maxNoShows={3} timeZone={TZ} />,
    );
    expect(screen.getByRole("region")).toHaveTextContent("marked as a no-show 1 time at this host's events in the last 6 months: Sep 4, 2026.");
  });

  it("shows the red warning with the event's limit once the count meets it", () => {
    render(<NoShowNoticeBanner noShowDates={TWO_DATES} deprioritizeNoShows maxNoShows={2} timeZone={TZ} />);

    const warning = screen.getByRole("region", { name: "No-show warning" });
    expect(warning).toHaveClass("border-[var(--color-text-danger)]/30");
    expect(warning).toHaveTextContent("marked as a no-show 2 times");
    expect(warning).toHaveTextContent("Sep 4, 2026 and Sep 11, 2026");
    const items = within(warning).getAllByRole("listitem").map((li) => li.textContent);
    expect(items).toEqual([
      "The rules for this event deprioritize players with 2 or more no-shows, which makes it less likely (but not impossible) you will be placed on a team. You may still be called on to substitute, so register if you are interested in playing.",
      "If you can't make a game you've registered for, edit your registration and un-toggle the game (remember to save or confirm the delete action!).",
      "If you've been placed on a team and can't make it, let the host know as soon as possible so a substitute can be found in time.",
      "If you believe an absence was marked by mistake, let the host know.",
    ]);
    expect(screen.queryByRole("region", { name: "No-show notice" })).not.toBeInTheDocument();
  });

  it("lists three or more dates with a serial conjunction", () => {
    render(
      <NoShowNoticeBanner
        noShowDates={[...TWO_DATES, "2026-09-18T19:00:00Z"]}
        deprioritizeNoShows
        maxNoShows={5}
        timeZone={TZ}
      />,
    );
    expect(screen.getByRole("region")).toHaveTextContent("Sep 4, 2026, Sep 11, 2026, and Sep 18, 2026.");
  });

  it("has no accessibility violations in either state", async () => {
    const { container, rerender } = render(
      <NoShowNoticeBanner noShowDates={TWO_DATES} deprioritizeNoShows maxNoShows={3} timeZone={TZ} />,
    );
    expect(await axe(container)).toHaveNoViolations();
    rerender(<NoShowNoticeBanner noShowDates={TWO_DATES} deprioritizeNoShows maxNoShows={2} timeZone={TZ} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
