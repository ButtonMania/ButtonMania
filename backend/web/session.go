package web

import (
	"errors"
	"fmt"
	"time"

	"buttonmania.win/protocol"
	"github.com/gorilla/websocket"
)

// Define session errors
var (
	ErrGameSessionAlreadyExists        = errors.New("game session is already in progress")
	ErrGameSessionFailedToStart        = errors.New("failed to start a new game session")
	ErrFailedToReadGameSessionUpdate   = errors.New("failed to read the game session update")
	ErrGameSessionInvalidUpdate        = errors.New("invalid game session update received")
	ErrGameSessionInvalidButtonPhase   = fmt.Errorf("%w: invalid button phase", ErrGameSessionInvalidUpdate)
	ErrGameSessionInvalidPushTimestamp = fmt.Errorf("%w: invalid push timestamp", ErrGameSessionInvalidUpdate)
	ErrGameSessionInvalidHoldDuration  = fmt.Errorf("%w: invalid hold duration", ErrGameSessionInvalidUpdate)
)

// Define message update frequencies and intervals
var (
	MessageUpdateFrequencies   = [...]int64{5, 10, 30, 60, 90, 120, 160, 180, 240, 320}
	MessageUpdateTimeIntervals = [...]int64{30, 60, 120, 240, 460, 780, 1280, 3240, 5760, 10240}
)

// GameSession represents a user's game session.
type GameSession struct {
	userID      protocol.UserID
	locale      protocol.UserLocale
	room        *GameRoom
	ctx         *protocol.GameplayContext
	lastMsgTime int64
}

// NewGameSession creates a new GameSession instance.
func NewGameSession(
	userID protocol.UserID,
	UserLocale protocol.UserLocale,
	room *GameRoom,
) GameSession {
	return GameSession{
		userID:      userID,
		locale:      UserLocale,
		room:        room,
		ctx:         nil,
		lastMsgTime: time.Now().Unix(),
	}
}

// validateGameSessionUpdate validates a game session update.
func (s *GameSession) validateGameSessionUpdate(
	gameplayCtx, GameplayMessageCtx *protocol.GameplayContext,
) error {
	if gameplayCtx.Timestamp != nil && GameplayMessageCtx.Timestamp != nil &&
		*gameplayCtx.Timestamp != *GameplayMessageCtx.Timestamp {
		return fmt.Errorf(
			"%w: %d != %d",
			ErrGameSessionInvalidPushTimestamp,
			*gameplayCtx.Timestamp,
			*GameplayMessageCtx.Timestamp,
		)
	}
	if gameplayCtx.Duration != nil && GameplayMessageCtx.Duration != nil &&
		*gameplayCtx.Duration > *GameplayMessageCtx.Duration {
		return fmt.Errorf(
			"%w: %d > %d",
			ErrGameSessionInvalidHoldDuration,
			*gameplayCtx.Duration,
			*GameplayMessageCtx.Duration,
		)
	}
	if gameplayCtx.ButtonPhase == protocol.Push || gameplayCtx.ButtonPhase == protocol.Hold {
		if GameplayMessageCtx.ButtonPhase == protocol.Push {
			return ErrGameSessionInvalidButtonPhase
		}
	}
	return nil
}

// shouldSendNewRandomMessage determines if a new random message should be sent.
func (s *GameSession) shouldSendNewRandomMessage() bool {
	var intervalIndex int
	now := time.Now().Unix()
	secsSinceLastMsg := now - s.lastMsgTime
	for i, v := range MessageUpdateTimeIntervals {
		if v > *s.ctx.Duration {
			intervalIndex = i
			break
		}
	}
	currentFreq := MessageUpdateFrequencies[intervalIndex]
	return secsSinceLastMsg >= currentFreq
}

// gameplayUpdate creates a gameplay update message.
func (s *GameSession) gameplayUpdate(
	gameplayCtx *protocol.GameplayContext,
	ws *websocket.Conn,
) protocol.GameplayMessage {
	var err error
	var msg *string
	var placeInActiveSessionsPtr *int64
	var placeInLeaderboardPtr *int64
	var countInActiveSessionsPtr *int64
	var countInLeaderboardPtr *int64

	db := s.room.DB
	msgLoc := s.room.MsgLoc
	btnType := s.room.ButtonType

	if s.shouldSendNewRandomMessage() {
		msg = msgLoc.RandomLocalizedMessage(s.locale)
		s.lastMsgTime = time.Now().Unix()
	}

	place, err_ := db.GetUserPlaceInActiveSessions(btnType, s.userID)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	count, err_ := db.GetUsersCountInActiveSessions(btnType)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	placeInActiveSessionsPtr = &place
	countInActiveSessionsPtr = &count

	return protocol.NewGameplayMessage(
		gameplayCtx,
		nil,
		nil,
		msg,
		placeInActiveSessionsPtr,
		countInActiveSessionsPtr,
		placeInLeaderboardPtr,
		countInLeaderboardPtr,
		nil,
		protocol.Update,
	)
}

// gameplayRecord creates a gameplay record message.
func (s *GameSession) gameplayRecord(
	gameplayRecord *protocol.GameplayRecord,
	ws *websocket.Conn,
) protocol.GameplayMessage {
	var err error
	var placeInActiveSessionsPtr *int64
	var placeInLeaderboardPtr *int64
	var countInActiveSessionsPtr *int64
	var countInLeaderboardPtr *int64
	var worldRecordPtr *bool

	db := s.room.DB
	btnType := s.room.ButtonType

	place, err_ := db.GetDurationPlaceInLeaderboard(btnType, gameplayRecord.Duration)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	count, err_ := db.GetUsersCountInLeaderboard(btnType)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	worldRecord := place == 1
	worldRecordPtr = &worldRecord
	placeInLeaderboardPtr = &place
	countInLeaderboardPtr = &count

	return protocol.NewGameplayMessage(
		nil,
		gameplayRecord,
		nil,
		nil,
		placeInActiveSessionsPtr,
		countInActiveSessionsPtr,
		placeInLeaderboardPtr,
		countInLeaderboardPtr,
		worldRecordPtr,
		protocol.Record,
	)
}

