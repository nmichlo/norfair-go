// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause
//
// This file contains color definitions ported from multiple sources:
//
// 1. CSS Color Constants - PIL/Pillow ImageColor module
//    Original Source: https://github.com/python-pillow/Pillow/blob/main/src/PIL/ImageColor.py
//    Original Copyright © 1997-2011 by Secret Labs AB
//    Original Copyright © 1995-2011 by Fredrik Lundh and contributors
//    Original Copyright © 2010 by Jeffrey A. Clark and contributors
//    Original License: MIT-CMU
//
// 2. Tableau Color Palettes (tab10, tab20) - Matplotlib
//    Original Source: https://github.com/matplotlib/matplotlib/blob/main/lib/matplotlib/_cm.py
//    Original Copyright (c) 2002-2011 John D. Hunter
//    Original Copyright (c) 2012- Matplotlib Development Team
//    Original License: Matplotlib License
//
// 3. Colorblind Palette - Seaborn
//    Original Source: https://github.com/mwaskom/seaborn/blob/master/seaborn/palettes.py
//    Original Copyright (c) 2012-2023, Michael L. Waskom
//    Original License: BSD-3-Clause

package imaging

import (
	"github.com/nmichlo/norfair-go/pkg/norfairgocolor"
)

// =============================================================================
// CSS Color Constants (140+ colors in BGR format)
// Source: PIL/Pillow ImageColor module
// =============================================================================

