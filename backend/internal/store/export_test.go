package store

// Expose unexported cursor helpers for black-box tests in store_test.
type DashboardCursorForTest = dashboardCursor

var EncodeDashboardCursorForTest = encodeDashboardCursor
var DecodeDashboardCursorForTest = decodeDashboardCursor
