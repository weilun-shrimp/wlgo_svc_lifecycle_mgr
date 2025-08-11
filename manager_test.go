package wlgo_svc_lifecycle_mgr

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestAddService(t *testing.T) {

	usedErr := errors.New("error")

	providers := []*serviceProvider{
		{
			name: "provider 1 - no error",
			begin: func() error {
				return nil
			},
		},
		{
			name: "provider 2 - error",
			begin: func() error {
				return usedErr
			},
		},
		{
			name: "provider 3 - no error",
			begin: func() error {
				return nil
			},
		},
	}

	t.Run("one provider", func(t *testing.T) {
		m := NewManager()
		m.AddService(providers[0])
		if len(m.services) < 1 {
			t.Errorf("Expected %d service, got %d", 1, len(m.services))
			return
		}
		if m.services[0] != providers[0] {
			t.Errorf("Expected %+v service, got %+v", providers[0], m.services[0])
			return
		}
	})

	t.Run("multiple provider", func(t *testing.T) {
		m := NewManager()
		m.AddService(
			providers[0],
			providers[1],
			providers[2],
		)
		if len(m.services) != len(providers) {
			t.Errorf("Expected %d service, got %d", len(providers), len(m.services))
			return
		}
		for i := 0; i <= len(m.services)-1; i++ {
			if m.services[i] != providers[i] {
				t.Errorf("Expected %+v service, got %+v, index %d", providers[i], m.services[i], i)
				return
			}
		}
	})
}

func TestStart(t *testing.T) {

	usedErr := errors.New("error")

	providerMap := map[string]*serviceProvider{
		"provider_1_no_error": {
			name: "provider 1 - no error",
			begin: func() error {
				return nil
			},
		},
		"provider_2_error": {
			name: "provider 2 - error",
			begin: func() error {
				return usedErr
			},
		},
		"provider_3_no_error": {
			name: "provider 3 - no error",
			begin: func() error {
				return nil
			},
		},
	}

	scenarios := []struct {
		name                      string
		providers                 []ServiceProvider
		expect_result             Result
		expect_started_result_len int
	}{
		{
			name:      "no provider",
			providers: []ServiceProvider{},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_started_result_len: 0,
		},
		{
			name: "one provider - no error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
			},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_started_result_len: 1,
		},
		{
			name: "one provider - error",
			providers: []ServiceProvider{
				providerMap["provider_2_error"],
			},
			expect_result: &result{
				err:                usedErr,
				errServiceProvider: providerMap["provider_2_error"],
			},
			expect_started_result_len: 0,
		},
		{
			name: "multiple provider - no error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
				providerMap["provider_3_no_error"],
			},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_started_result_len: 2,
		},
		{
			name: "multiple provider - error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
				providerMap["provider_2_error"],
				providerMap["provider_3_no_error"],
			},
			expect_result: &result{
				err:                usedErr,
				errServiceProvider: providerMap["provider_2_error"],
			},
			expect_started_result_len: 1,
		},
	}
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {

			m := NewManager()
			m.AddService(s.providers...)
			result := m.Start()
			if len(m.startedServices) != s.expect_started_result_len {
				t.Errorf("Expected %d started service, got %d", s.expect_started_result_len, len(m.startedServices))
				return
			}

			if !reflect.DeepEqual(result, s.expect_result) {
				t.Errorf("Expected %+v result \n"+
					"	(name: %s, error: %v) \n"+
					"got %+v \n"+
					"	(name: %s, error: %v)",
					s.expect_result,
					s.expect_result.GetErrServiceProvider().GetName(),
					s.expect_result.GetError(),
					result,
					result.GetErrServiceProvider().GetName(),
					result.GetError(),
				)
				return
			}
		})
	}
}