var (
	// Basic colors
	AliceBlue            = norfairgocolor.Color{255, 248, 240}
	AntiqueWhite         = norfairgocolor.Color{215, 235, 250}
	Aqua                 = norfairgocolor.Color{255, 255, 0}
	Aquamarine           = norfairgocolor.Color{212, 255, 127}
	Azure                = norfairgocolor.Color{255, 255, 240}
	Beige                = norfairgocolor.Color{220, 245, 245}
	Bisque               = norfairgocolor.Color{196, 228, 255}
	Black                = norfairgocolor.Color{0, 0, 0}
	BlanchedAlmond       = norfairgocolor.Color{205, 235, 255}
	Blue                 = norfairgocolor.Color{255, 0, 0}
	BlueViolet           = norfairgocolor.Color{226, 43, 138}
	Brown                = norfairgocolor.Color{42, 42, 165}
	BurlyWood            = norfairgocolor.Color{135, 184, 222}
	CadetBlue            = norfairgocolor.Color{160, 158, 95}
	Chartreuse           = norfairgocolor.Color{0, 255, 127}
	Chocolate            = norfairgocolor.Color{30, 105, 210}
	Coral                = norfairgocolor.Color{80, 127, 255}
	CornflowerBlue       = norfairgocolor.Color{237, 149, 100}
	Cornsilk             = norfairgocolor.Color{220, 248, 255}
	Crimson              = norfairgocolor.Color{60, 20, 220}
	Cyan                 = norfairgocolor.Color{255, 255, 0}
	DarkBlue             = norfairgocolor.Color{139, 0, 0}
	DarkCyan             = norfairgocolor.Color{139, 139, 0}
	DarkGoldenrod        = norfairgocolor.Color{11, 134, 184}
	DarkGray             = norfairgocolor.Color{169, 169, 169}
	DarkGreen            = norfairgocolor.Color{0, 100, 0}
	DarkKhaki            = norfairgocolor.Color{107, 183, 189}
	DarkMagenta          = norfairgocolor.Color{139, 0, 139}
	DarkOliveGreen       = norfairgocolor.Color{47, 107, 85}
	DarkOrange           = norfairgocolor.Color{0, 140, 255}
	DarkOrchid           = norfairgocolor.Color{204, 50, 153}
	DarkRed              = norfairgocolor.Color{0, 0, 139}
	DarkSalmon           = norfairgocolor.Color{122, 150, 233}
	DarkSeaGreen         = norfairgocolor.Color{143, 188, 143}
	DarkSlateBlue        = norfairgocolor.Color{139, 61, 72}
	DarkSlateGray        = norfairgocolor.Color{79, 79, 47}
	DarkTurquoise        = norfairgocolor.Color{209, 206, 0}
	DarkViolet           = norfairgocolor.Color{211, 0, 148}
	DeepPink             = norfairgocolor.Color{147, 20, 255}
	DeepSkyBlue          = norfairgocolor.Color{255, 191, 0}
	DimGray              = norfairgocolor.Color{105, 105, 105}
	DodgerBlue           = norfairgocolor.Color{255, 144, 30}
	FireBrick            = norfairgocolor.Color{34, 34, 178}
	FloralWhite          = norfairgocolor.Color{240, 250, 255}
	ForestGreen          = norfairgocolor.Color{34, 139, 34}
	Fuchsia              = norfairgocolor.Color{255, 0, 255}
	Gainsboro            = norfairgocolor.Color{220, 220, 220}
	GhostWhite           = norfairgocolor.Color{255, 248, 248}
	Gold                 = norfairgocolor.Color{0, 215, 255}
	Goldenrod            = norfairgocolor.Color{32, 165, 218}
	Gray                 = norfairgocolor.Color{128, 128, 128}
	Green                = norfairgocolor.Color{0, 128, 0}
	GreenYellow          = norfairgocolor.Color{47, 255, 173}
	Honeydew             = norfairgocolor.Color{240, 255, 240}
	HotPink              = norfairgocolor.Color{180, 105, 255}
	IndianRed            = norfairgocolor.Color{92, 92, 205}
	Indigo               = norfairgocolor.Color{130, 0, 75}
	Ivory                = norfairgocolor.Color{240, 255, 255}
	Khaki                = norfairgocolor.Color{140, 230, 240}
	Lavender             = norfairgocolor.Color{250, 230, 230}
	LavenderBlush        = norfairgocolor.Color{245, 240, 255}
	LawnGreen            = norfairgocolor.Color{0, 252, 124}
	LemonChiffon         = norfairgocolor.Color{205, 250, 255}
	LightBlue            = norfairgocolor.Color{230, 216, 173}
	LightCoral           = norfairgocolor.Color{128, 128, 240}
	LightCyan            = norfairgocolor.Color{255, 255, 224}
	LightGoldenrodYellow = norfairgocolor.Color{210, 250, 250}
	LightGray            = norfairgocolor.Color{211, 211, 211}
	LightGreen           = norfairgocolor.Color{144, 238, 144}
	LightPink            = norfairgocolor.Color{193, 182, 255}
	LightSalmon          = norfairgocolor.Color{122, 160, 255}
	LightSeaGreen        = norfairgocolor.Color{170, 178, 32}
	LightSkyBlue         = norfairgocolor.Color{250, 206, 135}
	LightSlateGray       = norfairgocolor.Color{153, 136, 119}
	LightSteelBlue       = norfairgocolor.Color{222, 196, 176}
	LightYellow          = norfairgocolor.Color{224, 255, 255}
	Lime                 = norfairgocolor.Color{0, 255, 0}
	LimeGreen            = norfairgocolor.Color{50, 205, 50}
	Linen                = norfairgocolor.Color{230, 240, 250}
	Magenta              = norfairgocolor.Color{255, 0, 255}
	Maroon               = norfairgocolor.Color{0, 0, 128}
	MediumAquamarine     = norfairgocolor.Color{170, 205, 102}
	MediumBlue           = norfairgocolor.Color{205, 0, 0}
	MediumOrchid         = norfairgocolor.Color{211, 85, 186}
	MediumPurple         = norfairgocolor.Color{219, 112, 147}
	MediumSeaGreen       = norfairgocolor.Color{113, 179, 60}
	MediumSlateBlue      = norfairgocolor.Color{238, 104, 123}
	MediumSpringGreen    = norfairgocolor.Color{154, 250, 0}
	MediumTurquoise      = norfairgocolor.Color{204, 209, 72}
	MediumVioletRed      = norfairgocolor.Color{133, 21, 199}
	MidnightBlue         = norfairgocolor.Color{112, 25, 25}
	MintCream            = norfairgocolor.Color{250, 255, 245}
	MistyRose            = norfairgocolor.Color{225, 228, 255}
	Moccasin             = norfairgocolor.Color{181, 228, 255}
	NavajoWhite          = norfairgocolor.Color{173, 222, 255}
	Navy                 = norfairgocolor.Color{128, 0, 0}
	OldLace              = norfairgocolor.Color{230, 245, 253}
	Olive                = norfairgocolor.Color{0, 128, 128}
	OliveDrab            = norfairgocolor.Color{35, 142, 107}
	Orange               = norfairgocolor.Color{0, 165, 255}
	OrangeRed            = norfairgocolor.Color{0, 69, 255}
	Orchid               = norfairgocolor.Color{214, 112, 218}
	PaleGoldenrod        = norfairgocolor.Color{170, 232, 238}
	PaleGreen            = norfairgocolor.Color{152, 251, 152}
	PaleTurquoise        = norfairgocolor.Color{238, 238, 175}
	PaleVioletRed        = norfairgocolor.Color{147, 112, 219}
	PapayaWhip           = norfairgocolor.Color{213, 239, 255}
	PeachPuff            = norfairgocolor.Color{185, 218, 255}
	Peru                 = norfairgocolor.Color{63, 133, 205}
	Pink                 = norfairgocolor.Color{203, 192, 255}
	Plum                 = norfairgocolor.Color{221, 160, 221}
	PowderBlue           = norfairgocolor.Color{230, 224, 176}
	Purple               = norfairgocolor.Color{128, 0, 128}
	Red                  = norfairgocolor.Color{0, 0, 255}
	RosyBrown            = norfairgocolor.Color{143, 143, 188}
	RoyalBlue            = norfairgocolor.Color{225, 105, 65}
	SaddleBrown          = norfairgocolor.Color{19, 69, 139}
	Salmon               = norfairgocolor.Color{114, 128, 250}
	SandyBrown           = norfairgocolor.Color{96, 164, 244}
	SeaGreen             = norfairgocolor.Color{87, 139, 46}
	Seashell             = norfairgocolor.Color{238, 245, 255}
	Sienna               = norfairgocolor.Color{45, 82, 160}
	Silver               = norfairgocolor.Color{192, 192, 192}
	SkyBlue              = norfairgocolor.Color{235, 206, 135}
	SlateBlue            = norfairgocolor.Color{205, 90, 106}
	SlateGray            = norfairgocolor.Color{144, 128, 112}
	Snow                 = norfairgocolor.Color{250, 250, 255}
	SpringGreen          = norfairgocolor.Color{127, 255, 0}
	SteelBlue            = norfairgocolor.Color{180, 130, 70}
	Tan                  = norfairgocolor.Color{140, 180, 210}
	Teal                 = norfairgocolor.Color{128, 128, 0}
	Thistle              = norfairgocolor.Color{216, 191, 216}
	Tomato               = norfairgocolor.Color{71, 99, 255}
	Turquoise            = norfairgocolor.Color{208, 224, 64}
	Violet               = norfairgocolor.Color{238, 130, 238}
	Wheat                = norfairgocolor.Color{179, 222, 245}
	White                = norfairgocolor.Color{255, 255, 255}
	WhiteSmoke           = norfairgocolor.Color{245, 245, 245}
	Yellow               = norfairgocolor.Color{0, 255, 255}
	YellowGreen          = norfairgocolor.Color{50, 205, 154}
)

