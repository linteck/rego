package model

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type pageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type page struct {
	title   string
	content widget.PreferredSizeLocateableWidget
}

func gamePage(menu *DemoMenu) *page {
	c := newPageContentContainer()
	res := menu.res

	resume := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Resume", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { menu.game.closeMenu() }),
	)
	c.AddChild(resume)

	if menu.game.cfg.OsType == OsTypeBrowser {
		// exit in browser kills but freezes the application, users can just close the tab/window
	} else {
		// show in game exit button
		c.AddChild(newSeparator(res, widget.RowLayoutData{
			Stretch: true,
		}))

		exit := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Text("Exit", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { exit(0) }),
		)
		c.AddChild(exit)
	}

	return &page{
		title:   "Game",
		content: c,
	}
}

func displayPage(menu *DemoMenu) *page {
	c := newPageContentContainer()
	res := menu.res

	// resolution combo box and label
	resolutionRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(resolutionRow)

	resolutionLabel := widget.NewLabel(widget.LabelOpts.Text("Resolution", res.label.face, res.label.text))
	resolutionRow.AddChild(resolutionLabel)

	resolutions := []interface{}{}
	var selectedResolution interface{}
	for _, r := range menu.resolutions {
		resolutions = append(resolutions, r)
		if menu.game.cfg.ScreenWidth == r.width && menu.game.cfg.ScreenHeight == r.height {
			selectedResolution = r
		}
	}

	if selectedResolution == nil {
		// generate custom entry to put at top of the list
		r := MenuResolution{
			width:  menu.game.cfg.ScreenWidth,
			height: menu.game.cfg.ScreenHeight,
		}
		resolutions = append([]interface{}{r}, resolutions...)
	}

	var fovSlider *widget.Slider
	resolutionCombo := newListComboButton(
		resolutions,
		selectedResolution,
		func(e interface{}) string {
			return fmt.Sprintf("%s", e)
		},
		func(e interface{}) string {
			return fmt.Sprintf("%s", e)
		},
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			r := args.Entry.(MenuResolution)
			if menu.game.cfg.ScreenWidth != r.width || menu.game.cfg.ScreenHeight != r.height {
				menu.game.cfg.ScreenWidth = r.width
				menu.game.cfg.ScreenHeight = r.height

				// also pre-select ideal FOV for the aspect ratio
				fovSlider.Current = r.aspectRatio.fov
			}
		},
		res)
	resolutionRow.AddChild(resolutionCombo)

	// horizontal FOV slider
	fovRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(fovRow)

	fovLabel := widget.NewLabel(widget.LabelOpts.Text("Horizontal FOV", res.label.face, res.label.text))
	fovRow.AddChild(fovLabel)

	var fovValueText *widget.Label

	fovSlider = widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(200, 6)),
		widget.SliderOpts.MinMax(60, 120),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			fovValueText.Label = fmt.Sprintf("%d", args.Current)
			menu.game.cfg.FovDegrees = float64(args.Current)
		}),
	)
	fovSlider.Current = int(menu.game.cfg.FovDegrees)
	fovRow.AddChild(fovSlider)

	fovValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", fovSlider.Current), res.label.face, res.label.text),
	)
	fovRow.AddChild(fovValueText)

	// render scaling combo box
	scalingRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(scalingRow)

	scalingLabel := widget.NewLabel(widget.LabelOpts.Text("Render Scaling", res.label.face, res.label.text))
	scalingRow.AddChild(scalingLabel)

	scalings := []interface{}{
		0.25,
		0.5,
		0.75,
		1.0,
	}

	var selectedScaling interface{}
	for _, s := range scalings {
		if s == menu.game.cfg.RenderScale {
			selectedScaling = s
		}
	}

	scalingCombo := newListComboButton(
		scalings,
		selectedScaling,
		func(e interface{}) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(e interface{}) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			s := args.Entry.(float64)
			menu.game.cfg.RenderScale = s
		},
		res)
	scalingRow.AddChild(scalingCombo)

	// fullscreen checkbox
	fsCheckbox := newCheckbox("Fullscreen", menu.game.cfg.Fullscreen,
		func(args *widget.CheckboxChangedEventArgs) {
			menu.game.cfg.Fullscreen = (args.State == widget.WidgetChecked)
		}, res)
	c.AddChild(fsCheckbox)

	// vsync checkbox
	vsCheckbox := newCheckbox("Use VSync", menu.game.cfg.Vsync,
		func(args *widget.CheckboxChangedEventArgs) {
			menu.game.cfg.Vsync = (args.State == widget.WidgetChecked)
		}, res)
	c.AddChild(vsCheckbox)

	return &page{
		title:   "Display",
		content: c,
	}
}

