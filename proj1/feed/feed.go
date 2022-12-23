package feed

import (
	"math"
	"proj1/lock"
)

// Feed represents a user's twitter feed
// You will add to this interface the implementations as you complete them.
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	Show() []interface{}
}

// feed is the internal representation of a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation. You can assume the feed will not have duplicate posts
type feed struct {
	head *post // a pointer to the beginning post
	tail *post // a pointer to the last post
	lock *lock.RWLock
}

// post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation.
type post struct {
	body      string  // the text of the post
	timestamp float64 // Unix timestamp of the post
	removed   bool    // used to determine if a post has been removed
	next      *post   // the next post in the feed
	lock      *lock.RWLock
}

// NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	rwLock := lock.NewRWLock()
	return &post{body: body, timestamp: timestamp, next: next, removed: false, lock: rwLock}
}

// NewFeed creates a empy user feed
func NewFeed() Feed {
	head := newPost("", math.MaxFloat64, nil)
	tail := newPost("", -math.MaxFloat64, nil)
	head.next = tail
	rwLock := lock.NewRWLock()
	return &feed{head, tail, rwLock}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	for {
		prev := f.head
		curr := f.head.next

		// Iterate till the end or when the timestamp is less than the current timestamp (place to insert)
		for curr != nil && curr.timestamp > timestamp {
			prev = curr
			curr = curr.next
		}

		// Lock the previous and current posts
		prev.lock.Lock()
		curr.lock.Lock()

		// Check the posts
		if validate(prev, curr) {
			// If the timestamp is the same as the current timestamp, then replace the body
			if curr.timestamp == timestamp {
				curr.body = body
				// Unlock the posts and return
				prev.lock.Unlock()
				curr.lock.Unlock()
				return
			} else {
				// We have found the place to insert the new post
				newPost := newPost(body, timestamp, curr)
				prev.next = newPost

				// Unlock the posts and return
				prev.lock.Unlock()
				curr.lock.Unlock()
				return
			}
		}

		// Unlock the posts
		curr.lock.Unlock()
		prev.lock.Unlock()
	}
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	for {
		prev := f.head
		curr := f.head.next
		// Iterate till the end or when the timestamp is less than the current timestamp (place that the post should be)
		for curr != nil && curr.timestamp > timestamp {
			prev = curr
			curr = curr.next
		}

		// If the timestamp to remove cannot be found, return false
		if curr == nil || curr.timestamp != timestamp {
			return false
		}

		// Lock the previous and current posts
		prev.lock.Lock()
		curr.lock.Lock()

		// Check the posts and if this is the post to remove
		if validate(prev, curr) && curr.timestamp == timestamp {
			// Remove the post
			curr.removed = true
			prev.next = curr.next
			curr.lock.Unlock()
			prev.lock.Unlock()
			return true
		}

		// Unlock the posts
		curr.lock.Unlock()
		prev.lock.Unlock()
	}
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	curr := f.head
	// Iterate till the end or when the timestamp is less than the current timestamp (place that the post should be)
	for curr != nil && curr.timestamp > timestamp {
		curr = curr.next
	}
	// If the timestamp cannot be found, return false
	if curr == nil || curr.timestamp != timestamp {
		return false
	}
	// If the timestamp is found, return true
	return true
}

func validate(prev, curr *post) bool {
	return !prev.removed && !curr.removed && prev.next == curr
}

// Function to display the entire feed
func (f *feed) Show() []interface{} {

	f.lock.RLock()

	var currPost *post
	currPost = f.head.next

	var displayFeed []interface{}

	for currPost != nil {
		// Create post to display
		displayPost := make(map[string]interface{})

		displayPost["body"] = currPost.body
		displayPost["timestamp"] = currPost.timestamp

		// Add post to the feed
		displayFeed = append(displayFeed, displayPost)

		currPost = currPost.next
	}

	if len(displayFeed) > 0 {
		displayFeed = displayFeed[:len(displayFeed)-1]
	}

	f.lock.RUnlock()

	return displayFeed
}
