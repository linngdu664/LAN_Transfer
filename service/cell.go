package service

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"time"
)

var (
	MainApp    fyne.App
	MainWindow fyne.Window

	Tabs *container.AppTabs

	SIpInput             *widget.Entry
	RIpInput             *widget.Entry
	SenderPortInput      *widget.Entry
	SenderFileSrcInput   *widget.Entry
	ReceiverPortInput    *widget.Entry
	ReceiverFileSrcInput *widget.Entry

	RList      *widget.List
	RListItems []string
	SList      *widget.List
	SListItems []string

	SenderFileSelectBtn   *widget.Button
	SearchIpBtn           *widget.Button
	SendFileBtn           *widget.Button
	ReceiverFileSelectBtn *widget.Button
	ReceiverSwitch        *widget.RadioGroup

	SenderProgressBar   *widget.ProgressBar
	ReceiverProgressBar *widget.ProgressBar
	SenderSpeedText     *canvas.Text
	ReceiverSpeedText   *canvas.Text

	Logger    *widget.Label
	LogScroll *container.Scroll

	SenderFileDialog   *dialog.FileDialog
	ReceiverFileDialog *dialog.FileDialog
)

func InitWidget() {
	MainApp = app.New()
	MainWindow = MainApp.NewWindow("main")
	MainWindow.Resize(fyne.Size{
		Width:  500,
		Height: 300,
	})

	Logger = widget.NewLabel("")

	SIpInput = widget.NewEntry()
	SIpInput.SetPlaceHolder("Target ip address")
	RIpInput = widget.NewEntry()
	RIpInput.SetPlaceHolder("Target lan address")
	SenderPortInput = widget.NewEntry()
	SenderPortInput.SetPlaceHolder("Target port")
	SenderFileSrcInput = widget.NewEntry()
	SenderFileSrcInput.SetPlaceHolder("src of the file to be sent")
	ReceiverPortInput = widget.NewEntry()
	ReceiverPortInput.SetPlaceHolder("Local port")
	ReceiverFileSrcInput = widget.NewEntry()
	ReceiverFileSrcInput.SetPlaceHolder("src of the folder to be received")

	RList = widget.NewList(
		func() int { return 1 },
		func() fyne.CanvasObject {
			button := widget.NewButton("", nil)
			button.OnTapped = func() {
				RIpInput.SetText(button.Text)
			}
			return button
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
		})
	SList = widget.NewList(
		func() int { return 1 },
		func() fyne.CanvasObject {
			button := widget.NewButton("", nil)
			button.OnTapped = func() {
				SIpInput.SetText(button.Text)
			}
			return button
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
		})

	SenderFileDialog = dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
		if err != nil {
			LogErr("Sender dialog error" + err.Error())
			return
		}
		if closer != nil {
			SenderFileSrcInput.SetText(closer.URI().Path())
		}
	}, MainWindow)
	ReceiverFileDialog = dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			LogErr("Receiver dialog error" + err.Error())
			return
		}
		if uri != nil {
			ReceiverFileSrcInput.SetText(uri.Path())
		}
	}, MainWindow)

	SenderFileSelectBtn = widget.NewButton("Browser", func() {
		SenderFileDialog.Show()
	})
	SearchIpBtn = widget.NewButton("Search IP", nil)
	SendFileBtn = widget.NewButton("Send File", func() {
		Sender.SendFile()
	})
	ReceiverFileSelectBtn = widget.NewButton("Browser", func() {
		ReceiverFileDialog.Show()
	})
	ReceiverSwitch = widget.NewRadioGroup([]string{"Receive Enable", "Receive Disable"}, nil)
	ReceiverSwitch.SetSelected("Receive Disable")
	ReceiverSwitch.OnChanged = func(s string) {
		if s == "Receive Enable" {
			err := Receiver.Run()
			if err != nil {
				ReceiverSwitch.SetSelected("Receive Disable")
				return
			}
		} else {
			Receiver.Stop()
		}
	}

	SenderProgressBar = widget.NewProgressBar()
	ReceiverProgressBar = widget.NewProgressBar()
	SenderSpeedText = canvas.NewText("  0.0B/s t:0s", color.Black)
	ReceiverSpeedText = canvas.NewText("  0.0B/s t:0s", color.Black)

	Tabs = container.NewAppTabs(
		container.NewTabItem("Sender",
			container.NewGridWithColumns(2,
				container.NewGridWithRows(2,
					SIpInput,
					SList,
				),
				container.NewGridWithRows(4,
					container.NewGridWithColumns(2,
						SenderPortInput,
						SenderFileSelectBtn,
					),
					SenderFileSrcInput,
					container.NewGridWithColumns(2,
						SearchIpBtn,
						SendFileBtn,
					),
					container.NewStack(SenderProgressBar, SenderSpeedText),
				),
			),
		),
		container.NewTabItem("Receiver",
			container.NewGridWithColumns(2,
				container.NewGridWithRows(2,
					RIpInput,
					RList,
				),
				container.NewGridWithRows(4,
					container.NewGridWithColumns(2,
						ReceiverPortInput,
						ReceiverFileSelectBtn,
					),
					ReceiverFileSrcInput,
					ReceiverSwitch,
					container.NewStack(ReceiverProgressBar, ReceiverSpeedText),
				),
			),
		),
	)
	Tabs.OnSelected = func(item *container.TabItem) {
		if item.Text == "Receiver" {
			Sender.StopIpSearcher()
		} else {
			err := Sender.RunIpSearcher()
			if err != nil {
				ReceiverSwitch.SetSelected("Receive Enable")
				return
			}
		}
	}
	LogScroll = container.NewScroll(Logger)
	box := container.NewGridWithRows(2,
		Tabs,
		LogScroll,
	)
	MainWindow.SetContent(box)
}
func RefreshRList(list []string) {
	RListItems = list
	RList.Length = func() int { return len(RListItems) }
	RList.UpdateItem = func(id widget.ListItemID, object fyne.CanvasObject) {
		object.(*widget.Button).SetText(RListItems[id])
	}
	RList.Refresh()

}
func AddSList(ip string) bool {
	for _, item := range SListItems {
		if item == ip {
			return false
		}
	}
	SListItems = append(SListItems, ip)
	SList.Length = func() int { return len(SListItems) }
	SList.UpdateItem = func(id widget.ListItemID, object fyne.CanvasObject) {
		object.(*widget.Button).SetText(SListItems[id])
	}
	SList.Refresh()
	if SIpInput.Text == "" {
		SIpInput.SetText(SListItems[0])
	}
	return true
}
func Log(msg string) {
	formatMsg := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	fmt.Print(formatMsg)
	Logger.SetText(Logger.Text + formatMsg)
	LogScroll.ScrollToBottom()
}
func LogErr(msg string) {
	formatMsg := fmt.Sprintf("[%s] [Error] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	fmt.Print(formatMsg)
	Logger.SetText(Logger.Text + formatMsg)
	LogScroll.ScrollToBottom()
}