func renderPage(menu *DemoMenu) *page {
	c := newPageContentContainer()
	res := menu.res

	// render distance slider
	distanceRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(distanceRow)

	distanceLabel := widget.NewLabel(widget.LabelOpts.Text("Render float64",
		res.label.face, res.label.text))
	distanceRow.AddChild(distanceLabel)

	var distanceValueText *widget.Label

	distanceSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(200, 6)),
		widget.SliderOpts.MinMax(-1, 100),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			distanceValueText.Label = fmt.Sprintf("%d", args.Current)
			menu.game.cfg.RenderDistance = float64(args.Current)
		}),
	)
	distanceSlider.Current = int(menu.game.cfg.RenderDistance)
	distanceRow.AddChild(distanceSlider)

	distanceValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", distanceSlider.Current), res.label.face, res.label.text),
	)
	distanceRow.AddChild(distanceValueText)

	// floor texturing checkbox
	floorCheckbox := newCheckbox("Floor Texturing", menu.game.cfg.RenderFloorTex,
		func(args *widget.CheckboxChangedEventArgs) {
			menu.game.cfg.RenderFloorTex = args.State == widget.WidgetChecked
		}, res)
	c.AddChild(floorCheckbox)

	// sprite boxes checkbox
	spriteBoxCheckbox := newCheckbox("Sprite Boxes", menu.game.cfg.ShowSpriteBoxes,
		func(args *widget.CheckboxChangedEventArgs) {
			menu.game.cfg.ShowSpriteBoxes = args.State == widget.WidgetChecked
		}, res)
	c.AddChild(spriteBoxCheckbox)

	return &page{
		title:   "Render",
		content: c,
	}
}

