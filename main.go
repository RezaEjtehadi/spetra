package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-autostart"
	"github.com/go-vgo/robotgo"
	"github.com/spf13/viper"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"select-speak/datastatic"
	"select-speak/engine"
	"select-speak/gettext"
	"strconv"
	"strings"
	"time"
)

type App struct {
	a         fyne.App
	window    fyne.Window
	isRunning bool
	previous  string

	desk desktop.App

	lRadioItems       []*fyne.MenuItem
	sRadioItems       []*fyne.MenuItem
	workgetRadioItems []*fyne.MenuItem
	trItems           []*fyne.MenuItem
	spItems           []*fyne.MenuItem
	startupItems      []*fyne.MenuItem

	gettItems         []*fyne.MenuItem
	currentlSelection *fyne.MenuItem

	currentWorkSelection     *fyne.MenuItem
	getWorkSelection         *fyne.MenuItem
	currentenginetrSelection *fyne.MenuItem
	currentSpetrSelection    *fyne.MenuItem
	curreStartupSelection    *fyne.MenuItem

	gettextSelection *fyne.MenuItem

	textSizeRadioItems       []*fyne.MenuItem
	currentTextSizeSelection *fyne.MenuItem
}

func runOnMain(f func()) {
	fyne.Do(f)
}

const (
	currentVersion = 1
	versionURL     = "https://raw.githubusercontent.com/RezaEjtehadi/spetra/refs/heads/main/updata"
	releasesURL    = "https://github.com/RezaEjtehadi/spetra/releases/latest"
)

func checkVersion() bool {
	resp, err := http.Get(versionURL)
	if err != nil {
		log.Printf("Error getting version information: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		return false
	}

	latestVersionStr := strings.TrimSpace(string(body))
	latestVersion, err := strconv.Atoi(latestVersionStr)
	if err != nil {
		log.Printf("Error converting version: %v\n", err)
		return false
	}

	if latestVersion > currentVersion {
		fmt.Println("New version available! Opening link...")

		if err := openBrowser(releasesURL); err != nil {
			log.Printf("Error opening browser: %v\n", err)
			return false
		}
		return true
	}

	fmt.Println("The program is up to date.")
	return false
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

func splitTextIntoLines(text string, wordsPerLine int) string {
	words := strings.Fields(text)

	var result strings.Builder
	var line strings.Builder
	wordCount := 0

	for i, word := range words {
		if wordCount > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
		wordCount++

		if wordCount >= wordsPerLine || i == len(words)-1 {
			result.WriteString(line.String())
			if i < len(words)-1 {
				result.WriteString("\n")
			}
			line.Reset()
			wordCount = 0
		}
	}

	return result.String()
}

type tappableWidget struct {
	widget.BaseWidget
	onTapped func()
}

func newTappableWidget(onTapped func()) *tappableWidget {
	t := &tappableWidget{onTapped: onTapped}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack())
}

func (t *tappableWidget) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

var lastOverlay fyne.Window

func showOverlay(app fyne.App, translated string, mx int, my int, isLoading bool) {
	if lastOverlay != nil {
		lastOverlay.Close()
	}

	w := app.NewWindow("")
	lastOverlay = w

	bg := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	formattedText := splitTextIntoLines(translated, 15)
	//println(formattedText)
	formattedTextj := "\u2067" + formattedText + "\u2069"

	labelf := widget.NewLabel(formattedTextj)
	//labelf.Alignment = fyne.TextAlignTrailing

	switch viper.GetString("text_size") {
	case "heading":
		labelf.SizeName = theme.SizeNameHeadingText
	case "subheading":
		labelf.SizeName = theme.SizeNameSubHeadingText
	case "text":
		labelf.SizeName = theme.SizeNameText
	case "caption":
		labelf.SizeName = theme.SizeNameCaptionText
	default:
		labelf.SizeName = theme.SizeNameText
	}

	var content *fyne.Container
	if isLoading {
		progress := widget.NewProgressBarInfinite()
		progress.Theme()
		progress.Start()
		content = container.New(
			layout.NewMaxLayout(),
			bg,
			newTappableWidget(func() {
				w.Close()
				if lastOverlay == w {
					lastOverlay = nil
				}
			}),
			container.NewVBox(
				/*container.NewCenter(labelf),*/
				progress,
			),
		)
	} else {
		tappable := newTappableWidget(func() {
			w.Close()
			if lastOverlay == w {
				lastOverlay = nil
			}
		})

		content = container.New(
			layout.NewMaxLayout(),
			bg,
			tappable,
			container.NewCenter(labelf),
		)
	}

	w.SetContent(content)
	w.SetPosition(fyne.Position{X: float32(mx), Y: float32(my)})
	w.SetFixedSize(true)
	w.RequestFocus()
	w.Show()

}

func startup(bb string) {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		return
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		return
	}

	gpp := &autostart.App{
		Name:        "spetra",
		DisplayName: "spetr",
		Exec:        []string{exePath},
	}

	switch bb {
	case "on":
		if enabled := gpp.IsEnabled(); err != nil {
			//log.Printf("Error checking autostart: %v", err)
			return
		} else if enabled {
			println("App is already set to autostart.")
		} else {
			if err := gpp.Enable(); err != nil {
				log.Printf("Error enabling autostart: %v", err)
				return
			}
			println("Autostart has been enabled.")
		}
	case "off":
		if err := gpp.Disable(); err != nil {
			log.Printf("Error disabling autostart: %v", err)
			return
		}
		fmt.Println("disable")

	}

}