// ColorMap maps color names to Color values (lowercase for case-insensitive lookup).
var ColorMap = map[string]norfairgocolor.Color{
	"aliceblue":            AliceBlue,
	"antiquewhite":         AntiqueWhite,
	"aqua":                 Aqua,
	"aquamarine":           Aquamarine,
	"azure":                Azure,
	"beige":                Beige,
	"bisque":               Bisque,
	"black":                Black,
	"blanchedalmond":       BlanchedAlmond,
	"blue":                 Blue,
	"blueviolet":           BlueViolet,
	"brown":                Brown,
	"burlywood":            BurlyWood,
	"cadetblue":            CadetBlue,
	"chartreuse":           Chartreuse,
	"chocolate":            Chocolate,
	"coral":                Coral,
	"cornflowerblue":       CornflowerBlue,
	"cornsilk":             Cornsilk,
	"crimson":              Crimson,
	"cyan":                 Cyan,
	"darkblue":             DarkBlue,
	"darkcyan":             DarkCyan,
	"darkgoldenrod":        DarkGoldenrod,
	"darkgray":             DarkGray,
	"darkgreen":            DarkGreen,
	"darkkhaki":            DarkKhaki,
	"darkmagenta":          DarkMagenta,
	"darkolivegreen":       DarkOliveGreen,
	"darkorange":           DarkOrange,
	"darkorchid":           DarkOrchid,
	"darkred":              DarkRed,
	"darksalmon":           DarkSalmon,
	"darkseagreen":         DarkSeaGreen,
	"darkslateblue":        DarkSlateBlue,
	"darkslategray":        DarkSlateGray,
	"darkturquoise":        DarkTurquoise,
	"darkviolet":           DarkViolet,
	"deeppink":             DeepPink,
	"deepskyblue":          DeepSkyBlue,
	"dimgray":              DimGray,
	"dodgerblue":           DodgerBlue,
	"firebrick":            FireBrick,
	"floralwhite":          FloralWhite,
	"forestgreen":          ForestGreen,
	"fuchsia":              Fuchsia,
	"gainsboro":            Gainsboro,
	"ghostwhite":           GhostWhite,
	"gold":                 Gold,
	"goldenrod":            Goldenrod,
	"gray":                 Gray,
	"green":                Green,
	"greenyellow":          GreenYellow,
	"honeydew":             Honeydew,
	"hotpink":              HotPink,
	"indianred":            IndianRed,
	"indigo":               Indigo,
	"ivory":                Ivory,
	"khaki":                Khaki,
	"lavender":             Lavender,
	"lavenderblush":        LavenderBlush,
	"lawngreen":            LawnGreen,
	"lemonchiffon":         LemonChiffon,
	"lightblue":            LightBlue,
	"lightcoral":           LightCoral,
	"lightcyan":            LightCyan,
	"lightgoldenrodyellow": LightGoldenrodYellow,
	"lightgray":            LightGray,
	"lightgreen":           LightGreen,
	"lightpink":            LightPink,
	"lightsalmon":          LightSalmon,
	"lightseagreen":        LightSeaGreen,
	"lightskyblue":         LightSkyBlue,
	"lightslategray":       LightSlateGray,
	"lightsteelblue":       LightSteelBlue,
	"lightyellow":          LightYellow,
	"lime":                 Lime,
	"limegreen":            LimeGreen,
	"linen":                Linen,
	"magenta":              Magenta,
	"maroon":               Maroon,
	"mediumaquamarine":     MediumAquamarine,
	"mediumblue":           MediumBlue,
	"mediumorchid":         MediumOrchid,
	"mediumpurple":         MediumPurple,
	"mediumseagreen":       MediumSeaGreen,
	"mediumslateblue":      MediumSlateBlue,
	"mediumspringgreen":    MediumSpringGreen,
	"mediumturquoise":      MediumTurquoise,
	"mediumvioletred":      MediumVioletRed,
	"midnightblue":         MidnightBlue,
	"mintcream":            MintCream,
	"mistyrose":            MistyRose,
	"moccasin":             Moccasin,
	"navajowhite":          NavajoWhite,
	"navy":                 Navy,
	"oldlace":              OldLace,
	"olive":                Olive,
	"olivedrab":            OliveDrab,
	"orange":               Orange,
	"orangered":            OrangeRed,
	"orchid":               Orchid,
	"palegoldenrod":        PaleGoldenrod,
	"palegreen":            PaleGreen,
	"paleturquoise":        PaleTurquoise,
	"palevioletred":        PaleVioletRed,
	"papayawhip":           PapayaWhip,
	"peachpuff":            PeachPuff,
	"peru":                 Peru,
	"pink":                 Pink,
	"plum":                 Plum,
	"powderblue":           PowderBlue,
	"purple":               Purple,
	"red":                  Red,
	"rosybrown":            RosyBrown,
	"royalblue":            RoyalBlue,
	"saddlebrown":          SaddleBrown,
	"salmon":               Salmon,
	"sandybrown":           SandyBrown,
	"seagreen":             SeaGreen,
	"seashell":             Seashell,
	"sienna":               Sienna,
	"silver":               Silver,
	"skyblue":              SkyBlue,
	"slateblue":            SlateBlue,
	"slategray":            SlateGray,
	"snow":                 Snow,
	"springgreen":          SpringGreen,
	"steelblue":            SteelBlue,
	"tan":                  Tan,
	"teal":                 Teal,
	"thistle":              Thistle,
	"tomato":               Tomato,
	"turquoise":            Turquoise,
	"violet":               Violet,
	"wheat":                Wheat,
	"white":                White,
	"whitesmoke":           WhiteSmoke,
	"yellow":               Yellow,
	"yellowgreen":          YellowGreen,
}