// gameplayError creates a gameplay error message.
func (s *GameSession) gameplayError(
	gameplayErr *protocol.GameplayError,
	ws *websocket.Conn,
) protocol.GameplayMessage {
	return protocol.NewGameplayMessage(
		nil,
		nil,
		gameplayErr,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		protocol.Error,
	)
}

// writeNetworkMessage sends a gameplay message to the client.
func (s *GameSession) writeNetworkMessage(
	gameplayCtx *protocol.GameplayContext,
	gameplayRecord *protocol.GameplayRecord,
	gameplayErr *protocol.GameplayError,
	ws *websocket.Conn,
) error {
	// Create a new gameplay message and send it to the client as JSON
	var msg protocol.GameplayMessage
	if gameplayErr != nil {
		msg = s.gameplayError(gameplayErr, ws)
	} else if gameplayRecord != nil {
		msg = s.gameplayRecord(gameplayRecord, ws)
	} else if gameplayCtx != nil {
		msg = s.gameplayUpdate(gameplayCtx, ws)
	}
	return ws.WriteJSON(msg)
}

// updateGameSession updates the game session state.
func (s *GameSession) updateGameSession(
	gameplayCtx, GameplayMessageCtx *protocol.GameplayContext,
	ws *websocket.Conn,
) (*protocol.GameplayContext, error) {
	var err error
	pushTimestamp := *gameplayCtx.Timestamp
	holdDuration := time.Now().Unix() - pushTimestamp
	GameplayMessageCtx.Duration = &holdDuration
	GameplayMessageCtx.Timestamp = &pushTimestamp

	if err := s.validateGameSessionUpdate(gameplayCtx, GameplayMessageCtx); err != nil {
		return nil, err
	}

	if err := s.room.DB.SetUserDurationToActiveSessions(s.room.ButtonType, s.userID, holdDuration); err != nil {
		return nil, err
	}

	s.ctx = GameplayMessageCtx
	if GameplayMessageCtx.ButtonPhase != protocol.Release {
		err = s.writeNetworkMessage(
			GameplayMessageCtx,
			nil,
			nil,
			ws,
		)
	}
	return GameplayMessageCtx, err
}

// closeGameSession closes the game session.
func (s *GameSession) closeGameSession(
	err error,
	ws *websocket.Conn,
) (*protocol.GameplayContext, error) {
	var gameRecordPtr *protocol.GameplayRecord
	var gameErrorPtr *protocol.GameplayError

	gameplayCtx := s.ctx
	buttonTime := s.room.ButtonType
	defer s.room.RemoveGameSession(s.userID)

	if gameplayCtx != nil {
		record := protocol.NewGameplayRecord(*gameplayCtx)

		if err_ := s.room.DB.AddRecordToLeaderboard(buttonTime, s.userID, record); err_ != nil {
			err = errors.Join(err, err_)
		}

		if err_ := s.room.DB.RemoveUserDurationFromActiveSessions(buttonTime, s.userID); err_ != nil {
			err = errors.Join(err, err_)
		}

		gameRecordPtr = &record
	}

	if err != nil {
		gameError := protocol.NewGameplayError(err.Error())
		gameErrorPtr = &gameError
	}

	err = s.writeNetworkMessage(
		nil,
		gameRecordPtr,
		gameErrorPtr,
		ws,
	)

	return gameplayCtx, err
}

// startGameSession starts a new game session.
func (s *GameSession) startGameSession(ws *websocket.Conn) (*protocol.GameplayContext, error) {
	if s.room.HasGameSession(s.userID) {
		return nil, ErrGameSessionAlreadyExists
	}

	gameplayCtx := protocol.NewGameplayContext()
	err := s.room.DB.SetUserDurationToActiveSessions(
		s.room.ButtonType,
		s.userID,
		*gameplayCtx.Duration,
	)
	if err != nil {
		return nil, err
	}

	s.ctx = &gameplayCtx
	s.room.AddGameSession(s.userID, s)

	err = s.writeNetworkMessage(
		&gameplayCtx,
		nil,
		nil,
		ws,
	)

	return &gameplayCtx, err
}

// MaintainGameSession maintains the game session for the user.
func (s *GameSession) MaintainGameSession(ws *websocket.Conn) error {
	var err error
	var updatedGameplayCtx *protocol.GameplayContext

	s.ctx, err = s.startGameSession(ws)
	if err != nil {
		gameError := protocol.NewGameplayError(err.Error())
		err_ := s.writeNetworkMessage(
			nil,
			nil,
			&gameError,
			ws,
		)
		return errors.Join(err, err_)
	}

	defer func() {
		defer s.closeGameSession(
			err,
			ws,
		)
	}()

	for {
		if err_ := ws.ReadJSON(&updatedGameplayCtx); err_ != nil {
			err = errors.Join(err, ErrFailedToReadGameSessionUpdate)
			break
		}
		s.ctx, err = s.updateGameSession(
			s.ctx,
			updatedGameplayCtx,
			ws,
		)
		if err != nil || s.ctx.ButtonPhase == protocol.Release {
			break
		}
	}

	return err
}