func (app *App) setTrayIcon(isActive bool) {
	desk, ok := app.a.(desktop.App)
	if !ok {
		return
	}

	if isActive {
		viper.Set("active", true)

		desk.SetSystemTrayIcon(datastatic.ResourceOnPng)
	} else {
		viper.Set("active", false)

		desk.SetSystemTrayIcon(datastatic.ResourceOffPng)
	}
	if err := viper.WriteConfig(); err != nil {
		log.Printf("Error writing config: %v", err)
	}

	desk.SetSystemTrayMenu(app.buildTrayMenu())
}

func (app *App) startMonitoring() {
	app.isRunning = true

	go func() {
		for app.isRunning {
			typee := viper.GetString("gettexttpye")
			current, err := gettext.GetSelection(typee)

			if err == nil && current != "" && current != app.previous {
				mx, my := robotgo.Location()
				scale := app.window.Canvas().Scale()
				//println(scale)

				if viper.GetString("workget") == "a" || viper.GetString("workget") == "c" {
					go func(text string, mx, my int, scale float32) {
						runOnMain(func() {
							switch runtime.GOOS {
							case "linux":
								showOverlay(app.a, "", int(float32(mx)), int(float32(my)), true)
							case "windows":
								showOverlay(app.a, "", int(float32(mx)/scale), int(float32(my)/scale), true)
							}
						})

						translated, err := engine.TranslateText(text)

						if err == nil {
							runOnMain(func() {
								switch runtime.GOOS {
								case "linux":
									showOverlay(app.a, translated, int(float32(mx)), int(float32(my)), false)
								case "windows":
									showOverlay(app.a, translated, int(float32(mx)/scale), int(float32(my)/scale), false)
								}
							})
						} else {
							runOnMain(func() {
								if lastOverlay != nil {
									lastOverlay.Close()
									lastOverlay = nil
								}
							})
							log.Println("errore ", err)

						}
					}(current, mx, my, scale)
				}

				if viper.GetString("workget") == "a" || viper.GetString("workget") == "b" {
					engine.Speak(current)
				}
			} else if err != nil {
				log.Printf("Error getting selection: %v", err)
			}

			runOnMain(func() {
				app.previous = current
			})

			time.Sleep(300 * time.Millisecond)
		}
	}()
}

func (app *App) stopMonitoring() {
	app.isRunning = false
}

