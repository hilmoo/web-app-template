package app_test

import (
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"web-app-template-core/internal/app"
	"web-app-template-core/internal/config"
	"web-app-template-core/internal/helper"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func testParams(t *testing.T) helper.HandlerParams {
	t.Helper()
	return helper.HandlerParams{
		Logger:   slog.New(slog.DiscardHandler),
		Validate: nil,
		Pool:     nil,
		Queries:  nil,
		Config:   config.Config{},
	}
}

func TestAllServicesWired(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	registered := map[string]bool{}
	for _, n := range app.RegisterConnect(testParams(t), mux) {
		registered[n] = true
	}

	protoregistry.GlobalFiles.RangeFilesByPackage(
		protoreflect.FullName("services.v1"),
		func(fd protoreflect.FileDescriptor) bool {
			services := fd.Services()
			for i := range services.Len() {
				name := string(services.Get(i).FullName())
				if !registered[name] {
					t.Errorf("service %s is defined in proto but not wired", name)
				}
			}
			return true
		},
	)

	// Every procedure in the access policy must belong to a wired service.
	for proc := range app.AccessInternalOnly {
		svc, _, _ := strings.Cut(strings.TrimPrefix(proc, "/"), "/")
		if !registered[svc] {
			t.Errorf("access policy references unwired service %s", svc)
		}
	}
}
