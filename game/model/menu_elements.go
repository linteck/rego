package model

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func titleBarContainer(menu *DemoMenu) *widget.Container {
	res := menu.res

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(3),
				widget.GridLayoutOpts.Stretch([]bool{true, false, false}, []bool{true}),
				widget.GridLayoutOpts.Padding(&widget.Insets{
					Left:   30,
					Right:  5,
					Top:    6,
					Bottom: 5,
				}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Raycaster Go Demo Settings", &res.text.titleFace, res.textInput.color.Idle),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c.AddChild(widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(&res.button.padding),
		widget.ButtonOpts.Text("X", &res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			menu.game.closeMenu()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	return c
}

func footerContainer(menu *DemoMenu) *widget.Container {
	res := menu.res

	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(&widget.Insets{
			Left:  25,
			Right: 25,
		}),
	)))
	c.AddChild(widget.NewText(
		widget.TextOpts.Text("://github.com", &res.text.smallFace, res.text.disabledColor)))
	return c
}

func settingsContainer(menu *DemoMenu) widget.PreferredSizeLocateableWidget {
	res := menu.res

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0),
		)))

	pages := []interface{}{
		gamePage(menu),
		displayPage(menu),
		renderPage(menu),
		lightingPage(menu),
	}

	pageContainer := newPageContainer(res)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*page).title
		}),
		widget.ListOpts.ScrollContainerImage(res.list.image),
		widget.ListOpts.SliderParams(&widget.SliderParams{
			TrackImage:    res.list.track,
			HandleImage:   res.list.handle,
			MinHandleSize: &res.list.handleSize,
			TrackPadding:  &res.list.trackPadding,
		}),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(&res.list.face),
		widget.ListOpts.EntryTextPadding(&res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			pageContainer.setPage(args.Entry.(*page))
			menu.root.RequestRelayout()
		}))
	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
}

func newCheckbox(label string, checked bool, changedHandler widget.CheckboxChangedHandlerFunc, res *uiResources) *widget.Checkbox {
	c := widget.NewCheckbox(
		widget.CheckboxOpts.WidgetOpts(
			// Set the location of the checkbox
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.CheckboxOpts.Image(res.checkbox.image),
		widget.CheckboxOpts.Text(label, &res.label.face, res.label.text),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			if changedHandler != nil {
				changedHandler(args)
			}
		}),
	)

	if checked {
		c.SetState(widget.WidgetChecked)
	}

	return c
}

func newListComboButton(entries []interface{}, selectedEntry interface{}, buttonLabel widget.SelectComboButtonEntryLabelFunc, entryLabel widget.ListEntryLabelFunc,
	entrySelectedHandler widget.ListComboButtonEntrySelectedHandlerFunc, res *uiResources,
) *widget.ListComboButton {
	c := widget.NewListComboButton(
		widget.ListComboButtonOpts.ButtonParams(&widget.ButtonParams{
			Image:       res.comboButton.image,
			TextPadding: &res.comboButton.padding,
		}),
		widget.ListComboButtonOpts.Text(&res.comboButton.face, res.comboButton.graphic, res.comboButton.text),
		// Provide your entry data array
		widget.ListComboButtonOpts.Entries(entries),

		// Configure the main button style and text properties
		widget.ListComboButtonOpts.ButtonParams(&widget.ButtonParams{
			Image:       res.comboButton.image,
			TextPadding: &res.comboButton.padding,
			TextFace:    &res.comboButton.face,
			// Graphic and primary text colors can be mapped to ButtonTextColor if needed
			TextColor: &widget.ButtonTextColor{
				Idle:     res.comboButton.text.Idle,     // Adjust to your color struct property names
				Disabled: res.comboButton.text.Disabled, // Adjust to your color struct property names
			},
		}),

		// Configure the complete list dropdown dropdown style flatly
		widget.ListComboButtonOpts.ListParams(&widget.ListParams{
			ScrollContainerImage: &widget.ScrollContainerImage{
				Idle:     res.list.image.Idle,
				Disabled: res.list.image.Disabled,
				Mask:     res.list.image.Mask,
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{
					Idle:  res.list.track.Idle,
					Hover: res.list.track.Hover,
				},
				HandleImage:   res.list.handle,
				MinHandleSize: &res.list.handleSize,
				TrackPadding:  &res.list.trackPadding,
			},
			EntryFace:        &res.list.face,
			EntryColor:       res.list.entry, // Direct assignment of your *widget.ListEntryColor struct
			EntryTextPadding: &res.list.entryPadding,
		}),
		widget.ListComboButtonOpts.EntryLabelFunc(buttonLabel, entryLabel),
		widget.ListComboButtonOpts.EntrySelectedHandler(entrySelectedHandler))

	if selectedEntry != nil {
		c.SetSelectedEntry(selectedEntry)
	}

	return c
}

func newSeparator(res *uiResources, ld interface{}) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Bottom: 20,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.separatorColor)),
	))

	return c
}
