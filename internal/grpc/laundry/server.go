package laundryServer

import (
	"context"
	laundryProto "dormitory-helper-service/generated/proto/laundry"
	laundryRepository "dormitory-helper-service/internal/repository/laundry"
	grpcUtils "dormitory-helper-service/internal/utils/grpc"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LaundryService interface {
	CreateLaundryBooking(ctx context.Context, userID int, startTime, endTime time.Time) (int, error)
	GetLaundryBookings(ctx context.Context, startTime, endTime *time.Time) ([]laundryRepository.Booking, error)
	GetUserLaundryBookings(ctx context.Context, userID int) ([]laundryRepository.Booking, error)
	DeleteLaundryBooking(ctx context.Context, bookingID, userID int) error
}

type Server struct {
	laundryProto.UnimplementedLaundryServiceServer
	service   LaundryService
	jwtSecret []byte
}

func NewServer(service LaundryService, jwtSecret []byte) *Server {
	return &Server{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

func (s *Server) CreateLaundryBooking(ctx context.Context, req *laundryProto.CreateLaundryBookingRequest) (*laundryProto.CreateLaundryBookingResponse, error) {
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

	return &laundryProto.CreateLaundryBookingResponse{
		BookingId: int32(bookingID),
		Message:   "Laundry booking created successfully",
	}, nil
}

func (s *Server) GetLaundryBookings(ctx context.Context, req *laundryProto.GetLaundryBookingsRequest) (*laundryProto.GetLaundryBookingsResponse, error) {
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

	response := &laundryProto.GetLaundryBookingsResponse{
		Bookings: make([]*laundryProto.LaundryBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &laundryProto.LaundryBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) GetUserLaundryBookings(ctx context.Context, req *laundryProto.GetUserLaundryBookingsRequest) (*laundryProto.GetUserLaundryBookingsResponse, error) {
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	bookings, err := s.service.GetUserLaundryBookings(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user laundry bookings: %v", err)
	}

	response := &laundryProto.GetUserLaundryBookingsResponse{
		Bookings: make([]*laundryProto.LaundryBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &laundryProto.LaundryBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) DeleteLaundryBooking(ctx context.Context, req *laundryProto.DeleteLaundryBookingRequest) (*laundryProto.DeleteLaundryBookingResponse, error) {
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	err = s.service.DeleteLaundryBooking(ctx, int(req.BookingId), userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete laundry booking: %v", err)
	}

	return &laundryProto.DeleteLaundryBookingResponse{
		Message: "Laundry booking deleted successfully",
	}, nil
}
