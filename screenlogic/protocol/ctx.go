// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"io"
	"log/slog"
)

type ctxKey string

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey("logger"), logger)
}

var discardLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))

func Logger(ctx context.Context) *slog.Logger {
	l := ctx.Value(ctxKey("logger"))
	if l == nil {
		return discardLogger
	}
	return l.(*slog.Logger)
}
