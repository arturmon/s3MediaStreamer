package handlers

import (
	"context"
	"s3MediaStreamer/app/handlers/REST/audio_handler"
	"s3MediaStreamer/app/handlers/REST/health_handler"
	"s3MediaStreamer/app/handlers/REST/jobs_handler"
	"s3MediaStreamer/app/handlers/REST/otp_handler"
	"s3MediaStreamer/app/handlers/REST/playlist_handler"
	"s3MediaStreamer/app/handlers/REST/track_handler"
	"s3MediaStreamer/app/handlers/REST/user_handler"
	amqp2 "s3MediaStreamer/app/handlers/amqp"
	"s3MediaStreamer/app/internal/app"
)

type Handlers struct {
	Audio    *audio_handler.AudioHandler
	Health   *health_handler.MonitoringHandler
	Job      *jobs_handler.JobHandler
	Otp      *otp_handler.OtpHandler
	Playlist *playlist_handler.PlaylistHandler
	Track    *track_handler.TrackHandler
	User     *user_handler.UserHandler
	Messages *amqp2.AmqpHandler
}

func NewHandlers(ctx context.Context, app *app.App) *Handlers {
	healthHandler := health_handler.NewMonitoringHandler(*app.Service.Health)
	jobHandler := jobs_handler.NewJobHandler()
	trackHandler := track_handler.NewTrackHandler(*app.Service.Track)
	userHandler := user_handler.NewUserHandler(*app.Service.Acl, *app.Service.User, *app.Service.AccessControl)
	playlistHandler := playlist_handler.NewPlaylistHandler(*app.Service.Playlist, *userHandler)
	otpHandler := otp_handler.NewOtpHandler(*app.Service.OTP)
	messageRepo, err := amqp2.NewRabbitMQHandlerWrapper(ctx, app.Cfg, app.Logger, app.Service.InitRepo.InitConnect.RabbitCon, *app.Service.Message)
	audioHandler := audio_handler.NewAudioHandler(app.Service.Audio, app.Logger)
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
