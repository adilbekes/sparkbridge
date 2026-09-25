package pb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type EdgeMessage struct {
	state         protoimpl.MessageState
	ClientId      string    `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	DeviceId      string    `protobuf:"bytes,2,opt,name=device_id,json=deviceId,proto3" json:"device_id,omitempty"`
	Metrics       []*Metric `protobuf:"bytes,3,rep,name=metrics,proto3" json:"metrics,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Metric struct {
	state         protoimpl.MessageState
	Name          string         `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Timestamp     int64          `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Value         isMetric_Value `protobuf_oneof:"value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type isMetric_Value interface{ isMetric_Value() }

type Metric_IntValue struct {
	IntValue int64 `protobuf:"varint,3,opt,name=int_value,json=intValue,proto3,oneof" json:"int_value,omitempty"`
}
type Metric_DoubleValue struct {
	DoubleValue float64 `protobuf:"fixed64,4,opt,name=double_value,json=doubleValue,proto3,oneof" json:"double_value,omitempty"`
}
type Metric_StringValue struct {
	StringValue string `protobuf:"bytes,5,opt,name=string_value,json=stringValue,proto3,oneof" json:"string_value,omitempty"`
}
type Metric_BoolValue struct {
	BoolValue bool `protobuf:"varint,6,opt,name=bool_value,json=boolValue,proto3,oneof" json:"bool_value,omitempty"`
}

func (*Metric_IntValue) isMetric_Value()    {}
func (*Metric_DoubleValue) isMetric_Value() {}
func (*Metric_StringValue) isMetric_Value() {}
func (*Metric_BoolValue) isMetric_Value()   {}

type CloudCommand struct {
	state         protoimpl.MessageState
	ClientId      string            `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	CommandId     string            `protobuf:"bytes,2,opt,name=command_id,json=commandId,proto3" json:"command_id,omitempty"`
	TargetDevice  string            `protobuf:"bytes,3,opt,name=target_device,json=targetDevice,proto3" json:"target_device,omitempty"`
	Action        string            `protobuf:"bytes,4,opt,name=action,proto3" json:"action,omitempty"`
	Parameters    map[string]string `protobuf:"bytes,5,rep,name=parameters,proto3" json:"parameters,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type SparkplugIngressServiceClient interface {
	Bridge(ctx context.Context, opts ...grpc.CallOption) (SparkplugIngressService_BridgeClient, error)
}

type sparkplugIngressServiceClient struct{ cc grpc.ClientConnInterface }

func NewSparkplugIngressServiceClient(cc grpc.ClientConnInterface) SparkplugIngressServiceClient {
	return &sparkplugIngressServiceClient{cc}
}

func (c *sparkplugIngressServiceClient) Bridge(ctx context.Context, opts ...grpc.CallOption) (SparkplugIngressService_BridgeClient, error) {
	stream, err := c.cc.NewStream(ctx, &SparkplugIngressService_ServiceDesc.Streams[0], "/sparkplug.v1.SparkplugIngressService/Bridge", opts...)
	if err != nil {
		return nil, err
	}
	return &sparkplugIngressServiceBridgeClient{stream}, nil
}

type SparkplugIngressService_BridgeClient interface {
	Send(*EdgeMessage) error
	Recv() (*CloudCommand, error)
	grpc.ClientStream
}

type sparkplugIngressServiceBridgeClient struct{ grpc.ClientStream }

func (x *sparkplugIngressServiceBridgeClient) Send(m *EdgeMessage) error {
	return x.ClientStream.SendMsg(m)
}
func (x *sparkplugIngressServiceBridgeClient) Recv() (*CloudCommand, error) {
	m := new(CloudCommand)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type SparkplugIngressServiceServer interface {
	Bridge(SparkplugIngressService_BridgeServer) error
}

type UnimplementedSparkplugIngressServiceServer struct{}

func (UnimplementedSparkplugIngressServiceServer) Bridge(SparkplugIngressService_BridgeServer) error {
	return status.Errorf(codes.Unimplemented, "method Bridge not implemented")
}

type SparkplugIngressService_BridgeServer interface {
	Send(*CloudCommand) error
	Recv() (*EdgeMessage, error)
	grpc.ServerStream
}

func RegisterSparkplugIngressServiceServer(s grpc.ServiceRegistrar, srv SparkplugIngressServiceServer) {
	s.RegisterService(&SparkplugIngressService_ServiceDesc, srv)
}

func _SparkplugIngressService_Bridge_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SparkplugIngressServiceServer).Bridge(&sparkplugIngressServiceBridgeServer{stream})
}

type sparkplugIngressServiceBridgeServer struct{ grpc.ServerStream }

func (x *sparkplugIngressServiceBridgeServer) Send(m *CloudCommand) error {
	return x.ServerStream.SendMsg(m)
}
func (x *sparkplugIngressServiceBridgeServer) Recv() (*EdgeMessage, error) {
	m := new(EdgeMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var SparkplugIngressService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sparkplug.v1.SparkplugIngressService",
	HandlerType: (*SparkplugIngressServiceServer)(nil),
	Streams: []grpc.StreamDesc{{
		StreamName:    "Bridge",
		Handler:       _SparkplugIngressService_Bridge_Handler,
		ServerStreams: true,
		ClientStreams: true,
	}},
}
