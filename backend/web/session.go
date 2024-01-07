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
	ctx         *protocol.GameplayContext
	ws          *websocket.Conn
	userID      protocol.UserID
	locale      protocol.UserLocale
	room        *GameRoom
	lastMsgTime int64
}

// NewGameSession creates a new GameSession instance.
func NewGameSession(
	userID protocol.UserID,
	UserLocale protocol.UserLocale,
	room *GameRoom,
	ws *websocket.Conn,
) GameSession {
	return GameSession{
		ctx:         nil,
		ws:          ws,
		userID:      userID,
		locale:      UserLocale,
		room:        room,
		lastMsgTime: time.Now().Unix(),
	}
}

// validateGameSessionUpdate validates a game session update.
func (s *GameSession) validateGameSessionUpdate(
	gameplayCtx *protocol.GameplayContext,
	GameplayMessageCtx *protocol.GameplayContext,
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
) protocol.GameplayMessage {
	var msg *string
	var placeInActiveSessionsPtr *int64
	var placeInLeaderboardPtr *int64
	var countInActiveSessionsPtr *int64
	var countInLeaderboardPtr *int64

	db := s.room.DB
	msgLoc := s.room.MsgLoc
	clientId := s.room.ClientID
	roodId := s.room.RoomID

	if msgLoc != nil && s.shouldSendNewRandomMessage() {
		msg = msgLoc.RandomLocalizedMessage(s.locale)
		s.lastMsgTime = time.Now().Unix()
	}

	place, _ := db.GetUserPlaceInActiveSessions(clientId, roodId, s.userID)
	count, _ := db.GetUsersCountInActiveSessions(clientId, roodId)

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
) protocol.GameplayMessage {
	var placeInActiveSessionsPtr *int64
	var placeInLeaderboardPtr *int64
	var countInActiveSessionsPtr *int64
	var countInLeaderboardPtr *int64
	var worldRecordPtr *bool

	db := s.room.DB
	clientId := s.room.ClientID
	roodId := s.room.RoomID

	place, _ := db.GetDurationPlaceInLeaderboard(clientId, roodId, gameplayRecord.Duration)
	count, _ := db.GetUsersCountInLeaderboard(clientId, roodId)

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
) error {
	// Create a new gameplay message and send it to the client as JSON
	var msg protocol.GameplayMessage
	if gameplayErr != nil {
		msg = s.gameplayError(gameplayErr)
	} else if gameplayRecord != nil {
		msg = s.gameplayRecord(gameplayRecord)
	} else if gameplayCtx != nil {
		msg = s.gameplayUpdate(gameplayCtx)
	}
	return s.ws.WriteJSON(msg)
}

// updateGameSession updates the game session state.
func (s *GameSession) updateGameSession(
	gameplayCtx *protocol.GameplayContext,
	GameplayMessageCtx *protocol.GameplayContext,
) (*protocol.GameplayContext, error) {
	var err error
	nowTimestamp := time.Now().Unix()
	pushTimestamp := *gameplayCtx.Timestamp
	holdDuration := nowTimestamp - pushTimestamp
	GameplayMessageCtx.Duration = &holdDuration
	GameplayMessageCtx.Timestamp = &pushTimestamp

	clientId := s.room.ClientID
	roodId := s.room.RoomID

	err = s.validateGameSessionUpdate(
		gameplayCtx,
		GameplayMessageCtx,
	)
	if err != nil {
		return nil, err
	}

	err = s.room.DB.SetUserDurationToActiveSessions(
		clientId,
		roodId,
		s.userID,
		holdDuration,
		nowTimestamp,
	)
	if err != nil {
		return nil, err
	}

	s.ctx = GameplayMessageCtx
	if GameplayMessageCtx.ButtonPhase != protocol.Release {
		err = s.writeNetworkMessage(
			GameplayMessageCtx,
			nil,
			nil,
		)
	}
	return GameplayMessageCtx, err
}

// closeGameSession closes the game session.
func (s *GameSession) closeGameSession() error {
	var err error
	var gameRecordPtr *protocol.GameplayRecord
	var gameErrorPtr *protocol.GameplayError

	gameplayCtx := s.ctx
	clientId := s.room.ClientID
	roodId := s.room.RoomID
	defer s.room.RemoveGameSession(s.userID)

	if gameplayCtx != nil {
		record := protocol.NewGameplayRecord(*gameplayCtx)
		addRecordToLeaderboardErr := s.room.DB.AddRecordToLeaderboard(
			clientId,
			roodId,
			s.userID,
			record,
		)
		remUserDurationFromActiveSessionsErr := s.room.DB.RemoveUserDurationFromActiveSessions(
			clientId,
			roodId,
			s.userID,
			*gameplayCtx.Timestamp,
		)
		err = errors.Join(
			err,
			addRecordToLeaderboardErr,
			remUserDurationFromActiveSessionsErr,
		)
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
	)

	return err
}

// startGameSession starts a new game session.
func (s *GameSession) startGameSession() (*protocol.GameplayContext, error) {
	if s.room.HasGameSession(s.userID) {
		return nil, ErrGameSessionAlreadyExists
	}

	gameplayCtx := protocol.NewGameplayContext()
	clientId := s.room.ClientID
	roodId := s.room.RoomID
	err := s.room.DB.SetUserDurationToActiveSessions(
		clientId,
		roodId,
		s.userID,
		*gameplayCtx.Duration,
		*gameplayCtx.Timestamp,
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
	)

	return &gameplayCtx, err
}

// MaintainGameSession maintains the game session for the user.
func (s *GameSession) MaintainGameSession() error {
	var err error
	var updatedGameplayCtx *protocol.GameplayContext

	s.ctx, err = s.startGameSession()
	if err != nil {
		gameError := protocol.NewGameplayError(err.Error())
		err_ := s.writeNetworkMessage(
			nil,
			nil,
			&gameError,
		)
		err = errors.Join(err, err_)
	} else {
		for {
			if err_ := s.ws.ReadJSON(&updatedGameplayCtx); err_ != nil {
				err = errors.Join(err, ErrFailedToReadGameSessionUpdate)
				break
			}
			s.ctx, err = s.updateGameSession(
				s.ctx,
				updatedGameplayCtx,
			)
			if err != nil || s.ctx.ButtonPhase == protocol.Release || s.room.closed {
				break
			}
		}
	}

	return errors.Join(err, s.closeGameSession())
}
