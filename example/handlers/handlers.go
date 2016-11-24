package handlers

import (
	"sync"

	"github.com/jargv/plumbus"
)

type Counter struct {
	lock     sync.Mutex
	HitCount int `json:"count"`
}

//go:generate plumbus Counter.Incr
func (c *Counter) Incr() *Counter {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.HitCount++
	return c
}

//go:generate plumbus Counter.Count
func (c Counter) Count() map[string]interface{} {
	return map[string]interface{}{
		"count": c.HitCount,
	}
}

type messageQueryParam string

func EchoParam(message messageQueryParam) map[string]string {
	return map[string]string{
		"message": string(message),
	}
}

//go:generate plumbus Error
func Error() error {
	return plumbus.Errorf(404, "this is an error")
}

type User struct {
	DisplayName string
}

func (example *User) Documentation() string {
	*example = User{
		DisplayName: "Some Guy",
	}
	return `
	  This represents the User
	`
}

type userIdQueryParam int

func (userIdQueryParam) Documentation() string {
	return `
	  the id of the user
	`
}

func EditUser(id userIdQueryParam, user *User) {

}

func GetUser(id userIdQueryParam) *User {
	return nil
}
