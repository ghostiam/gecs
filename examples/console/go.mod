module github.com/ghostiam/gecs/examples/console

go 1.17

replace github.com/ghostiam/gecs v0.0.0-20211219234822-d9cf0f8f1681 => ../../

require (
	github.com/ghostiam/gecs v0.0.0-20211219221908-436df6bc4d19
	github.com/nsf/termbox-go v1.1.1
)

require github.com/mattn/go-runewidth v0.0.9 // indirect
