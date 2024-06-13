package handlers

import (
	"context"
	"s3MediaStreamer/app/handlers/REST/audiohandler"
	"s3MediaStreamer/app/handlers/REST/healthhandler"
	"s3MediaStreamer/app/handlers/REST/jobshandler"
	"s3MediaStreamer/app/handlers/REST/otphandler"
	"s3MediaStreamer/app/handlers/REST/playlisthandler"
	"s3MediaStreamer/app/handlers/REST/trackhandler"
	"s3MediaStreamer/app/handlers/REST/userhandler"
	amqp2 "s3MediaStreamer/app/handlers/amqp"
	"s3MediaStreamer/app/internal/app"
)

type Handlers struct {
	Audio    *audiohandler.Handler
	Health   *healthhandler.Handler
	Job      *jobshandler.Handler
	Otp      *otphandler.Handler
	Playlist *playlisthandler.Handler
	Track    *trackhandler.Handler
	User     *userhandler.Handler
	Messages *amqp2.Handler
}

func NewHandlers(ctx context.Context, app *app.App) *Handlers {
	healthHandler := healthhandler.NewMonitoringHandler(*app.Service.Health)
	jobHandler := jobshandler.NewJobHandler()
	trackHandler := trackhandler.NewTrackHandler(*app.Service.Track)
	userHandler := userhandler.NewUserHandler(*app.Service.ACL, *app.Service.User, *app.Service.AccessControl)
	playlistHandler := playlisthandler.NewPlaylistHandler(*app.Service.Playlist, *userHandler)
	otpHandler := otphandler.NewOtpHandler(*app.Service.OTP)
	messageRepo, err := amqp2.NewRabbitMQHandlerWrapper(ctx, app.Cfg, app.Logger, app.Service.InitRepo.InitConnect.RabbitCon, *app.Service.Message)
	audioHandler := audiohandler.NewAudioHandler(app.Service.Audio, app.Logger)
	if err != nil {
		return nil
	}
	return &Handlers{
		audioHandler,
		healthHandler,
		jobHandler,
		otpHandler,
		playlistHandler,
		trackHandler,
		userHandler,
		messageRepo,
	}
}
