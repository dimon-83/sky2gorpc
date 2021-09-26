// Package interceptors
// @Description:
//     GRPC Client/Server tracing  interceptors , integrated with sky2go
package interceptors

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	agentV3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"time"
)

//reference： https://raw.githubusercontent.com/apache/skywalking/1730f2c84bbd4da999ec2c74d1c26db31d5a0d24/oap-server/server-starter/src/main/resources/component-libraries.yml
const componentIDGOHttpServer = 5004
const componentIDGOHttpClient = 5005

type Handler struct {
	tracer *go2sky.Tracer
}

func NewTracerHandler(re go2sky.Reporter, svcName string) (*Handler, error) {
	tracer, err := go2sky.NewTracer(svcName, go2sky.WithReporter(re))
	if err != nil {
		return nil, err
	}
	return &Handler{
		tracer: tracer,
	}, nil
}

func (h Handler) GetTracker() *go2sky.Tracer {
	return h.tracer
}

func (h Handler) RPCClientTracingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	span, err := h.tracer.CreateExitSpan(ctx, method, cc.Target(), func(headerKey, headerValue string) error {
		ctx = metadata.AppendToOutgoingContext(ctx, headerKey, headerValue)
		return nil
	})
	// Pre-processor phase
	log.Printf("======= [Client Interceptor] Method :  %s\n", method)
	if err != nil {
		return err
	}
	span.SetComponent(componentIDGOHttpClient)
	span.SetSpanLayer(agentV3.SpanLayer_RPCFramework)

	// Invoking the remote method
	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}

	// Post-processor phase
	log.Printf("======= [Client Interceptor] Post Proc Reply : %s \n", reply)
	defer span.End()
	return err
}

func (h Handler) RPCServerTracingInterceptor(tracer *go2sky.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Pre-processing logic
		// Gets info about the current RPC call by examining the args passed in

		log.Printf("======= [Server Interceptor] %s\n", info.FullMethod)
		log.Printf(" ======= [Server Interceptor] Pre Proc Message :  %s\n", req)
		span, ctx, err := tracer.CreateEntrySpan(context.Background(), info.FullMethod, func(headerKey string) (string, error) {
			md, b := metadata.FromIncomingContext(ctx)
			if b {
				return md.Get(headerKey)[0], nil
			}
			return "", nil
		})
		if err != nil {
			return nil, err
		}
		span.SetComponent(componentIDGOHttpServer)
		span.SetSpanLayer(agentV3.SpanLayer_RPCFramework)

		m, err := handler(ctx, req)

		defer func() {
			if err != nil {
				span.Error(time.Now(), err.Error())
			}
			span.End()
		}()

		// Post processing logic
		log.Printf(" ======= [Server Interceptor] Post Proc Message : %s\n", m)
		return m, err
	}
}
