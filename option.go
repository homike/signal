package signal

type Option func(*Signal)

func WithDup(dup bool) Option {
	return func(sig *Signal) {
		sig.dup = dup
	}
}

func WithSignalTerm(f func()) Option {
	return func(sig *Signal) {
		sig.signalTerm = f
	}
}

func WithSignalHup(f func()) Option {
	return func(sig *Signal) {
		sig.signalHUP = f
	}
}