// =============================================================================
// Color Palettes
// =============================================================================

// Tab10 palette (10 colors from Matplotlib).
// Source: https://github.com/matplotlib/matplotlib/blob/main/lib/matplotlib/_cm.py
var Tab10 = []norfairgocolor.Color{
	{214, 127, 31},  // Blue
	{134, 86, 255},  // Orange
	{113, 178, 44},  // Green
	{83, 64, 214},   // Red
	{190, 117, 148}, // Purple
	{107, 76, 140},  // Brown
	{218, 127, 227}, // Pink
	{114, 114, 127}, // Gray
	{51, 176, 188},  // Olive
	{201, 195, 23},  // Cyan
}

// Tab20 palette (20 colors from Matplotlib).
// Source: https://github.com/matplotlib/matplotlib/blob/main/lib/matplotlib/_cm.py
var Tab20 = []norfairgocolor.Color{
	{214, 127, 31}, {228, 173, 95}, // Blue
	{134, 86, 255}, {184, 154, 255}, // Orange
	{113, 178, 44}, {153, 208, 104}, // Green
	{83, 64, 214}, {133, 112, 237}, // Red
	{190, 117, 148}, {216, 165, 188}, // Purple
	{107, 76, 140}, {157, 126, 186}, // Brown
	{218, 127, 227}, {235, 172, 243}, // Pink
	{114, 114, 127}, {168, 168, 179}, // Gray
	{51, 176, 188}, {111, 216, 222}, // Olive
	{201, 195, 23}, {231, 227, 99}, // Cyan
}

// Colorblind palette (8 colorblind-friendly colors from Seaborn).
// Source: https://github.com/mwaskom/seaborn/blob/master/seaborn/palettes.py
var Colorblind = []norfairgocolor.Color{
	{30, 119, 180},  // Blue
	{255, 158, 74},  // Orange
	{153, 121, 44},  // Green
	{181, 77, 204},  // Purple
	{107, 74, 222},  // Brown
	{217, 127, 227}, // Pink
	{128, 128, 128}, // Gray
	{0, 153, 214},   // Cyan
}
