package rpc_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/cppforlife/bosh-cpi-go/rpc"
	"github.com/cppforlife/bosh-cpi-go/rpc/rpcfakes"
)

var _ = Describe("JSONDispatcher", func() {
	var (
		actionFactory *rpcfakes.FakeActionFactory
		caller        *rpcfakes.FakeCaller
		dispatcher    JSONDispatcher
	)

	BeforeEach(func() {
		actionFactory = &rpcfakes.FakeActionFactory{}

		actionFactory.CreateStub = func(method string, ctx apiv1.CallContext) (interface{}, error) {
			Expect(method).To(Equal("fake-action"))
			Expect(ctx).ToNot(BeNil())
			return nil, nil
		}

		caller = &rpcfakes.FakeCaller{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		dispatcher = NewJSONDispatcher(actionFactory, caller, logger)
	})

	Describe("Dispatch", func() {
		Context("when method is known", func() {
			It("runs action with provided simple arguments", func() {
				dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))

				method, ctx := actionFactory.CreateArgsForCall(0)
				Expect(method).To(Equal("fake-action"))
				Expect(ctx).To(Equal(apiv1.CloudPropsImpl{}))

				_, args := caller.CallArgsForCall(0)
				Expect(args).To(Equal([]interface{}{"fake-arg"}))
			})

			It("runs action with provided more complex arguments", func() {
				dispatcher.Dispatch([]byte(`{
          "method":"fake-action",
          "arguments":[
            123,
            "fake-arg",
            [123, "fake-arg"],
            {"fake-arg2-key":"fake-arg2-value"}
          ]
        }`))

				method, ctx := actionFactory.CreateArgsForCall(0)
				Expect(method).To(Equal("fake-action"))
				Expect(ctx).To(Equal(apiv1.CloudPropsImpl{}))

				_, args := caller.CallArgsForCall(0)
				Expect(args).To(Equal([]interface{}{
					float64(123),
					"fake-arg",
					[]interface{}{float64(123), "fake-arg"},
					map[string]interface{}{"fake-arg2-key": "fake-arg2-value"},
				}))
			})

			It("runs action with provided context", func() {
				dispatcher.Dispatch([]byte(`{
          "method":"fake-action",
          "arguments":[],
          "context":{"ctx1": "ctx1-val"}
        }`))

				type TestCtx struct {
					Ctx1 string
				}

				method, ctx := actionFactory.CreateArgsForCall(0)
				Expect(method).To(Equal("fake-action"))

				var parsedCtx TestCtx
				err := ctx.As(&parsedCtx)
				Expect(err).ToNot(HaveOccurred())
				Expect(parsedCtx).To(Equal(TestCtx{Ctx1: "ctx1-val"}))

				_, args := caller.CallArgsForCall(0)
				Expect(args).To(Equal([]interface{}{}))
			})

			Context("when running action succeeds", func() {
				It("returns serialized result without including error when result can be serialized", func() {
					caller.CallReturns("fake-result", nil)

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": "fake-result",
            "error": null,
            "log": ""
          }`))
				})

				It("returns Bosh::Clouds::CpiError when result cannot be serialized", func() {
					caller.CallReturns(func() {}, nil) // funcs do not serialize

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": null,
            "error": {
              "type":"Bosh::Clouds::CpiError",
              "message":"Failed to serialize result",
              "ok_to_retry": false
            },
            "log": ""
          }`))
				})
			})

			Context("when running action fails", func() {
				It("returns error without result when action error is a CloudError", func() {
					caller.CallReturns(nil, &rpcfakes.FakeCloudError{
						TypeStub:  func() string { return "fake-type" },
						ErrorStub: func() string { return "fake-message" },
					})

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": null,
            "error": {
              "type":"fake-type",
              "message":"fake-message",
              "ok_to_retry": false
            },
            "log": ""
          }`))
				})

				It("returns error with ok_to_retry=true when action error is a RetryableError and it can be retried", func() {
					caller.CallReturns(nil, &rpcfakes.FakeRetryableError{
						ErrorStub:    func() string { return "fake-message" },
						CanRetryStub: func() bool { return true },
					})

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": null,
            "error": {
              "type":"Bosh::Clouds::CloudError",
              "message":"fake-message",
              "ok_to_retry": true
            },
            "log": ""
          }`))
				})

				It("returns error with ok_to_retry=false when action error is a RetryableError and it cannot be retried", func() {
					caller.CallReturns(nil, &rpcfakes.FakeRetryableError{
						ErrorStub:    func() string { return "fake-message" },
						CanRetryStub: func() bool { return false },
					})

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": null,
            "error": {
              "type":"Bosh::Clouds::CloudError",
              "message":"fake-message",
              "ok_to_retry": false
            },
            "log": ""
          }`))
				})

				It("returns error without result when action error is neither CloudError or RetryableError", func() {
					caller.CallReturns(nil, errors.New("fake-run-err"))

					resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":["fake-arg"]}`))
					Expect(resp).To(MatchJSON(`{
            "result": null,
            "error": {
              "type":"Bosh::Clouds::CloudError",
              "message":"fake-run-err",
              "ok_to_retry": false
            },
            "log": ""
          }`))
				})
			})
		})

		Context("when method is unknown", func() {
			It("responds with Bosh::Clouds::NotImplemented error", func() {
				actionFactory.CreateReturns(nil, errors.New("fake-err"))

				resp := dispatcher.Dispatch([]byte(`{"method":"fake-action","arguments":[]}`))
				Expect(resp).To(MatchJSON(`{
          "result": null,
          "error": {
            "type":"Bosh::Clouds::NotImplemented",
            "message":"Must call implemented method",
            "ok_to_retry": false
          },
          "log": ""
        }`))
			})
		})

		Context("when method key is missing", func() {
			It("responds with Bosh::Clouds::CpiError error", func() {
				resp := dispatcher.Dispatch([]byte(`{}`))
				Expect(resp).To(MatchJSON(`{
          "result": null,
          "error": {
            "type":"Bosh::Clouds::CpiError",
            "message":"Must provide 'method' key",
            "ok_to_retry": false
          },
          "log": ""
        }`))
			})
		})

		Context("when arguments key is missing", func() {
			It("responds with Bosh::Clouds::CpiError error", func() {
				resp := dispatcher.Dispatch([]byte(`{"method":"fake-action"}`))
				Expect(resp).To(MatchJSON(`{
          "result": null,
          "error": {
            "type":"Bosh::Clouds::CpiError",
            "message":"Must provide 'arguments' key",
            "ok_to_retry": false
          },
          "log": ""
        }`))
			})
		})

		Context("when payload cannot be deserialized", func() {
			It("responds with Bosh::Clouds::CpiError error", func() {
				resp := dispatcher.Dispatch([]byte(`{-}`))
				Expect(resp).To(MatchJSON(`{
          "result": null,
          "error": {
            "type":"Bosh::Clouds::CpiError",
            "message":"Must provide valid JSON payload",
            "ok_to_retry": false
          },
          "log": ""
        }`))
			})
		})
	})
})