func (app *App) createTrayMenu() {
	if desk, ok := app.a.(desktop.App); ok {
		app.desk = desk

		for name, code := range datastatic.LanguageTranslatorMap {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("translatorlanguage") == langCode)

			item.Action = func() {
				if app.currentlSelection != nil && app.currentlSelection == item {
					return
				}

				if app.currentlSelection != nil {
					app.currentlSelection.Checked = false
				}

				item.Checked = true
				app.currentlSelection = item

				viper.Set("translatorlanguage", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}

				app.updateTrayMenu()
			}

			app.lRadioItems = append(app.lRadioItems, item)
		}

		for name, code := range datastatic.LanguageMapS {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("speechlanguage") == langCode)

			item.Action = func() {
				if app.currentWorkSelection != nil && app.currentWorkSelection == item {
					return
				}

				if app.currentWorkSelection != nil {
					app.currentWorkSelection.Checked = false
				}

				item.Checked = true
				app.currentWorkSelection = item

				viper.Set("speechlanguage", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.sRadioItems = append(app.sRadioItems, item)
		}

		for name, code := range datastatic.Workget {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("workget") == langCode)

			item.Action = func() {
				if app.getWorkSelection != nil && app.getWorkSelection == item {
					return
				}

				if app.getWorkSelection != nil {
					app.getWorkSelection.Checked = false
				}

				item.Checked = true
				app.getWorkSelection = item

				viper.Set("workget", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.workgetRadioItems = append(app.workgetRadioItems, item)
		}

		for name, code := range datastatic.Texttypeget {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("gettexttpye") == langCode)

			item.Action = func() {
				if app.gettextSelection != nil && app.gettextSelection == item {
					return
				}

				if app.gettextSelection != nil {
					app.gettextSelection.Checked = false
				}

				item.Checked = true
				app.gettextSelection = item

				viper.Set("gettexttpye", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.gettItems = append(app.gettItems, item)
		}

		for name, code := range datastatic.EngineTranslate {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("engineTranslate") == langCode)

			item.Action = func() {
				if app.currentenginetrSelection != nil && app.currentenginetrSelection == item {
					return
				}

				if app.currentenginetrSelection != nil {
					app.currentenginetrSelection.Checked = false
				}

				item.Checked = true
				app.currentenginetrSelection = item

				viper.Set("engineTranslate", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.trItems = append(app.trItems, item)
		}

		for name, code := range datastatic.EngineSpeooch {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("engineSpeech") == langCode)

			item.Action = func() {
				if app.currentSpetrSelection != nil && app.currentSpetrSelection == item {
					return
				}

				if app.currentSpetrSelection != nil {
					app.currentSpetrSelection.Checked = false
				}

				item.Checked = true
				app.currentSpetrSelection = item

				viper.Set("engineSpeech", langCode)
				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.spItems = append(app.spItems, item)
		}

		for name, code := range datastatic.StartUp {
			langName := name
			langCode := code

			item := fyne.NewMenuItem(langName, nil)
			item.Checked = (viper.GetString("startup") == langCode)

			item.Action = func() {
				if app.curreStartupSelection != nil && app.curreStartupSelection == item {
					return
				}

				if app.curreStartupSelection != nil {
					app.curreStartupSelection.Checked = false
				}

				item.Checked = true
				app.curreStartupSelection = item

				switch langCode {
				case "on":
					startup("on")
					viper.Set("startup", "on")
				case "off":
					startup("off")
					viper.Set("startup", "off")

				}

				if err := viper.WriteConfig(); err != nil {
					log.Printf("Error writing config: %v", err)
				}
				app.updateTrayMenu()
			}

			app.startupItems = append(app.startupItems, item)
		}

		textSizeMap := map[string]string{
			"Large":  "heading",
			"Medium": "subheading",
			"Normal": "text",
			"Small":  "caption",
		}

		currentTextSize := viper.GetString("text_size")
		if currentTextSize == "" {
			currentTextSize = "text"
			viper.Set("text_size", currentTextSize)
		}

		for label, value := range textSizeMap {
			item := fyne.NewMenuItem(label, nil)
			item.Checked = (currentTextSize == value)

			item.Action = func(val string) func() {
				return func() {
					if app.currentTextSizeSelection != nil && app.currentTextSizeSelection == item {
						return
					}
					if app.currentTextSizeSelection != nil {
						app.currentTextSizeSelection.Checked = false
					}
					item.Checked = true
					app.currentTextSizeSelection = item

					viper.Set("text_size", val)
					if err := viper.WriteConfig(); err != nil {
						log.Printf("Error writing config: %v", err)
					}

					app.updateTrayMenu()
				}
			}(value)

			app.textSizeRadioItems = append(app.textSizeRadioItems, item)
		}

		app.setupInitialMenu()
	}
}

func (app *App) updateTrayMenu() {
	if app.desk == nil {
		println("desk is nil")
		return
	}

	//println("Updating tray menu...")
	menu := app.buildTrayMenu()
	app.desk.SetSystemTrayMenu(menu)
}

func (app *App) buildTrayMenu() *fyne.Menu {
	var toggleText string
	var toggleAction func()
	desk, ok := app.a.(desktop.App)
	if !ok {
		return nil
	}

	isActive := viper.GetBool("active")

	switch isActive {
	case true:
		desk.SetSystemTrayIcon(datastatic.ResourceOnPng)

	case false:
		desk.SetSystemTrayIcon(datastatic.ResourceOffPng)

	}

	if app.isRunning {
		toggleText = "Disable"
		toggleAction = func() {
			app.stopMonitoring()
			app.setTrayIcon(false)
			//fmt.Println("dis")
		}
	} else {
		toggleText = "Enable"
		toggleAction = func() {
			app.startMonitoring()
			app.setTrayIcon(true)
			//fmt.Println("eib")

		}
	}

	return fyne.NewMenu("",
		fyne.NewMenuItem(toggleText, toggleAction),

		fyne.NewMenuItemSeparator(),
		&fyne.MenuItem{
			Label:     "Speech Language",
			ChildMenu: fyne.NewMenu("", app.sRadioItems...),
		},
		&fyne.MenuItem{
			Label:     "Translate Language",
			ChildMenu: fyne.NewMenu("", app.lRadioItems...),
		},
		fyne.NewMenuItemSeparator(),
		&fyne.MenuItem{
			Label:     "Mode",
			ChildMenu: fyne.NewMenu("", app.workgetRadioItems...),
		},
		fyne.NewMenuItemSeparator(),
		&fyne.MenuItem{
			Label:     "Get Text",
			ChildMenu: fyne.NewMenu("", app.gettItems...),
		}, fyne.NewMenuItemSeparator(),

		&fyne.MenuItem{
			Label:     "Engine Translate",
			ChildMenu: fyne.NewMenu("", app.trItems...),
		}, &fyne.MenuItem{
			Label:     "Engine Speech",
			ChildMenu: fyne.NewMenu("", app.spItems...),
		},
		fyne.NewMenuItemSeparator(),
		&fyne.MenuItem{
			Label:     "Text Size",
			ChildMenu: fyne.NewMenu("", app.textSizeRadioItems...),
		},
		fyne.NewMenuItemSeparator(),

		&fyne.MenuItem{
			Label:     "Startup",
			ChildMenu: fyne.NewMenu("", app.startupItems...),
		}, fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About", func() {
			url := "https://github.com/RezaEjtehadi/spetra"

			var cmd string
			var args []string

			switch runtime.GOOS {
			case "windows":
				cmd = "cmd"
				args = []string{"/c", "start", url}
			case "darwin":
				cmd = "open"
				args = []string{url}
			default:
				cmd = "xdg-open"
				args = []string{url}
			}

			err := exec.Command(cmd, args...).Start()
			if err != nil {
				log.Printf("Error opening about: %v", err)
			}

		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			app.stopMonitoring()
			app.a.Quit()
		}),
	)
}

func (app *App) setupInitialMenu() {
	menu := app.buildTrayMenu()
	app.desk.SetSystemTrayMenu(menu)
}

func (app *App) createMainWindow() {
	app.window = app.a.NewWindow(".")
	go func() {
		time.Sleep(time.Second)
		app.window.Close()
	}()
	app.window.Show()
}

func (app *App) SetTheme(theme fyne.Theme) {
	app.a.Settings().SetTheme(theme)
}

func Config() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	standardPaths := getStandardConfigPaths()
	for _, path := range standardPaths {
		viper.AddConfigPath(path)
		fmt.Printf("Searching in path: %s\n", path)
	}

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found. Creating file with default values...")
			setDefaultConfig()

			configDir := getDefaultConfigDir()
			if configDir != "" {
				if err := os.MkdirAll(configDir, 0755); err != nil {
					fmt.Printf("Error creating folder %s: %v\n", configDir, err)
					configDir = "."
				}
			} else {
				configDir = "."
			}

			configPath := filepath.Join(configDir, "config.yaml")
			err = viper.WriteConfigAs(configPath)
			if err != nil {
				return fmt.Errorf("Error creating config file: %s", err)
			}

			fmt.Printf("The config.yaml file was created in the following path: %s\n", configPath)
		} else {
			return fmt.Errorf("Error reading config file: %s", err)
		}
	} else {
		setDefaultConfig()
		fmt.Printf("Config file loaded from the following path: %s\n", viper.ConfigFileUsed())
	}

	fmt.Println("The config file was successfully uploaded.")
	return nil
}

