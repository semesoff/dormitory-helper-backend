package kitchenServer

import (
	"context"
	kitchenProto "dormitory-helper-service/generated/proto/kitchen"
	laundryRepository "dormitory-helper-service/internal/repository/laundry"
	grpcUtils "dormitory-helper-service/internal/utils/grpc"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KitchenService interface {
	CreateKitchenBooking(ctx context.Context, userID int, startTime, endTime time.Time) (int, error)
	GetKitchenBookings(ctx context.Context, startTime, endTime *time.Time) ([]laundryRepository.Booking, error)
	GetUserKitchenBookings(ctx context.Context, userID int) ([]laundryRepository.Booking, error)
	DeleteKitchenBooking(ctx context.Context, bookingID, userID int) error
}

// Server implementation
type Server struct {
	kitchenProto.UnimplementedKitchenServiceServer
	service   KitchenService
	jwtSecret []byte
}

func NewServer(service KitchenService, jwtSecret []byte) *Server {
	return &Server{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

func (s *Server) CreateKitchenBooking(ctx context.Context, req *kitchenProto.CreateKitchenBookingRequest) (*kitchenProto.CreateKitchenBookingResponse, error) {
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

	bookingID, err := s.service.CreateKitchenBooking(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create kitchen booking: %v", err)
	}

	return &kitchenProto.CreateKitchenBookingResponse{
		BookingId: int32(bookingID),
		Message:   "Kitchen booking created successfully",
	}, nil
}

func (s *Server) GetKitchenBookings(ctx context.Context, req *kitchenProto.GetKitchenBookingsRequest) (*kitchenProto.GetKitchenBookingsResponse, error) {
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	bookings, err := s.service.GetKitchenBookings(ctx, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get kitchen bookings: %v", err)
	}

	response := &kitchenProto.GetKitchenBookingsResponse{
		Bookings: make([]*kitchenProto.KitchenBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &kitchenProto.KitchenBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) GetUserKitchenBookings(ctx context.Context, req *kitchenProto.GetUserKitchenBookingsRequest) (*kitchenProto.GetUserKitchenBookingsResponse, error) {
	// Валидация JWT и получение user_id
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	bookings, err := s.service.GetUserKitchenBookings(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user kitchen bookings: %v", err)
	}

	response := &kitchenProto.GetUserKitchenBookingsResponse{
		Bookings: make([]*kitchenProto.KitchenBooking, len(bookings)),
	}

	for i, b := range bookings {
		response.Bookings[i] = &kitchenProto.KitchenBooking{
			Id:        int32(b.ID),
			UserId:    int32(b.UserID),
			StartTime: timestamppb.New(b.StartTime),
			EndTime:   timestamppb.New(b.EndTime),
		}
	}

	return response, nil
}

func (s *Server) DeleteKitchenBooking(ctx context.Context, req *kitchenProto.DeleteKitchenBookingRequest) (*kitchenProto.DeleteKitchenBookingResponse, error) {
	// Валидация JWT и получение user_id
	userID, err := grpcUtils.ValidateTokenAndGetUserID(req.Token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	err = s.service.DeleteKitchenBooking(ctx, int(req.BookingId), userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete kitchen booking: %v", err)
	}

	return &kitchenProto.DeleteKitchenBookingResponse{
		Message: "Kitchen booking deleted successfully",
	}, nil
}
