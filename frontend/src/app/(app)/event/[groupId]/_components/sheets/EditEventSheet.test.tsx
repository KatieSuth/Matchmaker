// Tests for the event-settings sheet wrapping EventForm in host (editable) vs. non-host (read-only)
// mode. Deeper EventForm behavior itself is covered by EventForm.test.tsx.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildEventGroupDetail, buildEventGroupEvent, buildUser } from "@/test/fixtures";
import { renderWithProviders, screen } from "@/test/render";
import { EditEventSheet } from "./EditEventSheet";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const host = buildUser({ id: "host-1" });

describe("EditEventSheet", () => {
  it("titles the sheet 'Edit event settings' for the host", () => {
    const group = buildEventGroupDetail({ owner_id: "host-1", events: [buildEventGroupEvent()] });
    renderWithProviders(<EditEventSheet isOpen={true} onClose={vi.fn()} isHost={true} group={group} onSubmitted={vi.fn()} />, {
      user: host,
      isAuthenticated: true,
    });
    expect(screen.getByRole("dialog", { name: "Edit event settings" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save Settings" })).toBeInTheDocument();
  });

  it("titles the sheet 'Event settings' and shows only Close for a non-host viewer", () => {
    const group = buildEventGroupDetail({ owner_id: "host-1", events: [buildEventGroupEvent()] });
    renderWithProviders(<EditEventSheet isOpen={true} onClose={vi.fn()} isHost={false} group={group} onSubmitted={vi.fn()} />, {
      user: buildUser({ id: "viewer-1" }),
      isAuthenticated: true,
    });
    expect(screen.getByRole("dialog", { name: "Event settings" })).toBeInTheDocument();
    // The sheet's own backdrop is also an (aria-labeled, textless) "Close" button, so disambiguate
    // by visible text content to target the footer button specifically.
    expect(screen.getByText("Close")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save Settings" })).not.toBeInTheDocument();
  });

  it("pre-fills the event name from the group", () => {
    const group = buildEventGroupDetail({ owner_id: "host-1", name: "Friday Customs", events: [buildEventGroupEvent()] });
    renderWithProviders(<EditEventSheet isOpen={true} onClose={vi.fn()} isHost={true} group={group} onSubmitted={vi.fn()} />, {
      user: host,
      isAuthenticated: true,
    });
    expect(screen.getByDisplayValue("Friday Customs")).toBeInTheDocument();
  });

  it("has no accessibility violations for a host", async () => {
    const group = buildEventGroupDetail({ owner_id: "host-1", events: [buildEventGroupEvent()] });
    renderWithProviders(
      <EditEventSheet isOpen={true} onClose={vi.fn()} isHost={true} group={group} onSubmitted={vi.fn()} />,
      { user: host, isAuthenticated: true },
    );
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });

  it("has no accessibility violations for a non-host viewer", async () => {
    const group = buildEventGroupDetail({ owner_id: "host-1", events: [buildEventGroupEvent()] });
    renderWithProviders(
      <EditEventSheet isOpen={true} onClose={vi.fn()} isHost={false} group={group} onSubmitted={vi.fn()} />,
      { user: buildUser({ id: "viewer-1" }), isAuthenticated: true },
    );
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
