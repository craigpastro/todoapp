// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: todoapp/v1/service.proto

package todoappv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/craigpastro/todoapp/internal/gen/todoapp/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// TodoAppServiceName is the fully-qualified name of the TodoAppService service.
	TodoAppServiceName = "todoapp.v1.TodoAppService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// TodoAppServiceCreateProcedure is the fully-qualified name of the TodoAppService's Create RPC.
	TodoAppServiceCreateProcedure = "/todoapp.v1.TodoAppService/Create"
	// TodoAppServiceReadProcedure is the fully-qualified name of the TodoAppService's Read RPC.
	TodoAppServiceReadProcedure = "/todoapp.v1.TodoAppService/Read"
	// TodoAppServiceReadAllProcedure is the fully-qualified name of the TodoAppService's ReadAll RPC.
	TodoAppServiceReadAllProcedure = "/todoapp.v1.TodoAppService/ReadAll"
	// TodoAppServiceUpdateProcedure is the fully-qualified name of the TodoAppService's Update RPC.
	TodoAppServiceUpdateProcedure = "/todoapp.v1.TodoAppService/Update"
	// TodoAppServiceDeleteProcedure is the fully-qualified name of the TodoAppService's Delete RPC.
	TodoAppServiceDeleteProcedure = "/todoapp.v1.TodoAppService/Delete"
)

// TodoAppServiceClient is a client for the todoapp.v1.TodoAppService service.
type TodoAppServiceClient interface {
	Create(context.Context, *connect_go.Request[v1.CreateRequest]) (*connect_go.Response[v1.CreateResponse], error)
	Read(context.Context, *connect_go.Request[v1.ReadRequest]) (*connect_go.Response[v1.ReadResponse], error)
	ReadAll(context.Context, *connect_go.Request[v1.ReadAllRequest]) (*connect_go.Response[v1.ReadAllResponse], error)
	Update(context.Context, *connect_go.Request[v1.UpdateRequest]) (*connect_go.Response[v1.UpdateResponse], error)
	Delete(context.Context, *connect_go.Request[v1.DeleteRequest]) (*connect_go.Response[v1.DeleteResponse], error)
}

// NewTodoAppServiceClient constructs a client for the todoapp.v1.TodoAppService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewTodoAppServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) TodoAppServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &todoAppServiceClient{
		create: connect_go.NewClient[v1.CreateRequest, v1.CreateResponse](
			httpClient,
			baseURL+TodoAppServiceCreateProcedure,
			opts...,
		),
		read: connect_go.NewClient[v1.ReadRequest, v1.ReadResponse](
			httpClient,
			baseURL+TodoAppServiceReadProcedure,
			opts...,
		),
		readAll: connect_go.NewClient[v1.ReadAllRequest, v1.ReadAllResponse](
			httpClient,
			baseURL+TodoAppServiceReadAllProcedure,
			opts...,
		),
		update: connect_go.NewClient[v1.UpdateRequest, v1.UpdateResponse](
			httpClient,
			baseURL+TodoAppServiceUpdateProcedure,
			opts...,
		),
		delete: connect_go.NewClient[v1.DeleteRequest, v1.DeleteResponse](
			httpClient,
			baseURL+TodoAppServiceDeleteProcedure,
			opts...,
		),
	}
}

// todoAppServiceClient implements TodoAppServiceClient.
type todoAppServiceClient struct {
	create  *connect_go.Client[v1.CreateRequest, v1.CreateResponse]
	read    *connect_go.Client[v1.ReadRequest, v1.ReadResponse]
	readAll *connect_go.Client[v1.ReadAllRequest, v1.ReadAllResponse]
	update  *connect_go.Client[v1.UpdateRequest, v1.UpdateResponse]
	delete  *connect_go.Client[v1.DeleteRequest, v1.DeleteResponse]
}

