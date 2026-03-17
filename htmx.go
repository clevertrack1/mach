package mach

import (
	"encoding/json"
)

// HTMX returns the HTMX helper.
func (ctx *Context) HTMX() *HTMX {
	return &HTMX{ctx: ctx}
}

// HTMX helper provides methods for interacting with HTMX headers.
type HTMX struct {
	ctx *Context
}

// IsHTMX returns true if the request is an HTMX request.
func (h *HTMX) IsHTMX() bool {
	return h.ctx.GetHeader("HX-Request") == "true"
}

// IsBoosted returns true if the request is via an element using hx-boost.
func (h *HTMX) IsBoosted() bool {
	return h.ctx.GetHeader("HX-Boosted") == "true"
}

// IsHistoryRestoreRequest returns true if the request is for history restoration after a miss in the local history cache.
func (h *HTMX) IsHistoryRestoreRequest() bool {
	return h.ctx.GetHeader("HX-History-Restore-Request") == "true"
}

// CurrentURL returns the current URL of the browser.
func (h *HTMX) CurrentURL() string {
	return h.ctx.GetHeader("HX-Current-URL")
}

// Prompt returns the user response to an hx-prompt.
func (h *HTMX) Prompt() string {
	return h.ctx.GetHeader("HX-Prompt")
}

// Target returns the id of the target element if it exists.
func (h *HTMX) Target() string {
	return h.ctx.GetHeader("HX-Target")
}

// Trigger returns the id of the triggered element if it exists.
func (h *HTMX) Trigger() string {
	return h.ctx.GetHeader("HX-Trigger")
}

// TriggerName returns the name of the triggered element if it exists.
func (h *HTMX) TriggerName() string {
	return h.ctx.GetHeader("HX-Trigger-Name")
}

// Response Headers

// PushURL pushes a new URL into the browser location bar.
func (h *HTMX) PushURL(url string) {
	h.ctx.SetHeader("HX-Push-Url", url)
}

// ReplaceURL replaces the current URL in the browser location bar.
func (h *HTMX) ReplaceURL(url string) {
	h.ctx.SetHeader("HX-Replace-Url", url)
}

// Location allows you to do a client-side redirect that does not do a full page reload.
func (h *HTMX) Location(url string) {
	h.ctx.SetHeader("HX-Location", url)
}

// Reswap allows you to specify how the response will be swapped.
func (h *HTMX) Reswap(swap string) {
	h.ctx.SetHeader("HX-Reswap", swap)
}

// Retarget allows you to target a different element on the page of the response.
func (h *HTMX) Retarget(selector string) {
	h.ctx.SetHeader("HX-Retarget", selector)
}

// Reselect allows you to choose which part of the response is used to be swapped in.
func (h *HTMX) Reselect(selector string) {
	h.ctx.SetHeader("HX-Reselect", selector)
}

// TriggerResponse allows you to trigger client side events.
// If events is a string, it sets it as the HX-Trigger header.
// If events is a map or struct, it serializes it to JSON and sets it as the HX-Trigger header.
// To avoid overwriting previously set events, this method appends to the existing HX-Trigger header if it exists.
func (h *HTMX) TriggerResponse(events any) {
	h.setTriggerHeader("HX-Trigger", events)
}

// TriggerAfterSettleResponse allows you to trigger client side events after the settle step.
func (h *HTMX) TriggerAfterSettleResponse(events any) {
	h.setTriggerHeader("HX-Trigger-After-Settle", events)
}

// TriggerAfterSwapResponse allows you to trigger client side events after the swap step.
func (h *HTMX) TriggerAfterSwapResponse(events any) {
	h.setTriggerHeader("HX-Trigger-After-Swap", events)
}

func (h *HTMX) setTriggerHeader(header string, events any) {
	if events == nil {
		return
	}

	var newPayload string
	if s, ok := events.(string); ok {
		newPayload = s
	} else {
		payload, err := json.Marshal(events)
		if err != nil {
			return
		}
		newPayload = string(payload)
	}

	existing := h.ctx.GetResponseHeader(header)
	if existing == "" {
		h.ctx.SetHeader(header, newPayload)
		return
	}

	// If it's already a JSON object (starts with {), we might want to merge it.
	// But simple concatenation with comma is usually what HTMX expects for multiple events if they are simple strings.
	// If they are JSON, it gets complicated. HTMX supports multiple events in one header via JSON object keys.
	// For simplicity, let's try to merge if both are JSON objects.
	if existing[0] == '{' && newPayload[0] == '{' {
		var existingMap, newMap map[string]any
		if err := json.Unmarshal([]byte(existing), &existingMap); err == nil {
			if err := json.Unmarshal([]byte(newPayload), &newMap); err == nil {
				for k, v := range newMap {
					existingMap[k] = v
				}
				merged, _ := json.Marshal(existingMap)
				h.ctx.SetHeader(header, string(merged))
				return
			}
		}
	}

	// Otherwise, just append if it's not already there.
	// HTMX supports comma-separated event names.
	h.ctx.SetHeader(header, existing+", "+newPayload)
}

// Refresh triggers a client side full refresh of the page.
func (h *HTMX) Refresh() {
	h.ctx.SetHeader("HX-Refresh", "true")
}

// Redirect allows you to do a client-side redirect to a new location.
func (h *HTMX) Redirect(url string) {
	h.ctx.SetHeader("HX-Redirect", url)
}
