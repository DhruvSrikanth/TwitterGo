package feed

import (
	"fmt"
	"proj1/lock"
)

// Feed represents a user's twitter feed
// You will add to this interface the implementations as you complete them.
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	Display()
}

// feed is the internal representation of a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation. You can assume the feed will not have duplicate posts
type feed struct {
	start *post        // a pointer to the beginning post
	lock  *lock.RWLock // a pointer to the read-write lock
}

// post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation.
type post struct {
	body      string  // the text of the post
	timestamp float64 // Unix timestamp of the post
	next      *post   // the next post in the feed
}

// NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

// NewFeed creates a empy user feed
func NewFeed() Feed {
	rwLock := lock.NewRWLock()
	return &feed{start: nil, lock: rwLock}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	// Create the post
	addPost := newPost(body, timestamp, nil)

	// Lock the feed
	f.lock.Lock()

	// Figure out where to insert the node
	// If the feed is empty, insert the post at the beginning
	if f.start == nil {
		f.start = addPost
		// Unlock the feed
		f.lock.Unlock()
		return
	}

	var currPost *post
	currPost = f.start

	// Find the post with a timestamp less than the given timestamp
	for (currPost.next != nil) && (currPost.next.timestamp > timestamp) {
		currPost = currPost.next
	}

	// If the post is at the end of the feed, insert the post at the end
	if currPost.next == nil {
		currPost.next = addPost
		// Unlock the feed
		f.lock.Unlock()
		return
	}

	// Insert the post
	addPost.next = currPost.next
	currPost.next = addPost

	// Unlock the feed
	f.lock.Unlock()

}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	// Lock the feed
	f.lock.Lock()

	// If the feed is empty, return false
	if f.start == nil {
		// Unlock the feed
		f.lock.Unlock()

		return false
	}

	var currPost *post
	currPost = f.start

	// Check if the post is at the beginning of the feed
	if currPost.timestamp == timestamp {
		f.start = currPost.next
		// Unlock the feed
		f.lock.Unlock()

		return true
	}

	// Find the post with the given timestamp
	for currPost.next != nil {
		if currPost.next.timestamp == timestamp {
			currPost.next = currPost.next.next
			// Unlock the feed
			f.lock.Unlock()

			return true
		}
		currPost = currPost.next
	}

	// Unlock the feed
	f.lock.Unlock()

	// Havent found the post
	return false

}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	// lock the feed
	f.lock.RLock()

	// Find the post with the given timestamp
	var currPost *post
	currPost = f.start

	for currPost != nil {
		if currPost.timestamp == timestamp {
			// Unlock the feed
			f.lock.RUnlock()
			return true
		}
		currPost = currPost.next
	}
	// Unlock the feed
	f.lock.RUnlock()

	return false
}

// Function to display the entire feed
func (f *feed) Display() {
	// lock the feed
	f.lock.RLock()

	var currPost *post
	currPost = f.start

	// Print every post in the feed
	for currPost != nil {
		fmt.Printf("\n--x--Post:\n%s\n\nMade at - %f--x--\n", currPost.body, currPost.timestamp)
		currPost = currPost.next
	}

	// Unlock the feed
	f.lock.RUnlock()
}
