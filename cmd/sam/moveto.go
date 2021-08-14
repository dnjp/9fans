// #include "sam.h"

package main

import "unicode"

func moveto(f *File, r Range) {
	p1 := r.p1
	p2 := r.p2

	f.dot.r.p1 = p1
	f.dot.r.p2 = p2
	if f.rasp != nil {
		telldot(f)
		outTsl(Hmoveto, f.tag, f.dot.r.p1)
	}
}

func telldot(f *File) {
	if f.rasp == nil {
		panic_("telldot")
	}
	if f.dot.r.p1 == f.tdot.p1 && f.dot.r.p2 == f.tdot.p2 {
		return
	}
	outTsll(Hsetdot, f.tag, f.dot.r.p1, f.dot.r.p2)
	f.tdot = f.dot.r
}

func tellpat() {
	outTS(Hsetpat, &lastpat)
	patset = false
}

const CHARSHIFT = 128

func lookorigin(f *File, p0 Posn, ls Posn) {
	if p0 > f.b.nc {
		p0 = f.b.nc
	}
	oldp0 := p0
	p := p0
	var c rune
	var nc, nl int
	for nl = nc; c != -1 && nl < ls && nc < ls*CHARSHIFT; nc++ {
		p--
		c = filereadc(f, p)
		if c == '\n' {
			nl++
			oldp0 = p0 - nc
		}
	}
	if c == -1 {
		p0 = 0
	} else if nl == 0 {
		if p0 >= CHARSHIFT/2 {
			p0 -= CHARSHIFT / 2
		} else {
			p0 = 0
		}
	} else {
		p0 = oldp0
	}
	outTsl(Horigin, f.tag, p0)
}

func inmode(r rune, mode bool) bool {
	if !mode {
		return unicode.IsLetter(r) && unicode.IsDigit(r)
	}
	return unicode.IsSpace(r)
}

func clickmatch(f *File, cl, cr rune, dir int, p *Posn) bool {
	nest := 1

	for {
		var c rune
		if dir > 0 {
			if *p >= f.b.nc {
				break
			}
			c = filereadc(f, (*p))
			(*p)++
		} else {
			if *p == 0 {
				break
			}
			(*p)--
			c = filereadc(f, (*p))
		}
		if c == cr {
			nest--
			if nest == 0 {
				return true
			}
		} else if c == cl {
			nest++
		}
	}
	return cl == '\n' && nest == 1
}

func indexRune(s []rune, c rune) int {
	for i, cc := range s {
		if cc == c {
			return i
		}
	}
	return -1
}


// Stretches a selection out over current text,
// selecting matching range if possible.
// If there's no matching range, mode 0 selects
// a single alphanumeric region. Mode 1 selects
// a non-whitespace region.
func stretchsel(f *File, p1 Posn, mode bool) {
	if p1 > f.b.nc {
		return
	}
	f.dot.r.p2 = p1
	f.dot.r.p1 = f.dot.r.p2
	var p Posn
	for i := 0; i < len(left); i++ {
		l := left[i]
		r := right[i]
		/* try left match */
		p = p1
		var c rune
		if p1 == 0 {
			c = '\n'
		} else {
			c = filereadc(f, p-1)
		}
		if j := indexRune(l, c); j >= 0 {
			if clickmatch(f, c, r[j], 1, &p) {
				f.dot.r.p1 = p1
				f.dot.r.p2 = p
				if c != '\n' {
					f.dot.r.p2--
				}
			}
			return
		}
		/* try right match */
		p = p1
		if p1 == f.b.nc {
			c = '\n'
		} else {
			c = filereadc(f, p)
		}
		if j := indexRune(r, c); j >= 0 {
			if clickmatch(f, c, l[j], -1, &p) {
				f.dot.r.p1 = p
				if c != '\n' || p != 0 || filereadc(f, 0) == '\n' {
					f.dot.r.p1++
				}
				f.dot.r.p2 = p1
				if p1 < f.b.nc && c == '\n' {
					f.dot.r.p2++
				}
			}
			return
		}
	}
	/* try filling out word to right */
	p = p1
	for p < f.b.nc {
		p++
		if !inmode(filereadc(f, p-1), mode) {
			break
		}
		f.dot.r.p2++
	}
	/* try filling out word to left */
	p = p1
	for {
		p--
		if p < 0 || !inmode(filereadc(f, p), mode) {
			break
		}
		f.dot.r.p1--
	}
}
