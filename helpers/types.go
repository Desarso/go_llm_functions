package llm

import "google.golang.org/protobuf/runtime/protoimpl"

type Request struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Stream     bool      `json:"stream"`
	ToolChoice string    `json:"tool_choice,omitempty"`
	Tools      []Tool    `json:"tools,omitempty"` // Added field for tools
}

type Tool struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`               // The name of the tool
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`               // The type of the tool (e.g., "function")
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"` // A description of what the tool does
	Function      *Function              `protobuf:"bytes,4,opt,name=function,proto3" json:"function,omitempty"`       // A nested struct for the function details
	unknownFields protoimpl.UnknownFields
}

type Message struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Role          string                 `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Content       string                 `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	ToolCalls     []*ToolCall            `protobuf:"bytes,4,rep,name=tool_calls,json=toolCalls,proto3" json:"tool_calls,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type ToolCall struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`             // ID for the tool call
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`         // Type of the tool (e.g., "function")
	Function      *Function              `protobuf:"bytes,3,opt,name=function,proto3" json:"function,omitempty"` // Function being called with its name and parameters
	Index         int32                  `protobuf:"varint,4,opt,name=index,proto3" json:"index,omitempty"`      // Index of the tool call
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Function struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"` // Name of the function
	Arguments     string                 `protobuf:"bytes,2,opt,name=arguments,proto3" json:"arguments,omitempty"`
	Parameters    *Parameters            `protobuf:"bytes,3,opt,name=parameters,proto3" json:"parameters,omitempty"` // Parameters of the function
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Parameters struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          string                 `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`                                                                                       // Type of the parameters (e.g., "object")
	Properties    map[string]*Field      `protobuf:"bytes,2,rep,name=properties,proto3" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // The properties of the parameters (a map of field names and types)
	Required      []string               `protobuf:"bytes,3,rep,name=required,proto3" json:"required,omitempty"`                                                                               // List of required fields
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Field struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Description   string                 `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"` // A description of the parameter
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`               // The type of the parameter (e.g., "string", "number")
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type ResponseData struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Id                string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Object            string                 `protobuf:"bytes,2,opt,name=object,proto3" json:"object,omitempty"`
	Created           int64                  `protobuf:"varint,3,opt,name=created,proto3" json:"created,omitempty"`
	Model             string                 `protobuf:"bytes,4,opt,name=model,proto3" json:"model,omitempty"`
	SystemFingerprint string                 `protobuf:"bytes,5,opt,name=system_fingerprint,json=systemFingerprint,proto3" json:"system_fingerprint,omitempty"`
	Choices           []*Choice              `protobuf:"bytes,6,rep,name=choices,proto3" json:"choices,omitempty"`
	XGroq             *XGroq                 `protobuf:"bytes,7,opt,name=x_groq,json=xGroq,proto3" json:"x_groq,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

type XGroq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Usage         *Usage                 `protobuf:"bytes,2,opt,name=usage,proto3" json:"usage,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Usage struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	QueueTime        float32                `protobuf:"fixed32,1,opt,name=queue_time,json=queueTime,proto3" json:"queue_time,omitempty"`
	PromptTokens     int32                  `protobuf:"varint,2,opt,name=prompt_tokens,json=promptTokens,proto3" json:"prompt_tokens,omitempty"`
	PromptTime       float32                `protobuf:"fixed32,3,opt,name=prompt_time,json=promptTime,proto3" json:"prompt_time,omitempty"`
	CompletionTokens int32                  `protobuf:"varint,4,opt,name=completion_tokens,json=completionTokens,proto3" json:"completion_tokens,omitempty"`
	CompletionTime   float32                `protobuf:"fixed32,5,opt,name=completion_time,json=completionTime,proto3" json:"completion_time,omitempty"`
	TotalTokens      int32                  `protobuf:"varint,6,opt,name=total_tokens,json=totalTokens,proto3" json:"total_tokens,omitempty"`
	TotalTime        float32                `protobuf:"fixed32,7,opt,name=total_time,json=totalTime,proto3" json:"total_time,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

type Choice struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Index         int32                  `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Delta         *Delta                 `protobuf:"bytes,2,opt,name=delta,proto3" json:"delta,omitempty"` // To represent an expandable data type
	FinishReason  string                 `protobuf:"bytes,4,opt,name=finish_reason,json=finishReason,proto3" json:"finish_reason,omitempty"`
	Content       string                 `protobuf:"bytes,5,opt,name=content,proto3" json:"content,omitempty"`
	Message       *Message               `protobuf:"bytes,6,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Delta struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Content       string                 `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`                      // Content as a string, for text updates
	ToolCalls     []*ToolCall            `protobuf:"bytes,2,rep,name=tool_calls,json=toolCalls,proto3" json:"tool_calls,omitempty"` // List of tool calls in the delta.
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}
