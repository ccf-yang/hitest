//go:build !xpack && !enterprise

package xpack

import "hitest/internal/agent/utils/xpack/helper"

var MultiNodeProvider = helper.NewIMultiNodeProvider()
