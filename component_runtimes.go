package sui

import (
	accordion "example.com/simpleui/components/accordion"
	carousel "example.com/simpleui/components/carousel"
	chart "example.com/simpleui/components/chart"
	combobox "example.com/simpleui/components/combobox"
	commandbar "example.com/simpleui/components/commandbar"
	datepicker "example.com/simpleui/components/datepicker"
	drawer "example.com/simpleui/components/drawer"
	menu "example.com/simpleui/components/menu"
	modal "example.com/simpleui/components/modal"
	overlay "example.com/simpleui/components/overlay"
	popover "example.com/simpleui/components/popover"
	richtext "example.com/simpleui/components/richtext"
	splitter "example.com/simpleui/components/splitter"
	stepper "example.com/simpleui/components/stepper"
	table "example.com/simpleui/components/table"
	tabs "example.com/simpleui/components/tabs"
	toast "example.com/simpleui/components/toast"
	tooltip "example.com/simpleui/components/tooltip"
	tree "example.com/simpleui/components/tree"
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