func getStandardConfigPaths() []string {
	var paths []string

	paths = append(paths, ".")

	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			windowsPath := filepath.Join(appData, "spetra")
			paths = append(paths, windowsPath)
		}
	case "linux":
		if home, err := os.UserHomeDir(); err == nil {
			linuxPath := filepath.Join(home, ".config", "spetra")
			paths = append(paths, linuxPath)
		}
	}

	return paths
}

func getDefaultConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "spetra")
		}
	case "linux":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "spetra")
		}
	}
	return "." 
}

func setDefaultConfig() {
	viper.SetDefault("active", true)
	viper.SetDefault("enginespeech", "google")
	viper.SetDefault("enginetranslate", "google")
	viper.SetDefault("gettexttpye", "0")
	viper.SetDefault("speechlanguage", "en")
	viper.SetDefault("translatorlanguage", "fa")
	viper.SetDefault("workget", "c")
	viper.SetDefault("startup", "on")
	viper.SetDefault("text_size", "subheading")
}

func validateConfig() error {
	requiredFields := []string{
		"active",
		"enginespeech",
		"enginetranslate",
		"gettexttpye",
		"speechlanguage",
		"translatorlanguage",
		"workget",
		"text_size",
	}

	for _, field := range requiredFields {
		if !viper.IsSet(field) {
			return fmt.Errorf("Required field '%s' is missing in config", field)
		}
	}

	return nil
}

func main() {
	myApp := &App{a: app.New()}
	myApp.SetTheme(&datastatic.PersianTheme{})
	checkVersion()
	time.Sleep(time.Second)
	err := Config()
	if err != nil {
		log.Printf("Error loading config: %s", err)
	}
	err = validateConfig()
	if err != nil {
		log.Printf("Error validating configuration: %s", err)
	}
	time.Sleep(time.Second)

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
	}

	gpp := &autostart.App{
		Name:        "spetra",
		DisplayName: "spetra",
		Exec:        []string{exePath},
	}

	if viper.GetString("startup") == "on" {
		if enabled := gpp.IsEnabled(); err != nil {
			//log.Printf("Error checking autostart: %v", err)
		} else if enabled {
			println("App is already set to autostart.")
		} else {
			if err := gpp.Enable(); err != nil {
				log.Printf("Error enabling autostart: %v", err)
			}
			println("Autostart has been enabled.")
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config: %v", err)
	}

	time.Sleep(time.Second)
	myApp.createTrayMenu()
	time.Sleep(time.Second)
	myApp.createMainWindow()
	if viper.GetBool("active") {
		myApp.startMonitoring()
	}

	myApp.window.ShowAndRun()
}