func TestRollback(t *testing.T) {

	type usedCtx struct {
		name   string
		ctx    context.Context
		cancel context.CancelFunc
	}
	gen_usedCtx := func(name string) *usedCtx {
		ctx, cancel := context.WithCancel(context.Background())
		return &usedCtx{
			name:   name,
			ctx:    ctx,
			cancel: cancel,
		}
	}

	usedCtxs := map[string]*usedCtx{
		"provider_1_ctx": nil,
		"provider_2_ctx": nil,
		"provider_3_ctx": nil,
		"provider_4_ctx": nil,
	}

	flush_usedCtxs := func() {
		for _, k := range []string{
			"provider_1_ctx",
			"provider_2_ctx",
			"provider_3_ctx",
			"provider_4_ctx",
		} {
			usedCtxs[k] = gen_usedCtx(k)
		}
	}

	usedErr := errors.New("error")

	providerMap := map[string]*serviceProvider{
		"provider_1_no_error": {
			name: "provider 1 - no error",
			begin: func() error {
				return nil
			},
			end: func() error {
				usedCtxs["provider_1_ctx"].cancel()
				return nil
			},
		},
		"provider_2_no_error": {
			name: "provider 2 - no error",
			begin: func() error {
				return nil
			},
			end: func() error {
				usedCtxs["provider_2_ctx"].cancel()
				return nil
			},
		},
		"provider_3_error": {
			name: "provider 3 - error",
			begin: func() error {
				return nil
			},
			end: func() error {
				usedCtxs["provider_3_ctx"].cancel()
				return usedErr
			},
		},
		"provider_4_no_error": {
			name: "provider 4 - no error",
			begin: func() error {
				return nil
			},
			end: func() error {
				usedCtxs["provider_4_ctx"].cancel()
				return nil
			},
		},
	}

	scenarios := []struct {
		name                    string
		providers               []ServiceProvider
		expect_result           Result
		expect_done_usedCtx_key []string
	}{
		{
			name:      "no provider",
			providers: []ServiceProvider{},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_done_usedCtx_key: []string{},
		},
		{
			name: "one provider - no error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
			},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_done_usedCtx_key: []string{
				"provider_1_ctx",
			},
		},
		{
			name: "one provider - error",
			providers: []ServiceProvider{
				providerMap["provider_3_error"],
			},
			expect_result: &result{
				err:                usedErr,
				errServiceProvider: providerMap["provider_3_error"],
			},
			expect_done_usedCtx_key: []string{},
		},
		{
			name: "multiple provider - no error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
				providerMap["provider_2_no_error"],
				providerMap["provider_4_no_error"],
			},
			expect_result: &result{
				err:                nil,
				errServiceProvider: nil,
			},
			expect_done_usedCtx_key: []string{
				"provider_1_ctx",
				"provider_2_ctx",
				"provider_4_ctx",
			},
		},
		{
			name: "multiple provider - error",
			providers: []ServiceProvider{
				providerMap["provider_1_no_error"],
				providerMap["provider_2_no_error"],
				providerMap["provider_3_error"],
				providerMap["provider_4_no_error"],
			},
			expect_result: &result{
				err:                usedErr,
				errServiceProvider: providerMap["provider_3_error"],
			},
			expect_done_usedCtx_key: []string{
				"provider_3_ctx",
				"provider_4_ctx",
			},
		},
	}
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			flush_usedCtxs()
			m := NewManager()
			m.AddService(s.providers...)
			m.Start()
			result := m.Rollback()

			if !reflect.DeepEqual(result, s.expect_result) {
				t.Errorf("Expected %+v result \n"+
					"	(name: %s, error: %v) \n"+
					"got %+v \n"+
					"	(name: %s, error: %v)",
					s.expect_result,
					s.expect_result.GetErrServiceProvider().GetName(),
					s.expect_result.GetError(),
					result,
					result.GetErrServiceProvider().GetName(),
					result.GetError(),
				)
				return
			}

			for _, uck := range s.expect_done_usedCtx_key {
				select {
				case <-usedCtxs[uck].ctx.Done():
					continue
				default:
					t.Errorf("Expected %+v done ctx but no | key %s", usedCtxs[uck], uck)
					return
				}
			}
		})
	}
}
