package core

import (
	accordion "github.com/salvatoresaba/simpleui/components/accordion"
	carousel "github.com/salvatoresaba/simpleui/components/carousel"
	chart "github.com/salvatoresaba/simpleui/components/chart"
	combobox "github.com/salvatoresaba/simpleui/components/combobox"
	commandbar "github.com/salvatoresaba/simpleui/components/commandbar"
	datepicker "github.com/salvatoresaba/simpleui/components/datepicker"
	drawer "github.com/salvatoresaba/simpleui/components/drawer"
	menu "github.com/salvatoresaba/simpleui/components/menu"
	modal "github.com/salvatoresaba/simpleui/components/modal"
	overlay "github.com/salvatoresaba/simpleui/components/overlay"
	popover "github.com/salvatoresaba/simpleui/components/popover"
	richtext "github.com/salvatoresaba/simpleui/components/richtext"
	splitter "github.com/salvatoresaba/simpleui/components/splitter"
	stepper "github.com/salvatoresaba/simpleui/components/stepper"
	table "github.com/salvatoresaba/simpleui/components/table"
	tabs "github.com/salvatoresaba/simpleui/components/tabs"
	toast "github.com/salvatoresaba/simpleui/components/toast"
	tooltip "github.com/salvatoresaba/simpleui/components/tooltip"
	tree "github.com/salvatoresaba/simpleui/components/tree"
)

// Component runtimes live beside their control under components/<control>.
// The aliases below keep compiler and root test call sites unchanged while the
// canonical definition is owned by the component package.
const (
	AccordionRuntime  = accordion.Runtime
	CarouselRuntime   = carousel.Runtime
	ChartRuntime      = chart.Runtime
	ComboboxRuntime   = combobox.Runtime
	CommandbarRuntime = commandbar.Runtime
	DatePickerRuntime = datepicker.Runtime
	DrawerRuntime     = drawer.Runtime
	MenuRuntime       = menu.Runtime
	ModalRuntime      = modal.Runtime
	OverlayRuntime    = overlay.Runtime
	PopoverRuntime    = popover.Runtime
	RichTextRuntime   = richtext.Runtime
	RichTextCSS       = richtext.CSS
	SplitterRuntime   = splitter.Runtime
	StepperRuntime    = stepper.Runtime
	TableRuntime      = table.Runtime
	TabsRuntime       = tabs.Runtime
	ToastRuntime      = toast.Runtime
	TooltipRuntime    = tooltip.Runtime
	TreeRuntime       = tree.Runtime

	suiAccordionRuntime  = AccordionRuntime
	suiCarouselRuntime   = CarouselRuntime
	suiChartRuntime      = ChartRuntime
	suiComboboxRuntime   = ComboboxRuntime
	suiCommandbarRuntime = CommandbarRuntime
	suiDatePickerRuntime = DatePickerRuntime
	suiDrawerRuntime     = DrawerRuntime
	suiMenuRuntime       = MenuRuntime
	suiModalRuntime      = ModalRuntime
	suiOverlayRuntime    = OverlayRuntime
	suiPopoverRuntime    = PopoverRuntime
	suiRichTextRuntime   = RichTextRuntime
	suiRichTextCss       = RichTextCSS
	suiSplitterRuntime   = SplitterRuntime
	suiStepperRuntime    = StepperRuntime
	suiTableRuntime      = TableRuntime
	suiTabsRuntime       = TabsRuntime
	suiToastRuntime      = ToastRuntime
	suiTooltipRuntime    = TooltipRuntime
	suiTreeRuntime       = TreeRuntime
)

// Rich text build helpers are exported as well so component tests in an
// external test package can exercise subsetting and toolbar metadata.
const (
	RichTextRuntimeMarker = richtext.RuntimeMarker
	HyphenPatternsStart   = richtext.HyphenPatternsStart
	HyphenPatternsEnd     = richtext.HyphenPatternsEnd

	richTextRuntimeMarker = RichTextRuntimeMarker
	hyphenPatternsStart   = HyphenPatternsStart
	hyphenPatternsEnd     = HyphenPatternsEnd
)

type (
	HyphenLangOption     = richtext.HyphenLangOption
	RichTextToolbarGroup = richtext.ToolbarGroup

	hyphenLangOption     = HyphenLangOption
	richTextToolbarGroup = RichTextToolbarGroup
)

var (
	HyphenPatterns = richtext.HyphenPatterns
	HyphenOptions  = richtext.HyphenOptions

	richHyphenPatterns = HyphenPatterns
	richHyphenOptions  = HyphenOptions
)

func RichTextRuntimeSubset(codes []string) string { return richtext.RuntimeSubset(codes) }
func HyphenPatternCodes(codes []string) []string  { return richtext.HyphenPatternCodes(codes) }
func HyphenLanguagesJSON() string                 { return richtext.HyphenLanguagesJSON() }
func RichTextToolbarGroups() []RichTextToolbarGroup {
	return richtext.ToolbarGroups()
}
func RichTextToolbarJSON() string { return richtext.ToolbarJSON() }

func richTextRuntimeSubset(codes []string) string { return RichTextRuntimeSubset(codes) }
func hyphenPatternCodes(codes []string) []string  { return HyphenPatternCodes(codes) }
func hyphenLanguagesJSON() string                 { return HyphenLanguagesJSON() }
func richTextToolbarGroups() []richTextToolbarGroup {
	return RichTextToolbarGroups()
}
func richTextToolbarJSON() string { return RichTextToolbarJSON() }
