package laundryServer

import (
	"context"
	laundryRepository "dormitory-helper-service/internal/repository/laundry"
	grpcUtils "dormitory-helper-service/internal/utils/grpc"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Placeholder types until proto files are generated
// These will be replaced by actual generated types after running make proto

type LaundryService interface {
	CreateLaundryBooking(ctx context.Context, userID int, startTime, endTime time.Time) (int, error)
	GetLaundryBookings(ctx context.Context, startTime, endTime *time.Time) ([]laundryRepository.Booking, error)
	GetUserLaundryBookings(ctx context.Context, userID int) ([]laundryRepository.Booking, error)
	DeleteLaundryBooking(ctx context.Context, bookingID, userID int) error
}

// Placeholder proto message types
type CreateLaundryBookingRequest struct {
	Token     string
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
}

type CreateLaundryBookingResponse struct {
	BookingId int32
	Message   string
}

type GetLaundryBookingsRequest struct {
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
}

type LaundryBooking struct {
	Id        int32
	UserId    int32
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
}

type GetLaundryBookingsResponse struct {
	Bookings []*LaundryBooking
}

type GetUserLaundryBookingsRequest struct {
	Token string
}

type GetUserLaundryBookingsResponse struct {
	Bookings []*LaundryBooking
}

type DeleteLaundryBookingRequest struct {
	Token     string
	BookingId int32
}

type DeleteLaundryBookingResponse struct {
	Message string
}

type UnimplementedLaundryServiceServer struct{}

func (UnimplementedLaundryServiceServer) CreateLaundryBooking(context.Context, *CreateLaundryBookingRequest) (*CreateLaundryBookingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateLaundryBooking not implemented")
}

func (UnimplementedLaundryServiceServer) GetLaundryBookings(context.Context, *GetLaundryBookingsRequest) (*GetLaundryBookingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLaundryBookings not implemented")
}

func (UnimplementedLaundryServiceServer) GetUserLaundryBookings(context.Context, *GetUserLaundryBookingsRequest) (*GetUserLaundryBookingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserLaundryBookings not implemented")
}

func (UnimplementedLaundryServiceServer) DeleteLaundryBooking(context.Context, *DeleteLaundryBookingRequest) (*DeleteLaundryBookingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLaundryBooking not implemented")
}

// Server implementation
type Server struct {
	UnimplementedLaundryServiceServer
	service   LaundryService
	jwtSecret []byte
}

func NewServer(service LaundryService, jwtSecret []byte) *Server {
	return &Server{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

func (s *Server) CreateLaundryBooking(ctx context.Context, req *CreateLaundryBookingRequest) (*CreateLaundryBookingResponse, error) {
	// Валидация JWT и получение user_id
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if req.StartTime == nil || req.EndTime == nil {
		return nil, status.Errorf(codes.InvalidArgument, "start_time and end_time are required")
	}

	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	bookingID, err := s.service.CreateLaundryBooking(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create laundry booking: %v", err)
	}

	return &CreateLaundryBookingResponse{
		BookingId: int32(bookingID),
		Message:   "Laundry booking created successfully",
	}, nil
}

func (s *Server) GetLaundryBookings(ctx context.Context, req *GetLaundryBookingsRequest) (*GetLaundryBookingsResponse, error) {
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	bookings, err := s.service.GetLaundryBookings(ctx, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get laundry bookings: %v", err)
	}

	response := &GetLaundryBookingsResponse{
		Bookings: make([]*LaundryBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &LaundryBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) GetUserLaundryBookings(ctx context.Context, req *GetUserLaundryBookingsRequest) (*GetUserLaundryBookingsResponse, error) {
	// Валидация JWT и получение user_id
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	bookings, err := s.service.GetUserLaundryBookings(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user laundry bookings: %v", err)
	}

	response := &GetUserLaundryBookingsResponse{
		Bookings: make([]*LaundryBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &LaundryBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) DeleteLaundryBooking(ctx context.Context, req *DeleteLaundryBookingRequest) (*DeleteLaundryBookingResponse, error) {
	// Валидация JWT и получение user_id
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	err = s.service.DeleteLaundryBooking(ctx, int(req.BookingId), userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete laundry booking: %v", err)
	}

	return &DeleteLaundryBookingResponse{
		Message: "Laundry booking deleted successfully",
	}, nil
}
