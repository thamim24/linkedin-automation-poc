package stealth

import (
	"time"

	"linkedin-automation-poc/internal/config"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Typing handles human-like keyboard input
type Typing struct {
	config config.TypingConfig
	page   *rod.Page
}

// NewTyping creates a new typing controller
func NewTyping(page *rod.Page, cfg config.TypingConfig) *Typing {
	return &Typing{
		config: cfg,
		page:   page,
	}
}

// Type types text with human-like characteristics
func (t *Typing) Type(text string) error {
	words := splitIntoWords(text)

	for i, word := range words {
		if err := t.typeWord(word); err != nil {
			return err
		}

		// Add space between words (except last word)
		if i < len(words)-1 {
			if err := t.typeCharacter(' '); err != nil {
				return err
			}

			// Longer pause between words
			delay := RandomDelay(t.config.WordPauseMin, t.config.WordPauseMax)
			time.Sleep(delay)
		}
	}

	return nil
}

// typeWord types a single word with possible errors and corrections
func (t *Typing) typeWord(word string) error {
	for i, char := range word {
		// Random chance of making a typo
		if shouldMakeError(t.config.ErrorRate) && i > 0 {
			// Type wrong character
			wrongChar := getRandomChar()
			if err := t.typeCharacter(wrongChar); err != nil {
				return err
			}

			// Small pause before correction
			time.Sleep(100 * time.Millisecond)

			// Backspace
			if err := t.page.Keyboard.Press(input.Backspace); err != nil {
				return err
			}

			time.Sleep(50 * time.Millisecond)
		}

		// Type the correct character
		if err := t.typeCharacter(char); err != nil {
			return err
		}

		// Variable delay between keystrokes
		delay := RandomDelay(t.config.MinKeystrokeDelay, t.config.MaxKeystrokeDelay)
		time.Sleep(delay)
	}

	return nil
}

// typeCharacter types a single character
func (t *Typing) typeCharacter(char rune) error {
	return t.page.Keyboard.Type(input.Key(char))
}

// TypeIntoElement types text into a specific element
func (t *Typing) TypeIntoElement(element *rod.Element, text string) error {
	// Click the element first
	if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}

	// Small delay after click
	time.Sleep(200 * time.Millisecond)

	// Clear existing text if any
	if err := t.ClearElement(element); err != nil {
		return err
	}

	// Type the text
	return t.Type(text)
}

// ClearElement clears text from an element
func (t *Typing) ClearElement(element *rod.Element) error {
	// Select all
	if err := t.page.Keyboard.Press(input.ControlLeft); err != nil {
		return err
	}
	if err := t.page.Keyboard.Type(input.Key('a')); err != nil {
		return err
	}
	if err := t.page.Keyboard.Release(input.ControlLeft); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	// Delete
	if err := t.page.Keyboard.Press(input.Backspace); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}

// PressEnter presses the Enter key
func (t *Typing) PressEnter() error {
	delay, _ := randomInt(100, 300)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if err := t.page.Keyboard.Press(input.Enter); err != nil {
		return err
	}

	delay, _ = randomInt(200, 500)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	return nil
}

// PressTab presses the Tab key
func (t *Typing) PressTab() error {
	delay, _ := randomInt(100, 300)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if err := t.page.Keyboard.Press(input.Tab); err != nil {
		return err
	}

	return nil
}

// splitIntoWords splits text into words
func splitIntoWords(text string) []string {
	var words []string
	var currentWord []rune

	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if len(currentWord) > 0 {
				words = append(words, string(currentWord))
				currentWord = []rune{}
			}
		} else {
			currentWord = append(currentWord, char)
		}
	}

	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	return words
}

// shouldMakeError determines if a typing error should occur
func shouldMakeError(errorRate float64) bool {
	if errorRate <= 0 {
		return false
	}

	rand := randomFloat(0, 1)
	return rand < errorRate
}

// getRandomChar returns a random character for typos
func getRandomChar() rune {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	idx, _ := randomInt(0, len(chars)-1)
	return chars[idx]
}

// TypeSlowly types text more slowly (for passwords or sensitive fields)
func (t *Typing) TypeSlowly(text string) error {
	for _, char := range text {
		if err := t.typeCharacter(char); err != nil {
			return err
		}

		// Longer delays for sensitive input
		delay, _ := randomInt(150, 400)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return nil
}
