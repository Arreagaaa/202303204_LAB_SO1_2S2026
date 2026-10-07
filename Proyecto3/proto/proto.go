package proto

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WarReportRequest struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

type WarReportResponse struct {
	Status string `json:"status"`
}

type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(value interface{}) ([]byte, error) { return json.Marshal(value) }
func (jsonCodec) Unmarshal(data []byte, value interface{}) error { return json.Unmarshal(data, value) }

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

type WarReportServiceClient interface {
	SendReport(context.Context, *WarReportRequest, ...grpc.CallOption) (*WarReportResponse, error)
}

type warReportServiceClient struct {
	connection *grpc.ClientConn
}

func NewWarReportServiceClient(connection *grpc.ClientConn) WarReportServiceClient {
	return &warReportServiceClient{connection: connection}
}

func (c *warReportServiceClient) SendReport(ctx context.Context, request *WarReportRequest, options ...grpc.CallOption) (*WarReportResponse, error) {
	response := new(WarReportResponse)
	err := c.connection.Invoke(ctx, "/wartweets.WarReportService/SendReport", request, response, append(options, grpc.ForceCodec(jsonCodec{}))...)
	return response, err
}

type WarReportServiceServer interface {
	SendReport(context.Context, *WarReportRequest) (*WarReportResponse, error)
}

type UnimplementedWarReportServiceServer struct{}

func (UnimplementedWarReportServiceServer) SendReport(context.Context, *WarReportRequest) (*WarReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "metodo no implementado")
}

func RegisterWarReportServiceServer(server grpc.ServiceRegistrar, implementation WarReportServiceServer) {
	server.RegisterService(&WarReportService_ServiceDesc, implementation)
}

func _WarReportService_SendReport_Handler(server interface{}, ctx context.Context, decoder func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	request := new(WarReportRequest)
	if err := decoder(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(WarReportServiceServer).SendReport(ctx, request)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/wartweets.WarReportService/SendReport"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(WarReportServiceServer).SendReport(ctx, request.(*WarReportRequest))
	}
	return interceptor(ctx, request, info, handler)
}

var WarReportService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wartweets.WarReportService",
	HandlerType: (*WarReportServiceServer)(nil),
	Methods: []grpc.MethodDesc{{MethodName: "SendReport", Handler: _WarReportService_SendReport_Handler}},
	Streams: []grpc.StreamDesc{},
	Metadata: "report.proto",
}
