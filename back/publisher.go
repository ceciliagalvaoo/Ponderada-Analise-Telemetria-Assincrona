package main

type Publisher interface {
	Publish(body []byte) error
}