func lightingPage(menu *DemoMenu) *page {
	c := newPageContentContainer()
	res := menu.res

	// light falloff slider
	falloffRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(falloffRow)

	falloffLabel := widget.NewLabel(widget.LabelOpts.Text("Light Falloff", res.label.face, res.label.text))
	falloffRow.AddChild(falloffLabel)

	var falloffValueText *widget.Label

	falloffSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(200, 6)),
		widget.SliderOpts.MinMax(-500, 500),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			falloffValueText.Label = fmt.Sprintf("%d", args.Current)
			menu.game.cfg.LightFalloff = float64(args.Current)
		}),
	)
	falloffSlider.Current = int(menu.game.cfg.LightFalloff)
	falloffRow.AddChild(falloffSlider)

	falloffValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", falloffSlider.Current), res.label.face, res.label.text),
	)
	falloffRow.AddChild(falloffValueText)

	// global illumination slider
	globalRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(globalRow)

	globalLabel := widget.NewLabel(widget.LabelOpts.Text("Illumination", res.label.face, res.label.text))
	globalRow.AddChild(globalLabel)

	var globalValueText *widget.Label

	globalSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(200, 6)),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			globalValueText.Label = fmt.Sprintf("%d", args.Current)
			menu.game.cfg.GlobalIllumination = float64(args.Current)
		}),
	)
	globalSlider.Current = int(menu.game.cfg.GlobalIllumination)
	globalRow.AddChild(globalSlider)

	globalValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", globalSlider.Current), res.label.face, res.label.text),
	)
	globalRow.AddChild(globalValueText)

	// min lighting RGB selection
	minGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, nil),
			widget.GridLayoutOpts.Spacing(5, 5))))
	c.AddChild(minGrid)

	minLabel := widget.NewLabel(widget.LabelOpts.Text("Min Light", res.label.face, res.label.text))
	var rMinText, gMinText, bMinText *widget.Label
	var rgbMinValue *widget.Container

	rMinSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			rMinText.Label = fmt.Sprintf("R: %d", args.Current)
			rgb := menu.game.cfg.MinLightRGB
			menu.game.cfg.MinLightRGB = color.NRGBA{R: uint8(args.Current), G: rgb.G, B: rgb.B, A: 255}
			rgbMinValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MinLightRGB)
		}),
	)
	rMinSlider.Current = int(menu.game.cfg.MinLightRGB.R)

	gMinSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			gMinText.Label = fmt.Sprintf("G: %d", args.Current)
			rgb := menu.game.cfg.MinLightRGB
			menu.game.cfg.MinLightRGB = color.NRGBA{R: rgb.R, G: uint8(args.Current), B: rgb.B, A: 255}
			rgbMinValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MinLightRGB)
		}),
	)
	gMinSlider.Current = int(menu.game.cfg.MinLightRGB.G)

	bMinSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			bMinText.Label = fmt.Sprintf("B: %d", args.Current)
			rgb := menu.game.cfg.MinLightRGB
			menu.game.cfg.MinLightRGB = color.NRGBA{R: rgb.R, G: rgb.G, B: uint8(args.Current), A: 255}
			rgbMinValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MinLightRGB)
		}),
	)
	bMinSlider.Current = int(menu.game.cfg.MinLightRGB.B)

	rMinText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("R: %d", rMinSlider.Current), res.label.face, res.label.text))
	gMinText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("G: %d", gMinSlider.Current), res.label.face, res.label.text))
	bMinText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("B: %d", bMinSlider.Current), res.label.face, res.label.text))

	rgbMinBackground := image.NewNineSliceColor(menu.game.cfg.MinLightRGB)
	rgbMinValue = widget.NewContainer(widget.ContainerOpts.BackgroundImage(rgbMinBackground))

	// min RGB row 1
	minGrid.AddChild(minLabel)
	minGrid.AddChild(rMinText)
	minGrid.AddChild(gMinText)
	minGrid.AddChild(bMinText)

	// min RGB row 2
	minGrid.AddChild(rgbMinValue)
	minGrid.AddChild(rMinSlider)
	minGrid.AddChild(gMinSlider)
	minGrid.AddChild(bMinSlider)

	// max lighting RGB selection
	maxGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, nil),
			widget.GridLayoutOpts.Spacing(5, 5))))
	c.AddChild(maxGrid)

	maxLabel := widget.NewLabel(widget.LabelOpts.Text("Max Light", res.label.face, res.label.text))
	var rMaxText, gMaxText, bMaxText *widget.Label
	var rgbMaxValue *widget.Container

	rMaxSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			rMaxText.Label = fmt.Sprintf("R: %d", args.Current)
			rgb := menu.game.cfg.MaxLightRGB
			menu.game.cfg.MaxLightRGB = color.NRGBA{R: uint8(args.Current), G: rgb.G, B: rgb.B, A: 255}
			rgbMaxValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MaxLightRGB)
		}),
	)
	rMaxSlider.Current = int(menu.game.cfg.MaxLightRGB.R)

	gMaxSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			gMaxText.Label = fmt.Sprintf("G: %d", args.Current)
			rgb := menu.game.cfg.MaxLightRGB
			menu.game.cfg.MaxLightRGB = color.NRGBA{R: rgb.R, G: uint8(args.Current), B: rgb.B, A: 255}
			rgbMaxValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MaxLightRGB)
		}),
	)
	gMaxSlider.Current = int(menu.game.cfg.MaxLightRGB.G)

	bMaxSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 8)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			bMaxText.Label = fmt.Sprintf("B: %d", args.Current)
			rgb := menu.game.cfg.MaxLightRGB
			menu.game.cfg.MaxLightRGB = color.NRGBA{R: rgb.R, G: rgb.G, B: uint8(args.Current), A: 255}
			rgbMaxValue.BackgroundImage = image.NewNineSliceColor(menu.game.cfg.MaxLightRGB)
		}),
	)
	bMaxSlider.Current = int(menu.game.cfg.MaxLightRGB.B)

	rMaxText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("R: %d", rMaxSlider.Current),
		res.label.face, res.label.text))
	gMaxText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("G: %d", gMaxSlider.Current),
		res.label.face, res.label.text))
	bMaxText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("B: %d", bMaxSlider.Current),
		res.label.face, res.label.text))

	rgbMaxBackground := image.NewNineSliceColor(menu.game.cfg.MaxLightRGB)
	rgbMaxValue = widget.NewContainer(widget.ContainerOpts.BackgroundImage(rgbMaxBackground))

	// max RGB row 1
	maxGrid.AddChild(maxLabel)
	maxGrid.AddChild(rMaxText)
	maxGrid.AddChild(gMaxText)
	maxGrid.AddChild(bMaxText)

	// max RGB row 2
	maxGrid.AddChild(rgbMaxValue)
	maxGrid.AddChild(rMaxSlider)
	maxGrid.AddChild(gMaxSlider)
	maxGrid.AddChild(bMaxSlider)

	return &page{
		title:   "Lighting",
		content: c,
	}
}

func newPageContentContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
}

func newPageContainer(res *uiResources) *pageContainer {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(15))),
	)

	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &pageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *pageContainer) setPage(page *page) {
	p.titleText.Label = page.title
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func exit(rc int) {
	// TODO: any cleanup?
	os.Exit(rc)
}
