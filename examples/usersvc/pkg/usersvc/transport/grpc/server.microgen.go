// Code generated by microgen. DO NOT EDIT.

package grpc

import (
	opentracing "github.com/go-kit/kit/tracing/opentracing"
	opentracinggo "github.com/opentracing/opentracing-go"
)

func TraceServer(tracer opentracinggo.Tracer) func(endpoints Endpoints) Endpoints {
	return func(endpoints Endpoints) Endpoints {
		return Endpoints{
			CreateComment_Endpoint:   opentracing.TraceServer(tracer, "CreateComment")(endpoints.CreateComment_Endpoint),
			CreateUser_Endpoint:      opentracing.TraceServer(tracer, "CreateUser")(endpoints.CreateUser_Endpoint),
			FindUsers_Endpoint:       opentracing.TraceServer(tracer, "FindUsers")(endpoints.FindUsers_Endpoint),
			GetComment_Endpoint:      opentracing.TraceServer(tracer, "GetComment")(endpoints.GetComment_Endpoint),
			GetUserComments_Endpoint: opentracing.TraceServer(tracer, "GetUserComments")(endpoints.GetUserComments_Endpoint),
			GetUser_Endpoint:         opentracing.TraceServer(tracer, "GetUser")(endpoints.GetUser_Endpoint),
			UpdateUser_Endpoint:      opentracing.TraceServer(tracer, "UpdateUser")(endpoints.UpdateUser_Endpoint),
		}
	}
}