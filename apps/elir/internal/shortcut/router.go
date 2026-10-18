package shortcut

// Router runs on the event tap/UI thread. Available callbacks read atomic
// business state; actions queue work and never perform I/O here.
type Router struct {
	Bindings  Bindings
	Available map[Action]func() bool
	Fire      map[Action]func()
	Exclusive func() bool
	Selection func(string, Mods) bool
	Changed   func()
	Editing   bool
	Candidate Binding
	held      map[int64]bool
	taken     map[int64]bool
	mods      Mods
	pending   *gesture
	blocked   bool
}
type gesture struct {
	code  int64
	key   string
	mods  Mods
	dirty bool
}

func New(b Bindings) *Router {
	return &Router{Bindings: Clone(b), Available: map[Action]func() bool{}, Fire: map[Action]func(){}, held: map[int64]bool{}, taken: map[int64]bool{}}
}
func (r *Router) Register(a Action, available func() bool, fire func()) {
	r.Available[a] = available
	r.Fire[a] = fire
}
func (r *Router) Set(b Bindings) {
	r.Bindings = Clone(b)
	r.pending = nil
	r.blocked = len(r.held) > 0 || r.mods != 0
}
func (r *Router) Begin() {
	r.Editing = true
	r.Candidate = Binding{}
	r.pending = nil
	r.blocked = len(r.held) > 0 || r.mods != 0
}
func (r *Router) End() {
	r.Editing = false
	r.Candidate = Binding{}
	r.pending = nil
	r.blocked = len(r.held) > 0 || r.mods != 0
}

// Reset abandons triggers but retains swallowed releases until they arrive.
func (r *Router) Reset() {
	r.pending = nil
	clear(r.held)
	r.mods = 0
	r.blocked = false
	r.Editing = false
}
func (r *Router) flags(code int64, f uint64) {
	r.mods = Flags(f)
	if r.pending != nil {
		r.pending.mods |= r.mods
	}
}
func (r *Router) FlagsChanged(code int64, f uint64) bool {
	if code == 179 {
		code = 63
	}
	r.flags(code, f)
	if Key(code) == "Fn" {
		if f&0x800000 != 0 {
			return r.down(code, false)
		}
		return r.KeyUp(code)
	}
	if mod, ok := Modifier(code); ok {
		down := r.mods&mod != 0
		if down {
			r.held[code] = true
			if r.Editing && !r.blocked {
				r.taken[code] = true
			}
			return r.taken[code]
		}
		return r.KeyUp(code)
	}
	return false
}
func (r *Router) KeyDown(code int64, f uint64, repeat bool) bool {
	r.flags(code, f)
	return r.down(code, repeat)
}
func (r *Router) down(code int64, repeat bool) bool {
	if r.pending != nil && r.pending.code != code {
		r.pending.dirty = true
	}
	if r.held[code] || repeat {
		return r.taken[code]
	}
	delete(r.taken, code)
	r.held[code] = true
	key := Key(code)
	if r.blocked {
		return false
	}
	if r.Exclusive != nil && r.Exclusive() {
		if r.Selection != nil && r.Selection(key, r.mods) {
			r.taken[code] = true
			return true
		}
		return false
	}
	if r.Editing {
		r.taken[code] = true
		if r.pending == nil {
			r.pending = &gesture{code: code, key: key, mods: r.mods}
		}
		return true
	}
	// Fn combinations with other keys belong to the system.
	if key != "Fn" && r.held[63] {
		return false
	}
	// Voice gestures resolve on release, including Shift pressed after the key.
	for _, a := range []Action{Dictate, Translate} {
		b := r.Bindings[a]
		if b.Key != key || !r.enabled(a) {
			continue
		}
		if b.Match(key, r.mods) {
			r.pending = &gesture{code: code, key: key, mods: r.mods}
			for heldCode := range r.held {
				if heldCode != code {
					if _, modifier := Modifier(heldCode); !modifier {
						r.pending.dirty = true
					}
				}
			}
			r.taken[code] = true
			return true
		}
	}
	swallowed := false
	for _, a := range Actions {
		if local(a) || a == Dictate || a == Translate {
			continue
		}
		if r.enabled(a) && r.Bindings[a].Match(key, r.mods) {
			r.taken[code] = true
			swallowed = true
			r.Fire[a]()
		}
	}
	return swallowed
}
func (r *Router) KeyUp(code int64) bool {
	owned := r.taken[code]
	delete(r.taken, code)
	delete(r.held, code)
	p := r.pending
	if p != nil && p.code == code {
		r.pending = nil
		exclusive := r.Exclusive != nil && r.Exclusive()
		if !p.dirty && !r.blocked && !exclusive {
			if r.Editing {
				r.Candidate = Binding{Key: p.key, Modifiers: Modifiers(p.mods)}
				if r.Candidate.Key == "" {
					r.Candidate.Key = "Unsupported"
				}
				if r.Changed != nil {
					r.Changed()
				}
			} else {
				for _, a := range []Action{Dictate, Translate} {
					if r.enabled(a) && r.Bindings[a].Match(p.key, p.mods) {
						r.Fire[a]()
						break
					}
				}
			}
		}
	}
	if r.blocked && len(r.held) == 0 && r.mods == 0 {
		r.blocked = false
	}
	return owned
}
func (r *Router) enabled(a Action) bool {
	return r.Fire[a] != nil && (r.Available[a] == nil || r.Available[a]())
}

// Claim pairs a key consumed by a native selection window with its release.
func (r *Router) Claim(key string) {
	for code := range r.held {
		if Key(code) == key {
			r.taken[code] = true
		}
	}
}
