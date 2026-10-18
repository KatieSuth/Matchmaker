package store_test

import (
	"errors"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetPlayerAbsenceForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")

	cases := []struct {
		name               string
		absent             bool
		failOnQueryRowCall int
		failOnQueryCall    int
		failOnExecCall     int
		wantContains       string
	}{
		{name: "event meta fails", absent: true, failOnQueryRowCall: 1, wantContains: "get event group meta"},
		{name: "count lobbies fails", absent: true, failOnQueryRowCall: 2, wantContains: "count lobbies for event"},
		{name: "player placements fails", absent: true, failOnQueryCall: 1, wantContains: "get player placements"},
		{name: "upsert fails", absent: true, failOnExecCall: 1, wantContains: "upsert no-show"},
		{name: "delete fails", absent: false, failOnExecCall: 1, wantContains: "delete no-show"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := setupAbsenceFixture(t)
			faulty := &faultInjectTx{
				DBTX:               f.tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, tc.absent)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
