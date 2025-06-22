package tui

import (
	"bytes"
	"fmt"
	"github.com/crookdc/pia"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type App struct {
	resolver pia.KeyResolver
	*tview.Application
	pages   *tview.Pages
	console *console
	content *content
	finder  *finder
	history *history
}

type SwitchPageCommand struct {
	Page     string
	Callback func(*App) error
}

func (a *App) view(path string) {
	tx, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer tx.Close()
	src, err := io.ReadAll(pia.WrapReader(a.resolver, tx))
	if err != nil {
		panic(err)
	}
	a.display(string(src))
}

func (a *App) execute(path string) {
	cfg, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	tx, err := pia.ParseTransaction(filepath.Dir(path), pia.WrapReader(a.resolver, bytes.NewReader(cfg)))
	if err != nil {
		panic(err)
	}
	client := pia.Pia{
		WorkingDirectory: filepath.Dir(path),
		Output:           a.console.log,
		Resolver:         a.resolver,
	}
	res, err := client.Execute(tx)
	if err != nil {
		panic(err)
	}
	sb := strings.Builder{}
	var color string
	switch res.StatusCode / 100 {
	case 1, 3:
		color = "yellow"
	case 2:
		color = "green"
	case 4, 5:
		color = "red"
	default:
		color = "white"
	}
	sb.WriteString(fmt.Sprintf("[%s]Status: %s\n", color, res.Status))
	for k, v := range res.Header {
		sb.WriteString(fmt.Sprintf("[blue]%s: [white]%s\n", k, strings.Join(v, ", ")))
	}
	sb.WriteString("\n\n")
	if res.Body != nil {
		raw, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		sb.WriteString(string(raw))
	}
	a.history.push(entry{
		method:    tx.Method,
		endpoint:  tx.URL.Target,
		timestamp: time.Now(),
		text:      sb.String(),
	})
	a.display(sb.String())
}

func (a *App) display(text string) {
	a.content.text.SetText(text)
	a.pages.SwitchToPage("content")
}

func (a *App) input(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEsc {
		a.pages.SwitchToPage("dashboard")
		return nil
	}
	switch ev.Rune() {
	case 'h':
		a.history.enter()
		a.pages.SwitchToPage("history")
		return nil
	case 'f':
		a.pages.SwitchToPage("finder")
		return nil
	case 'c':
		a.console.enter()
		if a.pages.HasPage("console") {
			a.pages.RemovePage("console")
		} else {
			a.pages.AddPage("console", a.console.root(), true, true)
		}
		return nil
	default:
		return ev
	}
}

func Run(wd string, props map[string]string) error {
	if err := clipboard.Init(); err != nil {
		return err
	}
	app := App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
		console:     newConsole(bytes.NewBufferString("")),
		content:     newContent(),
		finder:      newFinder(wd),
		history:     newHistory(128),
		resolver: pia.DelegatingKeyResolver{
			Delegates: map[string]pia.KeyResolver{
				"env":   pia.EnvironmentResolver{},
				"props": pia.MapResolver(props),
			},
		},
	}
	app.history.viewCallback = func(e *entry) {
		app.display(e.text)
	}
	app.finder.executeCallback = app.execute
	app.finder.viewCallback = app.view
	app.pages.AddPage("dashboard", tview.NewTextView().SetText(`
	
	pia - the postman alternative for technical people. 

	Usage:
	f - open finder window
		x - execute currently selected file
			y - copy output to clipboard
		v - view file contents after preprocessing
			y - copy output to clipboard
	h - open history
	c - toggle console

	<ESC> brings you back here.

	created by crookdc @ github.com/crookdc
	`), true, true)
	app.pages.AddPage("finder", app.finder.root(), true, false)
	app.pages.AddPage("content", app.content.root(), true, false)
	app.pages.AddPage("history", app.history.root(), true, false)
	app.SetInputCapture(app.input)
	return app.SetRoot(app.pages, true).Run()
}
