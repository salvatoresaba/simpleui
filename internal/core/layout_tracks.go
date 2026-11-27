package core

import (
	"regexp"
	"strconv"
	"strings"
)

// layoutTrackAuto is the CSS track of a cell whose size is not expressed: the
// cell takes the space that the other cells leave free.
const layoutTrackAuto = "minmax(0,1fr)"

var (
	layoutTrackNumber  = regexp.MustCompile(`^\d+(\.\d+)?$`)
	layoutTrackPercent = regexp.MustCompile(`^\d+(\.\d+)?%$`)
	layoutTrackPixels  = regexp.MustCompile(`^\d+(\.\d+)?px$`)
)

// layoutTrackValue parses one size of a track list.
//
// An empty value, "auto" and "1fr" mean automatic: the cell takes the space
// left free by the other cells. A bare number is a size in pixels, a
// percentage is a share of the container. The second result reports whether
// the value is valid.
func layoutTrackValue(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" || v == "auto" || v == "1fr" {
		return layoutTrackAuto, true
	}
	if layoutTrackNumber.MatchString(v) {
		return v + "px", true
	}
	if layoutTrackPixels.MatchString(v) || layoutTrackPercent.MatchString(v) {
		return v, true
	}
	return "", false
}

// layoutTrackPercentOf returns the percentage written in a track value.
func layoutTrackPercentOf(v string) (float64, bool) {
	v = strings.TrimSpace(v)
	if !layoutTrackPercent.MatchString(v) {
		return 0, false
	}
	n, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// layoutTrackItems splits a track list attribute ("60%, 40%") into its values.
// It returns nil when the attribute is missing or is not a literal.
func layoutTrackItems(e *Element, key string) []string {
	a := findAttr(e, key)
	if a == nil || a.hasExpr {
		return nil
	}
	raw := strings.TrimSpace(partsContent(a.text))
	if raw == "" {
		return nil
	}
	items := strings.Split(raw, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, strings.TrimSpace(item))
	}
	return out
}

// layoutTracks builds the CSS track list of a container. The count is the
// declared number of tracks; when it is not declared the number of written
// sizes is used. A track without a size is automatic.
func layoutTracks(e *Element, count int, listKey string) []string {
	items := layoutTrackItems(e, listKey)
	n := count
	if n <= 0 {
		n = len(items)
	}
	if n <= 0 {
		return nil
	}
	tracks := make([]string, 0, n)
	for i := 0; i < n; i++ {
		v := ""
		if i < len(items) {
			v = items[i]
		}
		track, ok := layoutTrackValue(v)
		if !ok || track == "" {
			track = layoutTrackAuto
		}
		tracks = append(tracks, track)
	}
	return tracks
}

// layoutRowTracks returns the CSS tracks of the columns of a row.
func layoutRowTracks(e *Element) []string {
	return layoutTracks(e, layoutTableIntAttr(e, "columns", 0, 0, 24), "columnWidths")
}

// layoutColTracks returns the CSS tracks of the rows of a column container.
func layoutColTracks(e *Element) []string {
	return layoutTracks(e, layoutTableIntAttr(e, "rows", 0, 0, 50), "rowHeights")
}

// layoutContainerStyle returns the inline style of a row or of a column
// container that declares its tracks or its size. An empty result means the
// container keeps the default flex layout.
func layoutContainerStyle(e *Element) string {
	var parts []string
	switch e.tag {
	case "row":
		if tracks := layoutRowTracks(e); len(tracks) > 0 {
			parts = append(parts, "display:grid", "grid-template-columns:"+strings.Join(tracks, " "))
		}
	case "col":
		if tracks := layoutColTracks(e); len(tracks) > 0 {
			parts = append(parts, "display:grid", "grid-template-rows:"+strings.Join(tracks, " "))
		}
	}
	if v := layoutSafeSize(layoutTableTextAttr(e, "width", "")); v != "" {
		parts = append(parts, "width:"+v)
	}
	if v := layoutSafeSize(layoutTableTextAttr(e, "height", "")); v != "" {
		parts = append(parts, "height:"+v)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ";") + ";"
}

// layoutContainerStyleAttr writes the inline style of a row or of a column
// container that declares its tracks or its size.
func layoutContainerStyleAttr(sb *strings.Builder, e *Element) {
	style := layoutContainerStyle(e)
	if style == "" {
		return
	}
	sb.WriteString(" style=\"")
	sb.WriteString(htmlEscape(style))
	sb.WriteString("\"")
}
