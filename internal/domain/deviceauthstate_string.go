// Code generated by "stringer -type=DeviceAuthState -linecomment"; DO NOT EDIT.

package domain

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[DeviceAuthStateUndefined-0]
	_ = x[DeviceAuthStateInitiated-1]
	_ = x[DeviceAuthStateApproved-2]
	_ = x[DeviceAuthStateDenied-3]
	_ = x[DeviceAuthStateExpired-4]
	_ = x[DeviceAuthStateDone-5]
	_ = x[deviceAuthStateCount-6]
}

const _DeviceAuthState_name = "undefinedinitiatedapproveddeniedexpireddoneinvalid"

var _DeviceAuthState_index = [...]uint8{0, 9, 18, 26, 32, 39, 43, 50}

func (i DeviceAuthState) String() string {
	if i >= DeviceAuthState(len(_DeviceAuthState_index)-1) {
		return "DeviceAuthState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DeviceAuthState_name[_DeviceAuthState_index[i]:_DeviceAuthState_index[i+1]]
}