// Create calls todoapp.v1.TodoAppService.Create.
func (c *todoAppServiceClient) Create(ctx context.Context, req *connect_go.Request[v1.CreateRequest]) (*connect_go.Response[v1.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Read calls todoapp.v1.TodoAppService.Read.
func (c *todoAppServiceClient) Read(ctx context.Context, req *connect_go.Request[v1.ReadRequest]) (*connect_go.Response[v1.ReadResponse], error) {
	return c.read.CallUnary(ctx, req)
}

// ReadAll calls todoapp.v1.TodoAppService.ReadAll.
func (c *todoAppServiceClient) ReadAll(ctx context.Context, req *connect_go.Request[v1.ReadAllRequest]) (*connect_go.Response[v1.ReadAllResponse], error) {
	return c.readAll.CallUnary(ctx, req)
}

// Update calls todoapp.v1.TodoAppService.Update.
func (c *todoAppServiceClient) Update(ctx context.Context, req *connect_go.Request[v1.UpdateRequest]) (*connect_go.Response[v1.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Delete calls todoapp.v1.TodoAppService.Delete.
func (c *todoAppServiceClient) Delete(ctx context.Context, req *connect_go.Request[v1.DeleteRequest]) (*connect_go.Response[v1.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// TodoAppServiceHandler is an implementation of the todoapp.v1.TodoAppService service.
type TodoAppServiceHandler interface {
	Create(context.Context, *connect_go.Request[v1.CreateRequest]) (*connect_go.Response[v1.CreateResponse], error)
	Read(context.Context, *connect_go.Request[v1.ReadRequest]) (*connect_go.Response[v1.ReadResponse], error)
	ReadAll(context.Context, *connect_go.Request[v1.ReadAllRequest]) (*connect_go.Response[v1.ReadAllResponse], error)
	Update(context.Context, *connect_go.Request[v1.UpdateRequest]) (*connect_go.Response[v1.UpdateResponse], error)
	Delete(context.Context, *connect_go.Request[v1.DeleteRequest]) (*connect_go.Response[v1.DeleteResponse], error)
}

// NewTodoAppServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewTodoAppServiceHandler(svc TodoAppServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	todoAppServiceCreateHandler := connect_go.NewUnaryHandler(
		TodoAppServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	todoAppServiceReadHandler := connect_go.NewUnaryHandler(
		TodoAppServiceReadProcedure,
		svc.Read,
		opts...,
	)
	todoAppServiceReadAllHandler := connect_go.NewUnaryHandler(
		TodoAppServiceReadAllProcedure,
		svc.ReadAll,
		opts...,
	)
	todoAppServiceUpdateHandler := connect_go.NewUnaryHandler(
		TodoAppServiceUpdateProcedure,
		svc.Update,
		opts...,
	)
	todoAppServiceDeleteHandler := connect_go.NewUnaryHandler(
		TodoAppServiceDeleteProcedure,
		svc.Delete,
		opts...,
	)
	return "/todoapp.v1.TodoAppService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case TodoAppServiceCreateProcedure:
			todoAppServiceCreateHandler.ServeHTTP(w, r)
		case TodoAppServiceReadProcedure:
			todoAppServiceReadHandler.ServeHTTP(w, r)
		case TodoAppServiceReadAllProcedure:
			todoAppServiceReadAllHandler.ServeHTTP(w, r)
		case TodoAppServiceUpdateProcedure:
			todoAppServiceUpdateHandler.ServeHTTP(w, r)
		case TodoAppServiceDeleteProcedure:
			todoAppServiceDeleteHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedTodoAppServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedTodoAppServiceHandler struct{}

func (UnimplementedTodoAppServiceHandler) Create(context.Context, *connect_go.Request[v1.CreateRequest]) (*connect_go.Response[v1.CreateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("todoapp.v1.TodoAppService.Create is not implemented"))
}

func (UnimplementedTodoAppServiceHandler) Read(context.Context, *connect_go.Request[v1.ReadRequest]) (*connect_go.Response[v1.ReadResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("todoapp.v1.TodoAppService.Read is not implemented"))
}

func (UnimplementedTodoAppServiceHandler) ReadAll(context.Context, *connect_go.Request[v1.ReadAllRequest]) (*connect_go.Response[v1.ReadAllResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("todoapp.v1.TodoAppService.ReadAll is not implemented"))
}

func (UnimplementedTodoAppServiceHandler) Update(context.Context, *connect_go.Request[v1.UpdateRequest]) (*connect_go.Response[v1.UpdateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("todoapp.v1.TodoAppService.Update is not implemented"))
}

func (UnimplementedTodoAppServiceHandler) Delete(context.Context, *connect_go.Request[v1.DeleteRequest]) (*connect_go.Response[v1.DeleteResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("todoapp.v1.TodoAppService.Delete is not implemented"))
}
