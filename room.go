package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
)

type (

	// Room is a component that represents a room
	Room struct {
		group *nano.Group
	}

	// RoomManager represents a component that contains a bundle of room
	RoomManager struct {
		component.Base
		timer *scheduler.Timer
		rooms map[int]*Room
	}

	// UserMessage represents a message that a user sent
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// JoinRequest message will be receieved when a user requests
	// to join a room
	JoinRequest struct {
		RoomID int `json:"roomID"`
	}

	// NewRoomRequest message will be receieved when a user requests to
	// create a new room
	NewRoomRequest struct {
		Name string `json:"name"`
	}

	// NewRoomResponse message will be received when a new room is successfully
	// created
	NewRoomResponse struct {
		Name   string `json:"name"`
		RoomID int    `json:"roomID"`
	}

	// NewUser message will be receieved when a new user join room
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	AllMembers struct {
		Members []int64 `json:"members"`
	}

	// JoinResponse represents the result of joining room
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

const (
	testRoomID = 1
	roomIDKey  = "ROOM_ID"
)

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: map[int]*Room{},
	}
}

// AfterInit component lifetime callback
func (mgr *RoomManager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		if !s.HasKey(roomIDKey) {
			return
		}
		room := s.Value(roomIDKey).(*Room)
		room.group.Leave(s)
	})
	mgr.timer = scheduler.NewTimer(time.Minute, func() {
		for roomID, room := range mgr.rooms {
			println(fmt.Sprintf("UserCount: RoomID=%d, Time=%s, Count=%d",
				roomID, time.Now().String(), room.group.Count()))
		}
	})
}

// Join adds the the new player to a group
func (mgr *RoomManager) Join(s *session.Session, request *JoinRequest) error {

	room, found := mgr.rooms[request.RoomID]
	if !found {
		return s.Response(&JoinResponse{Result: fmt.Sprintf("failure: roomID %d does not exist", request.RoomID)})
	}

	fakeUID := s.ID()
	s.Bind(fakeUID)
	s.Set(roomIDKey, room)
	s.Push("onMembers", &AllMembers{Members: room.group.Members()})
	room.group.Broadcast("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
	room.group.Add(s)
	return s.Response(&JoinResponse{Result: "success"})
}

// Create will create a room and return the roomID to the client
func (mgr *RoomManager) Create(s *session.Session, request *NewRoomRequest) error {

	name := request.Name
	roomID := rand.Intn(100)
	room := &Room{
		group: nano.NewGroup(fmt.Sprintf("room-%d", roomID)),
	}
	mgr.rooms[roomID] = room
	return s.Response(&NewRoomResponse{RoomID: roomID, Name: name})

}

// Message sync last message to all members
func (mgr *RoomManager) Message(s *session.Session, msg *UserMessage) error {
	if !s.HasKey(roomIDKey) {
		return fmt.Errorf("not join room yet")
	}
	println("Message", msg.Name)
	room := s.Value(roomIDKey).(*Room)
	return room.group.Broadcast("onMessage", msg)
}
