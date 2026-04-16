package websocket

import (
	"encoding/json"
	"log/slog"
	"time"
)

func (c *Client) JoinRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Rooms[room] {
		slog.Debug("Client already in room", "client_id", c.ID, "room", room)
		return
	}

	c.Hub.mu.Lock()

	if c.Hub.rooms[room] == nil {
		c.Hub.rooms[room] = make(map[string]*Client)
	}
	c.Hub.rooms[room][c.ID] = c
	c.Rooms[room] = true

	slog.Info("Client joined room", "client_id", c.ID, "room", room, "user_id", c.UserID.Hex(), "total_clients_in_room", len(c.Hub.rooms[room]))

	userID := c.UserID.Hex()
	clientID := c.ID

	c.Hub.mu.Unlock()

	c.Hub.BroadcastToRoom(room, EventUserJoined, map[string]any{
		"user_id":   userID,
		"client_id": clientID,
	})

	c.Hub.mu.RLock()
	if len(room) > 6 && room[len(room)-6:] == ":users" {
		seenUsers := make(map[string]bool)
		var users []map[string]any

		for _, client := range c.Hub.rooms[room] {
			uid := client.UserID.Hex()
			if !seenUsers[uid] {
				seenUsers[uid] = true
				userInfo := map[string]any{
					"user_id":   uid,
					"name":      client.UserName,
					"team_id":   client.TeamID.Hex(),
					"team_name": client.TeamName,
					"presence":  client.PresenceState,
				}
				if client.PresenceState == "away" && client.AwaySince != nil {
					awayDuration := int64(time.Since(*client.AwaySince).Seconds())
					userInfo["away_duration_secs"] = awayDuration
				}
				users = append(users, userInfo)
			}
		}

		snapshot := Message{
			Type: EventPresenceSnapshot,
			Data: map[string]any{
				"type":  EventPresenceSnapshot,
				"room":  room,
				"users": users,
			},
		}
		jsonData, err := json.Marshal(snapshot)
		if err != nil {
			slog.Error("Error marshaling presence snapshot", "error", err)
		} else {
			select {
			case c.Send <- jsonData:
			default:
				slog.Debug("Failed to send presence snapshot to client", "client_id", c.ID)
			}
		}
	}
	c.Hub.mu.RUnlock()
}

func (c *Client) LeaveRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Rooms[room] {
		return
	}

	c.Hub.mu.Lock()

	userID := c.UserID.Hex()
	clientID := c.ID

	if roomClients, exists := c.Hub.rooms[room]; exists {
		delete(roomClients, c.ID)
		if len(roomClients) == 0 {
			delete(c.Hub.rooms, room)
		}
	}
	delete(c.Rooms, room)

	c.Hub.mu.Unlock()

	c.Hub.BroadcastToRoom(room, EventUserLeft, map[string]any{
		"user_id":   userID,
		"client_id": clientID,
	})

	slog.Debug("Client left room", "client_id", c.ID, "room", room)
}
