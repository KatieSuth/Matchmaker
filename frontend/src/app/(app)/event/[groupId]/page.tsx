"use client";

// Event group detail: metadata, per-game registration panels, host controls (teams, registration),
// and participant registration / profile actions. Composes hooks + presentational pieces from
// ./_hooks, ./_components, and ./_lib; this file is orchestration only.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { LobbyHostAssignmentBanner } from "@/app/_components/LobbyHostAssignmentBanner";
import { useAuth } from "@/app/_context/AuthContext";
import { useCopyStatus } from "@/app/_hooks/useCopyStatus";
import { DEFAULT_FEEDBACK_URL } from "@/app/_lib/constants";
import { buildDiscordPingMessage } from "@/app/_lib/discordPings";
import { EventGroupDetail, EventRegistration } from "@/app/_types/types";
import { EventGroupHeaderCard } from "./_components/EventGroupHeaderCard";
import { GameTabsStrip } from "./_components/GameTabsStrip";
import {
  RegistrationEditorForm,
  RegistrationSaveFooter,
  RegistrationToggleButton,
} from "./_components/RegistrationEditorPanel";
import { EventPanel } from "./_components/EventPanel";
import { TeamsPanel } from "./_components/TeamsPanel";
import { DeleteRegistrationSheet } from "./_components/sheets/DeleteRegistrationSheet";
import { DeleteTeamsWarningSheet } from "./_components/sheets/DeleteTeamsWarningSheet";
import { EditEventSheet } from "./_components/sheets/EditEventSheet";
import { JoinLobbySheet } from "./_components/sheets/JoinLobbySheet";
import { LobbyHostConfirmSheet } from "./_components/sheets/LobbyHostConfirmSheet";
import { MoveToSubsSheet } from "./_components/sheets/MoveToSubsSheet";
import { RegistrationDetailsSheet } from "./_components/sheets/RegistrationDetailsSheet";
import { SubCapacitySheet } from "./_components/sheets/SubCapacitySheet";
import { SwapPlayerSheet } from "./_components/sheets/SwapPlayerSheet";
import { useDeleteRegistration } from "./_hooks/useDeleteRegistration";
import { useEventGroupData } from "./_hooks/useEventGroupData";
import { useHostRosterActions } from "./_hooks/useHostRosterActions";
import { useJoinLobbySheet } from "./_hooks/useJoinLobbySheet";
import { useRegistrationEditor } from "./_hooks/useRegistrationEditor";
import { discordLockLeadSentence, formatDateTime, formatPlayerCount, joinDiscordGuildNames } from "./_lib/formatters";
import { findUserLobbyHostAssignments } from "./_lib/placements";

