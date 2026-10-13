//go:build !xpack && !enterprise

package xpack

import "hitest/internal/core/utils/xpack/helper"

var MultiNodeProvider = helper.NewIMultiNodeProvider()
