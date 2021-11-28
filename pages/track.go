package pages

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	uniswap "github.com/hirokimoto/uniswap-api"
	unitrade "github.com/hirokimoto/uniswap-api/swap"
	unitrades "github.com/hirokimoto/uniswap-api/swaps"
	"github.com/uniswap-auto-gui/data"
	"github.com/uniswap-auto-gui/services"
)

func trackScreen(_ fyne.Window) fyne.CanvasObject {
	var selected uniswap.Swaps
	var activePair string

	pairs := data.ReadTrackPairs()
	records, _ := data.ReadTrackSettings()
	oldPrices := make([]float64, 0)
	oldTimes := make([]int64, 0)

	for _, _ = range pairs {
		oldPrices = append(oldPrices, 0.0)
		oldTimes = append(oldTimes, 1638118581)
	}

	name := widget.NewEntry()
	name.SetPlaceHolder("0x385769E84B650C070964398929DB67250B7ff72C")
	append := widget.NewButton("Append", func() {
		if name.Text != "" {
			isExisted := false
			for _, item := range pairs {
				if item == name.Text {
					isExisted = true
				}
			}
			if !isExisted {
				pairs = append(pairs, name.Text)
				oldPrices = append(oldPrices, 0.0)
				oldTimes = append(oldTimes, 1638118581)
				data.SaveTrackPairs(pairs)
			}
		}
	})

	control := container.NewVBox(name, append)

	rightList := widget.NewList(
		func() int {
			return len(selected.Data.Swaps)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("target"), widget.NewLabel("price"), widget.NewLabel("amount"), widget.NewLabel("amount1"), widget.NewLabel("amount2"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			price, target, amount, amount1, amount2 := unitrade.Trade(selected.Data.Swaps[id])

			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(target)
			item.(*fyne.Container).Objects[2].(*widget.Label).SetText(fmt.Sprintf("$%f", price))
			item.(*fyne.Container).Objects[3].(*widget.Label).SetText(amount)
			item.(*fyne.Container).Objects[4].(*widget.Label).SetText(amount1)
			item.(*fyne.Container).Objects[5].(*widget.Label).SetText(amount2)
		},
	)

	table := widget.NewTable(
		func() (int, int) { return len(pairs), 5 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			go func() {
				for {
					fmt.Print(".")
					pair := pairs[id.Row]

					var swaps uniswap.Swaps
					cc := make(chan string, 1)
					go uniswap.SwapsByCounts(cc, 2, pair)

					msg := <-cc
					json.Unmarshal([]byte(msg), &swaps)

					if len(swaps.Data.Swaps) == 0 || swaps.Data.Swaps == nil {
						time.Sleep(time.Second * 1)
						continue
					}

					n := unitrade.Name(swaps.Data.Swaps[0])
					p, _ := unitrade.Price(swaps.Data.Swaps[0])
					_, c := unitrades.WholePriceChanges(swaps)
					_, _, d := unitrades.Duration(swaps)

					current, _ := strconv.ParseInt(swaps.Data.Swaps[0].Timestamp, 10, 64)
					if oldPrices[id.Row] != p && oldPrices[id.Row] != 0.0 && current > oldTimes[id.Row] {
						alert(records, pair, n, p, c, d)
						oldPrices[id.Row] = p
						oldTimes[id.Row] = current
					}

					if pair == activePair {
						selected = swaps
						rightList.Refresh()
					}

					switch id.Col {
					case 0:
						if label.Text == "" {
							label.SetText(fmt.Sprintf("%d", id.Row+1))
						}
					case 1:
						if label.Text == "" {
							if len(n) > 20 {
								label.SetText(n[0:20] + "...")
							} else {
								label.SetText(n)
							}
						}
					case 2:
						if label.Text != fmt.Sprintf("%f", p) {
							label.SetText(fmt.Sprintf("%f", p))
						}
					case 3:
						if label.Text != fmt.Sprintf("%f", c) {
							label.SetText(fmt.Sprintf("%f", c))
						}
					case 4:
						if label.Text != fmt.Sprintf("%f", d) {
							label.SetText(fmt.Sprintf("%f", d))
						}
					default:
					}
					time.Sleep(time.Second * 1)
				}
			}()
		})
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 202)
	table.SetColumnWidth(2, 100)
	table.SetColumnWidth(3, 100)
	table.SetColumnWidth(4, 100)
	table.OnSelected = func(id widget.TableCellID) {
		pair := pairs[id.Row]
		if id.Col == 1 {
			activePair = pair
		}
		if id.Col == 2 {
			w := fyne.CurrentApp().NewWindow("Settings")

			min, max := data.ReadMinMax(records, pair)
			mindata := binding.BindFloat(&min)
			minLabel := widget.NewLabel("Minimum")
			minEntry := widget.NewEntryWithData(binding.FloatToString(mindata))
			minPanel := container.NewGridWithColumns(2, minLabel, minEntry)

			maxdata := binding.BindFloat(&max)
			maxLabel := widget.NewLabel("Maximum")
			maxEntry := widget.NewEntryWithData(binding.FloatToString(maxdata))
			maxPanel := container.NewGridWithColumns(2, maxLabel, maxEntry)

			btnSave := widget.NewButton("Save", func() {
				data.SaveTrackSettings(pair, min, max)
			})

			settingsPanel := container.NewVBox(minPanel, maxPanel, btnSave)
			w.SetContent(settingsPanel)

			w.Resize(fyne.NewSize(340, 180))
			w.SetFixedSize(true)
			w.Show()
		}
	}

	listPanel := container.NewBorder(nil, control, nil, nil, table)

	return container.NewHSplit(listPanel, rightList)
}

func alert(records [][]string, pair string, n string, p float64, c float64, d float64) {
	message := fmt.Sprintf("%s: %f %f %f", n, p, c, d)
	title := "Priced Up!"
	if c < 0 {
		title = "Priced Down!"
	}
	link := fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair)

	min, max := data.ReadMinMax(records, pair)

	if p < min {
		title = fmt.Sprintf("Warning Low! Watch %s", n)
	}
	if p > max {
		title = fmt.Sprintf("Warning High! Watch %s", n)
	}

	services.Alert(title, message, link, gosxnotifier.Morse)
}