export default function EventGroupPage() {
  const params = useParams<{ groupId: string }>();
  const groupId = Array.isArray(params?.groupId) ? params.groupId[0] : params?.groupId;
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const {
    group,
    setGroup,
    pageError,
    setPageError,
    accessChecking,
    accessDenial,
    setAccessDenial,
    loading,
    gameRanks,
    loadGroup,
    activeEventId,
    setActiveEventId,
  } = useEventGroupData(groupId, user, authLoading, isAuthenticated);

  const [working, setWorking] = useState(false);
  const [showAllEvents, setShowAllEvents] = useState(true);
  const [editSheetOpen, setEditSheetOpen] = useState(false);
  const [detailsSheetOpen, setDetailsSheetOpen] = useState(false);
  const [selectedRegistration, setSelectedRegistration] = useState<EventRegistration | null>(null);
  const [toast, setToast] = useState<string | null>(null);
  const topAnchorRef = useRef<HTMLDivElement | null>(null);
  const eventSectionRefs = useRef<Record<string, HTMLDivElement | null>>({});
  const isHost = !!(group && user?.id && user.id === group.owner_id);
  const hasAnyLobbies = !!group?.events.some((event) => event.lobbies_count > 0);

  const myRegistrationsByEvent = useMemo(() => {
    const map = new Map<string, EventRegistration>();
    if (!group || !user?.id) return map;
    for (const event of group.events) {
      const registration = event.registrations.find((item) => item.user_id === user.id);
      if (registration) {
        map.set(event.id, registration);
      }
    }
    return map;
  }, [group, user]);

  const activeEvent = useMemo(
    () => group?.events.find((event) => event.id === activeEventId) ?? group?.events[0] ?? null,
    [activeEventId, group?.events],
  );

  const activeEventNumber = useMemo(() => {
    if (!group || !activeEvent) return 1;
    const idx = group.events.findIndex((event) => event.id === activeEvent.id);
    return idx >= 0 ? idx + 1 : 1;
  }, [activeEvent, group]);

  const myLobbyHostAssignments = useMemo(() => {
    if (!group || !user?.id) return [];
    return findUserLobbyHostAssignments(group.events, user.id);
  }, [group, user]);

  const scrollToEventSection = useCallback((eventId: string) => {
    eventSectionRefs.current[eventId]?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }, []);

  const scrollToTop = useCallback(() => {
    topAnchorRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }, []);

  const firstEventStart = group?.events[0]?.start_time ?? "";

  const deleteRegistration = useDeleteRegistration(group, user, loadGroup, setWorking, setPageError);
  const registrationEditor = useRegistrationEditor({
    groupId,
    group,
    user,
    authLoading,
    isAuthenticated,
    pageLoading: loading,
    isHost,
    myRegistrationsByEvent,
    hasAnyLobbies,
    working,
    setWorking,
    loadGroup,
    setAccessDenial,
    setGroup,
    openDeleteAllForCurrentUserConfirmation: deleteRegistration.openDeleteAllForCurrentUserConfirmation,
  });
  const hostActions = useHostRosterActions(group, loadGroup, setWorking, setPageError);
  const joinLobby = useJoinLobbySheet(group, user, isHost, loadGroup, setWorking, setToast);

  const { status: shareStatus, copy: copyShareLink } = useCopyStatus();
  const { status: pingStatus, copy: copyDiscordPings } = useCopyStatus();

  const handleShare = () => {
    const shareUrl = typeof window !== "undefined" ? window.location.href : "";
    void copyShareLink(shareUrl);
  };

  const handleCopyDiscordPings = () => {
    if (!group) return;
    void copyDiscordPings(buildDiscordPingMessage(group));
  };

  const handleShowDetails = (registration: EventRegistration) => {
    setSelectedRegistration(registration);
    setDetailsSheetOpen(true);
  };

  useEffect(() => {
    if (!toast) return;
    const timer = window.setTimeout(() => setToast(null), 2500);
    return () => window.clearTimeout(timer);
  }, [toast]);

  const feedbackUrl = process.env.NEXT_PUBLIC_FEEDBACK_URL || DEFAULT_FEEDBACK_URL;

  if (authLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">Loading event group...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">Please sign in to view this page.</p>
      </div>
    );
  }

  if (accessChecking || (loading && !accessDenial && !pageError && !group)) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">
          {accessChecking ? "Checking Discord server membership..." : "Loading event group..."}
        </p>
      </div>
    );
  }

  if (accessDenial) {
    const guildNames = accessDenial.discord_guilds.map((g) => g.name).filter(Boolean);
    const serverNames = joinDiscordGuildNames(guildNames);
    const title = accessDenial.event_title.trim();
    const singular = accessDenial.discord_guilds.length <= 1;
    const lead = discordLockLeadSentence(accessDenial.event_named, title, serverNames);
    const membership = singular
      ? "You need to be a member of that Discord server to open it or register."
      : "You need to be a member of at least one of those Discord servers to open it or register.";
    const serverPhrase = singular ? "this Discord server" : "these Discord servers";
    return (
      <div className="flex-1 flex flex-col items-center justify-center px-4 py-8">
        <div className="card w-full max-w-lg rounded-xl p-6 flex flex-col gap-3">
          <h1 className="text-lg font-semibold text-[var(--color-text)]">This event is locked</h1>
          <p className="text-sm leading-relaxed text-[var(--color-text-muted)]">
            {lead} {membership}
          </p>
          <p className="text-xs leading-relaxed text-[var(--color-text-muted)]">
            If you are a member of {serverPhrase} and you believe this message is an error, check{" "}
            <a
              href="https://discordstatus.com"
              className="body-link"
              target="_blank"
              rel="noopener noreferrer"
            >
              discordstatus.com
            </a>
            , or send feedback via{" "}
            <a
              href={feedbackUrl}
              className="body-link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Issues/Feedback
            </a>
            .
          </p>
        </div>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-danger)]">{pageError ?? "Event group not found."}</p>
      </div>
    );
  }

  const hasVisibleGames = registrationEditor.registrationEditorOpen || showAllEvents || (!showAllEvents && !!activeEvent);

  return (
    <div className="flex-1 flex flex-col items-center px-4 py-8">
      <div className="w-full max-w-3xl flex flex-col gap-5" style={{ animation: "var(--animate-rise)" }}>
        <div ref={topAnchorRef} />

        <EventGroupHeaderCard
          group={group}
          isHost={isHost}
          hasAnyLobbies={hasAnyLobbies}
          working={working}
          firstEventStart={firstEventStart}
          pingStatus={pingStatus}
          onCopyDiscordPings={handleCopyDiscordPings}
          shareStatus={shareStatus}
          onShare={handleShare}
          onOpenEditSheet={() => setEditSheetOpen(true)}
          onLockInClick={() => {
            if (!group.registration_open && hasAnyLobbies) {
              hostActions.setWarningSheetOpen(true);
              return;
            }
            void hostActions.lockInTeams();
          }}
        />

        <LobbyHostAssignmentBanner assignments={myLobbyHostAssignments} />

        <GameTabsStrip
          events={group.events}
          activeEventId={activeEvent?.id ?? null}
          showAllEvents={showAllEvents}
          registrationEditorOpen={registrationEditor.registrationEditorOpen}
          onToggleShowAll={() => setShowAllEvents((v) => !v)}
          onSelectEvent={(eventId) => {
            setActiveEventId(eventId);
            setShowAllEvents(false);
          }}
          onScrollToEvent={scrollToEventSection}
        />

        {hasVisibleGames && (
          <RegistrationToggleButton
            group={group}
            registrationEditorOpen={registrationEditor.registrationEditorOpen}
            working={working}
            myRegistrationsCount={myRegistrationsByEvent.size}
            onOpen={() => void registrationEditor.handleOpenRegistrationSheet()}
            onClose={() => void registrationEditor.handleCloseRegistrationEditor()}
          />
        )}

        {pageError && (
          <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
            {pageError}
          </p>
        )}

        {registrationEditor.registrationEditorOpen ? (
          <RegistrationEditorForm
            group={group}
            regionMismatchWarning={registrationEditor.regionMismatchWarning}
            myRegistrationsCount={myRegistrationsByEvent.size}
            userGameDraft={registrationEditor.userGameDraft}
            setUserGameDraft={registrationEditor.setUserGameDraft}
            userGameRanks={registrationEditor.userGameRanks}
            registrationLoading={registrationEditor.registrationLoading}
            userGameErrors={registrationEditor.userGameErrors}
            registrationDraft={registrationEditor.registrationDraft}
            setRegistrationDraft={registrationEditor.setRegistrationDraft}
            selectedValidEventIds={registrationEditor.selectedValidEventIds}
            canDeleteAllViaSave={registrationEditor.canDeleteAllViaSave}
            registrationError={registrationEditor.registrationError}
          />
        ) : showAllEvents ? (
          <div className="flex flex-col gap-4">
            {group.events.map((event, index) => (
              <div key={event.id}>
                {index > 0 && <hr className="mb-4 border-0 border-t-2 border-white/20" />}
                <div
                  ref={(node) => {
                    eventSectionRefs.current[event.id] = node;
                  }}
                  className="flex flex-col gap-3 scroll-mt-24"
                >
                  <h2 className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-x-2 text-base font-semibold text-[var(--color-text)]">
                    <span className="min-w-0 truncate">
                      Game {index + 1}
                      <span className="font-normal text-[var(--color-text-muted)]"> · {event.game_mode_name}</span>
                    </span>
                    <span className="shrink-0 whitespace-nowrap text-center text-xs font-normal text-[var(--color-text-muted)]">
                      {formatDateTime(event.start_time)}
                    </span>
                    <span className="shrink-0 justify-self-end whitespace-nowrap text-right text-xs font-normal text-[var(--color-text-soft)]">
                      {formatPlayerCount(event.registrations.length)}
                    </span>
                  </h2>
                  {!group.registration_open && event.lobbies_count > 0 ? (
                    <TeamsPanel
                      event={event}
                      gameNumber={index + 1}
                      eventRegion={group.region}
                      currentUserRegion={user?.region}
                      isHostView={isHost}
                      currentUserId={user?.id}
                      gameRanks={gameRanks}
                      onShowDetails={handleShowDetails}
                      onDeleteRegistrationForGame={(registration, gameNumber) =>
                        deleteRegistration.openDeleteConfirmation(registration, gameNumber, "single")
                      }
                      onDeleteAllFromUser={(registration, gameNumber) =>
                        deleteRegistration.openDeleteConfirmation(registration, gameNumber, "all")
                      }
                      onSwapPlayer={isHost ? hostActions.openSwapSheet : undefined}
                      onMoveToUnplaced={isHost ? hostActions.handleMoveToUnplaced : undefined}
                      onMoveToSubs={isHost ? hostActions.handleMoveToSubs : undefined}
                      onMakeLobbyHost={isHost ? hostActions.handleMakeLobbyHost : undefined}
                      onJoinLobby={joinLobby.openJoinLobbySheet}
                      showJoinLobby={isHost || myRegistrationsByEvent.size > 0}
                    />
                  ) : (
                    <EventPanel
                      event={event}
                      gameNumber={index + 1}
                      eventRegion={group.region}
                      currentUserRegion={user?.region}
                      isHostView={isHost}
                      currentUserId={user?.id}
                      onShowDetails={handleShowDetails}
                      onDeleteRegistrationForGame={(registration, gameNumber) =>
                        deleteRegistration.openDeleteConfirmation(registration, gameNumber, "single")
                      }
                      onDeleteAllFromUser={(registration, gameNumber) =>
                        deleteRegistration.openDeleteConfirmation(registration, gameNumber, "all")
                      }
                    />
                  )}
                  <div className="flex justify-end">
                    <button
                      type="button"
                      onClick={scrollToTop}
                      className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-white/[0.08] hover:text-[var(--color-text-soft)]"
                    >
                      Back to top
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : activeEvent ? (
          <div className="flex flex-col gap-3">
            <h2 className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-x-2 text-base font-semibold text-[var(--color-text)]">
              <span className="min-w-0 truncate">
                Game {activeEventNumber}
                <span className="font-normal text-[var(--color-text-muted)]"> · {activeEvent.game_mode_name}</span>
              </span>
              <span className="shrink-0 whitespace-nowrap text-center text-xs font-normal text-[var(--color-text-muted)]">
                {formatDateTime(activeEvent.start_time)}
              </span>
              <span className="shrink-0 justify-self-end whitespace-nowrap text-right text-xs font-normal text-[var(--color-text-soft)]">
                {formatPlayerCount(activeEvent.registrations.length)}
              </span>
            </h2>
            {!group.registration_open && activeEvent.lobbies_count > 0 ? (
              <TeamsPanel
                event={activeEvent}
                gameNumber={activeEventNumber}
                eventRegion={group.region}
                currentUserRegion={user?.region}
                isHostView={isHost}
                currentUserId={user?.id}
                gameRanks={gameRanks}
                onShowDetails={handleShowDetails}
                onDeleteRegistrationForGame={(registration, gameNumber) =>
                  deleteRegistration.openDeleteConfirmation(registration, gameNumber, "single")
                }
                onDeleteAllFromUser={(registration, gameNumber) =>
                  deleteRegistration.openDeleteConfirmation(registration, gameNumber, "all")
                }
                onSwapPlayer={isHost ? hostActions.openSwapSheet : undefined}
                onMoveToUnplaced={isHost ? hostActions.handleMoveToUnplaced : undefined}
                onMoveToSubs={isHost ? hostActions.handleMoveToSubs : undefined}
                onMakeLobbyHost={isHost ? hostActions.handleMakeLobbyHost : undefined}
                onJoinLobby={joinLobby.openJoinLobbySheet}
                showJoinLobby={isHost || myRegistrationsByEvent.size > 0}
              />
            ) : (
              <EventPanel
                event={activeEvent}
                gameNumber={activeEventNumber}
                eventRegion={group.region}
                currentUserRegion={user?.region}
                isHostView={isHost}
                currentUserId={user?.id}
                onShowDetails={handleShowDetails}
                onDeleteRegistrationForGame={(registration, gameNumber) =>
                  deleteRegistration.openDeleteConfirmation(registration, gameNumber, "single")
                }
                onDeleteAllFromUser={(registration, gameNumber) =>
                  deleteRegistration.openDeleteConfirmation(registration, gameNumber, "all")
                }
              />
            )}
            <div className="flex justify-end">
              <button
                type="button"
                onClick={scrollToTop}
                className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-white/[0.08] hover:text-[var(--color-text-soft)]"
              >
                Back to top
              </button>
            </div>
          </div>
        ) : null}

        {registrationEditor.registrationEditorOpen && (
          <RegistrationSaveFooter
            working={working}
            canDeleteAllViaSave={registrationEditor.canDeleteAllViaSave}
            hasUserGameErrors={registrationEditor.hasUserGameErrors}
            canSubmitRegistration={registrationEditor.canSubmitRegistration}
            onSave={() => void registrationEditor.handleSaveRegistration()}
          />
        )}
      </div>

      <EditEventSheet
        isOpen={editSheetOpen}
        onClose={() => setEditSheetOpen(false)}
        isHost={isHost}
        group={group}
        onSubmitted={() => {
          void loadGroup();
          setToast("Event settings updated.");
        }}
      />

      <RegistrationDetailsSheet
        isOpen={detailsSheetOpen}
        onClose={() => setDetailsSheetOpen(false)}
        registration={selectedRegistration}
      />

      <DeleteRegistrationSheet
        isOpen={deleteRegistration.deleteWarningSheetOpen}
        onClose={deleteRegistration.closeDeleteWarningSheet}
        pendingDeleteAction={deleteRegistration.pendingDeleteAction}
        deletingSelf={deleteRegistration.deletingSelf}
        working={working}
        onConfirm={() => void deleteRegistration.handleDeleteRegistration()}
      />

      <LobbyHostConfirmSheet
        isOpen={hostActions.lobbyHostConfirmOpen}
        onClose={hostActions.closeLobbyHostConfirm}
        pendingLobbyHostChange={hostActions.pendingLobbyHostChange}
        working={working}
        onConfirm={() => {
          if (hostActions.pendingLobbyHostChange) {
            void hostActions.submitLobbyHostChange(hostActions.pendingLobbyHostChange.placement);
          }
        }}
      />

      <SwapPlayerSheet
        isOpen={hostActions.swapSheetOpen}
        onClose={hostActions.closeSwapSheet}
        pendingSwap={hostActions.pendingSwap}
        swapTargetUserId={hostActions.swapTargetUserId}
        onChangeSwapTarget={(value) => {
          hostActions.setSwapTargetUserId(value);
          hostActions.setSwapError(null);
        }}
        swapCandidateOptions={hostActions.swapCandidateOptions}
        swapError={hostActions.swapError}
        working={working}
        onSubmit={() => void hostActions.handleSwapSubmit()}
      />

      <MoveToSubsSheet
        isOpen={hostActions.moveToSubsSheetOpen}
        onClose={hostActions.closeMoveToSubsSheet}
        pendingMoveToSubs={hostActions.pendingMoveToSubs}
        moveToSubsLobbyId={hostActions.moveToSubsLobbyId}
        onChangeLobby={(value) => {
          hostActions.setMoveToSubsLobbyId(value);
          hostActions.setMoveToSubsError(null);
        }}
        moveToSubsLobbyOptions={hostActions.moveToSubsLobbyOptions}
        moveToSubsError={hostActions.moveToSubsError}
        working={working}
        onSubmit={() => void hostActions.handleMoveToSubsSubmit()}
      />

      <JoinLobbySheet
        isOpen={joinLobby.joinLobbySheetOpen}
        onClose={joinLobby.closeJoinLobbySheet}
        pendingJoinLobby={joinLobby.pendingJoinLobby}
        canEdit={joinLobby.canEditPendingJoinLobby}
        joinLobbyDraft={joinLobby.joinLobbyDraft}
        onChangeDraft={joinLobby.handleJoinLobbyDraftChange}
        joinLobbyError={joinLobby.joinLobbyError}
        copyStatus={joinLobby.joinLobbyCopyStatus}
        onCopy={() => void joinLobby.handleCopyJoinLobby()}
        onSave={() => void joinLobby.handleSaveJoinLobby()}
        pendingJoinDisplay={joinLobby.pendingJoinDisplay}
        joinLinkBase={group.join_link_base}
        working={working}
      />

      <SubCapacitySheet
        isOpen={hostActions.subCapacitySheetOpen}
        onClose={() => hostActions.setSubCapacitySheetOpen(false)}
      />

      <DeleteTeamsWarningSheet
        isOpen={hostActions.warningSheetOpen}
        onClose={() => hostActions.setWarningSheetOpen(false)}
        working={working}
        onConfirm={hostActions.confirmDeleteTeams}
      />

      {toast && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 rounded-lg border border-white/10 bg-[var(--color-bg-soft)] px-3 py-2 text-xs text-[var(--color-text-soft)] shadow-[0_18px_45px_rgba(0,0,0,0.45)]">
          {toast}
        </div>
      )}
    </div>
  );
}
