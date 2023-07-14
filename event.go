package bubbly

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sulaiman-coder/goeventbus"
)

var (
	_ EventHandler = (*EventDispatcher)(nil)
	_ interface {
		EventHandler
		MessageListener
		HandleWaiter
	} = (*HandlerCollection)(nil)
)

type EventHandlerFn func(eventbus.Event) []tea.Model

type EventHandler interface {
	eventbus.Responder
	Handle(eventbus.Event) []tea.Model
}

type MessageListener interface {
	OnMessage(tea.Msg)
}

type HandleWaiter interface {
	Wait()
}

type EventDispatcher struct {
	dispatch map[eventbus.EventType]EventHandlerFn
	types    []eventbus.EventType
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		dispatch: map[eventbus.EventType]EventHandlerFn{},
	}
}

func (d *EventDispatcher) AddHandlers(handlers map[eventbus.EventType]EventHandlerFn) {
	for t, h := range handlers {
		d.AddHandler(t, h)
	}
}

func (d *EventDispatcher) AddHandler(t eventbus.EventType, fn EventHandlerFn) {
	d.dispatch[t] = fn
	d.types = append(d.types, t)
}

func (d EventDispatcher) RespondsTo() []eventbus.EventType {
	return d.types
}

func (d EventDispatcher) Handle(e eventbus.Event) []tea.Model {
	if fn, ok := d.dispatch[e.Type]; ok {
		return fn(e)
	}
	return nil
}

type HandlerCollection struct {
	handlers []EventHandler
}

func NewHandlerCollection(handlers ...EventHandler) *HandlerCollection {
	return &HandlerCollection{
		handlers: handlers,
	}
}

func (h *HandlerCollection) Append(handlers ...EventHandler) {
	h.handlers = append(h.handlers, handlers...)
}

func (h HandlerCollection) RespondsTo() []eventbus.EventType {
	var ret []eventbus.EventType
	for _, handler := range h.handlers {
		ret = append(ret, handler.RespondsTo()...)
	}
	return ret
}

func (h HandlerCollection) Handle(event eventbus.Event) []tea.Model {
	var ret []tea.Model
	for _, handler := range h.handlers {
		ret = append(ret, handler.Handle(event)...)
	}
	return ret
}

func (h HandlerCollection) OnMessage(msg tea.Msg) {
	for _, handler := range h.handlers {
		if listener, ok := handler.(MessageListener); ok {
			listener.OnMessage(msg)
		}
	}
}

func (h HandlerCollection) Wait() {
	for _, handler := range h.handlers {
		if listener, ok := handler.(HandleWaiter); ok {
			listener.Wait()
		}
	}
}
