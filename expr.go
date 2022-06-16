package main

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

type fieldType int

const (
	fieldSeconds fieldType = iota
	fieldMinutes
	fieldHours
	fieldDaysOfMonth
	fieldMonths
	fieldDaysOfWeek
)

func (t fieldType) String() string {
	switch t {
	case fieldSeconds:
		return "seconds"
	case fieldMinutes:
		return "minutes"
	case fieldHours:
		return "hours"
	case fieldDaysOfMonth:
		return "days of month"
	case fieldMonths:
		return "months"
	case fieldDaysOfWeek:
		return "days of week"
	default:
		return strconv.FormatInt(int64(t), 10)
	}
}

var (
	dowFromName = map[string]int{
		"sun": 1, "suN": 1, "sUn": 1, "sUN": 1, "Sun": 1, "SuN": 1, "SUn": 1, "SUN": 1,
		"mon": 2, "moN": 2, "mOn": 2, "mON": 2, "Mon": 2, "MoN": 2, "MOn": 2, "MON": 2,
		"tue": 3, "tuE": 3, "tUe": 3, "tUE": 3, "Tue": 3, "TuE": 3, "TUe": 3, "TUE": 3,
		"wed": 4, "weD": 4, "wEd": 4, "wED": 4, "Wed": 4, "WeD": 4, "WEd": 4, "WED": 4,
		"thu": 5, "thU": 5, "tHu": 5, "tHU": 5, "Thu": 5, "ThU": 5, "THu": 5, "THU": 5,
		"fri": 6, "frI": 6, "fRi": 6, "fRI": 6, "Fri": 6, "FrI": 6, "FRi": 6, "FRI": 6,
		"sat": 7, "saT": 7, "sAt": 7, "sAT": 7, "Sat": 7, "SaT": 7, "SAt": 7, "SAT": 7,
	}
	monFromName = map[string]int{
		"jan": 1, "jaN": 1, "jAn": 1, "jAN": 1, "Jan": 1, "JaN": 1, "JAn": 1, "JAN": 1,
		"feb": 2, "feB": 2, "fEb": 2, "fEB": 2, "Feb": 2, "FeB": 2, "FEb": 2, "FEB": 2,
		"mar": 3, "maR": 3, "mAr": 3, "mAR": 3, "Mar": 3, "MaR": 3, "MAr": 3, "MAR": 3,
		"apr": 4, "apR": 4, "aPr": 4, "aPR": 4, "Apr": 4, "ApR": 4, "APr": 4, "APR": 4,
		"may": 5, "maY": 5, "mAy": 5, "mAY": 5, "May": 5, "MaY": 5, "MAy": 5, "MAY": 5,
		"jun": 6, "juN": 6, "jUn": 6, "jUN": 6, "Jun": 6, "JuN": 6, "JUn": 6, "JUN": 6,
		"jul": 7, "juL": 7, "jUl": 7, "jUL": 7, "Jul": 7, "JuL": 7, "JUl": 7, "JUL": 7,
		"aug": 8, "auG": 8, "aUg": 8, "aUG": 8, "Aug": 8, "AuG": 8, "AUg": 8, "AUG": 8,
		"sep": 9, "seP": 9, "sEp": 9, "sEP": 9, "Sep": 9, "SeP": 9, "SEp": 9, "SEP": 9,
		"oct": 10, "ocT": 10, "oCt": 10, "oCT": 10, "Oct": 10, "OcT": 10, "OCt": 10, "OCT": 10,
		"nov": 11, "noV": 11, "nOv": 11, "nOV": 11, "Nov": 11, "NoV": 11, "NOv": 11, "NOV": 11,
		"dec": 12, "deC": 12, "dEc": 12, "dEC": 12, "Dec": 12, "DeC": 12, "DEc": 12, "DEC": 12,
	}
)

const (
	domRange29 = (uint64(1)<<(29+iota) - 1) << 1
	domRange30
)

type Expr struct {
	expr                   string
	s, m, h, dom, mon, dow uint64
}

func main() {
	expr, err := Parse("0,2 * * * 12 SUN")
	if err != nil {
		panic(err)
	}
	fmt.Println(strconv.FormatUint(expr.s, 2))
}

func MustParse(expr string) Expr {
	e, err := Parse(expr)
	if err != nil {
		panic(err)
	}
	return e
}

func Parse(expr string) (e Expr, err error) {
	if e, err = New(splitFields(expr)); err != nil {
		return e, fmt.Errorf("parsing %q: %v", expr, err)
	}
	return
}

func splitFields(expr string) (s, m, h, dom, mon, dow string) {
	s, expr, _ = strings.Cut(expr, " ")
	m, expr, _ = strings.Cut(expr, " ")
	h, expr, _ = strings.Cut(expr, " ")
	dom, expr, _ = strings.Cut(expr, " ")
	mon, expr, _ = strings.Cut(expr, " ")
	dow, _, _ = strings.Cut(expr, " ")
	return
}

func MustNew(seconds, minutes, hours, daysOfMonth, months, daysOfWeek string) Expr {
	e, err := New(seconds, minutes, hours, daysOfMonth, months, daysOfWeek)
	if err != nil {
		panic(err)
	}
	return e
}

func New(seconds, minutes, hours, daysOfMonth, months, daysOfWeek string) (e Expr, err error) {
	parseField := func(typ fieldType, field string, min, max int) (v uint64) {
		if err == nil {
			v, err = parseGroups(typ, field, min, max)
		}
		return v
	}
	s := parseField(fieldSeconds, seconds, 0, 59)
	m := parseField(fieldMinutes, minutes, 0, 59)
	h := parseField(fieldHours, hours, 0, 23)
	dom := parseField(fieldDaysOfMonth, daysOfMonth, 1, 31)
	mon := parseField(fieldMonths, months, 1, 12)
	dow := parseField(fieldDaysOfWeek, daysOfWeek, 1, 7) >> 1 // Move range 1-7 to 0-6.
	if err != nil {
		return e, err
	}

	// Detect impossible combinations of month/day pairs, e.g., February 30th.
	const monthsWith31Days = 1<<1 | 1<<3 | 1<<5 | 1<<7 | 1<<8 | 1<<10 | 1<<12
	if mon&monthsWith31Days == 0 {
		onlyFeb := mon == 1<<2
		domAllowed := domRange30
		if onlyFeb {
			domAllowed = domRange29
		}
		if dom&domAllowed == 0 && onlyFeb {
			return e, fmt.Errorf("field %q doesn't match any day of month 2", fieldDaysOfMonth)
		} else if dom&domAllowed == 0 {
			return e, fmt.Errorf("field %q doesn't match any day of months 4, 6, 9 or 11", fieldDaysOfMonth)
		}
	}

	return Expr{
		expr: join(" ", seconds, minutes, hours, daysOfMonth, months, daysOfWeek),
		s:    s,
		m:    m,
		h:    h,
		dom:  dom,
		mon:  mon,
		dow:  dow,
	}, nil
}

func join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}

/*
parseGroups implements the following BNF:

	groups     ::= expr ( ',' expr )*
	expr       ::= '*' | '?' | rangeOrNum ( '/' step )?
	rangeOrNum ::= number ( '-' number )?
	step       ::= number
	number     ::= digit+
	digit      ::= '0'..'9'
*/
func parseGroups(typ fieldType, groups string, min, max int) (val uint64, err error) {
	if groups == "" {
		return 0, &parseGroupsError{
			typ: typ,
			err: errors.New("field is empty"),
		}
	}
	for len(groups) > 0 {
		expr, groupsRest, found := strings.Cut(groups, ",")
		if found && groupsRest == "" {
			return 0, &parseGroupsError{
				typ: typ,
				err: errors.New("field with trailing comma"),
			}
		}
		groups = groupsRest

		if expr == "*" || expr == "?" && (typ == fieldDaysOfMonth || typ == fieldDaysOfWeek) {
			for i := min; i <= max; i++ {
				val |= uint64(1) << i
			}
			continue
		}

		incr := 1
		rangeOrNum, step, _ := strings.Cut(expr, "/")
		if step != "" {
			if incr, err = strconv.Atoi(step); err != nil {
				return 0, &parseGroupsError{
					typ: typ,
					err: fmt.Errorf("step %q: %v", step, err),
				}
			} else if incr <= 0 {
				return 0, &parseGroupsError{
					typ: typ,
					err: fmt.Errorf("step %q is non-negative", step),
				}
			}

		}
		from, to := min, max
		if rangeFrom, rangeTo, _ := strings.Cut(rangeOrNum, "-"); rangeTo != "" {
			// Range.
			if from, err = parseNumber(typ, rangeFrom); err == nil {
				to, err = parseNumber(typ, rangeTo)
			}
			if err != nil {
				return 0, &parseGroupsError{
					typ: typ,
					err: fmt.Errorf("rangeOrNum %q: %v", rangeOrNum, err),
				}
			}
			if from > to {
				from, to = to, from
			}
		} else {
			// Number.
			if from, err = parseNumber(typ, rangeOrNum); err != nil {
				return 0, &parseGroupsError{
					typ: typ,
					err: fmt.Errorf("rangeOrNum %q: %v", rangeOrNum, err),
				}
			}
			if step == "" {
				to = from
			}
		}
		if from < min || from > max || to < min || to > max {
			return 0, &parseGroupsError{
				typ: typ,
				err: fmt.Errorf("rangeOrNum %q specifies values not accepted by field", rangeOrNum),
			}
		}
		for i := from; i <= to; i += incr {
			val |= uint64(1) << i
		}
	}
	return val, nil
}

func parseNumber(typ fieldType, s string) (v int, err error) {
	switch typ {
	case fieldDaysOfWeek:
		if v, ok := dowFromName[s]; ok {
			return v, nil
		}
	case fieldMonths:
		if v, ok := monFromName[s]; ok {
			return v, nil
		}
	}
	return strconv.Atoi(s)
}

type parseGroupsError struct {
	typ fieldType
	err error
}

func (e *parseGroupsError) Error() string {
	return fmt.Sprintf("field %q: %v", e.typ, e.err)
}

func (e *Expr) String() string {
	return e.expr
}

func (e *Expr) MarshalText() ([]byte, error) {
	return []byte(e.expr), nil
}

func (e *Expr) UnmarshalText(text []byte) (err error) {
	*e, err = Parse(string(text))
	return err
}

func (e *Expr) Prev(from time.Time) time.Time {
	t := from.Truncate(time.Minute).Add(-time.Minute)
	eS, eM, eH, eDom, eMon, eDow := e.s, e.m, e.h, e.dom, e.mon, e.dow

	var y int
	var mon time.Month
	var dom int
	var dow time.Weekday
	var h, m, s int
day:
	for {
		y, mon, dom = t.Date()
		dow = t.Weekday()
		switch {
		case eMon&(1<<mon) == 0:
			mon = prev(mon, time.January, eMon) + 1
			dom = 0
		case eDom&(1<<dom) == 0:
			dom = prev(dom, 1, eDom)
		case eDow&(1<<dow) == 0:
			dowPrev := prev(dow, time.Sunday, eDow)
			dom -= int(dow - dowPrev)
		default:
			break day
		}
		t = time.Date(y, mon, dom, 23, 59, 0, 0, t.Location())
	}
	doy := t.YearDay()
hour:
	for {
		h, m, s = t.Clock()
		switch {
		case eH&(1<<h) == 0:
			h = prev(h, 0, eH) + 1
			m = -1
			s = 59
		case eM&(1<<m) == 0:
			m = prev(m, 0, eM) + 1
			s = -1
		case eS&(1<<s) == 0:
			s = prev(s, 0, eS)
		default:
			break hour
		}
		t = time.Date(y, mon, dom, h, m, s, 0, t.Location())
		if t.YearDay() != doy {
			// We hit a different day.
			goto day
		}
	}
	return t
}

func (e *Expr) Next(from time.Time) time.Time {
	t := from.Truncate(time.Second).Add(time.Second)
	eS, eM, eH, eDom, eMon, eDow := e.s, e.m, e.h, e.dom, e.mon, e.dow

	var y int
	var mon time.Month
	var dom int
	var dow time.Weekday
	var h, m, s int
day:
	for {
		y, mon, dom = t.Date()
		dow = t.Weekday()
		switch {
		case eMon&(1<<mon) == 0:
			mon = next(mon, time.December, eMon)
			dom = 1
		case eDom&(1<<dom) == 0:
			dom = next(dom, maxDomForMon(y, mon), eDom)
		case eDow&(1<<dow) == 0:
			dowNext := next(dow, time.Saturday, eDow)
			dom += int(dowNext - dow)
		default:
			break day
		}
		t = time.Date(y, mon, dom, 0, 0, 0, 0, t.Location())
	}
	doy := t.YearDay()
hour:
	for {
		h, m, s = t.Clock()
		switch {
		case eH&(1<<h) == 0:
			h = next(h, 23, eH)
			m = 0
			s = 0
		case eM&(1<<m) == 0:
			m = next(m, 59, eM)
			s = 0
		case eS&(1<<s) == 0:
			s = next(s, 59, eS)
		default:
			break hour
		}
		t = time.Date(y, mon, dom, h, m, s, 0, t.Location())
		if t.YearDay() != doy {
			// We hit a different day.
			goto day
		}
	}
	return t
}

func maxDomForMon(y int, mon time.Month) int {
	switch mon {
	case time.February:
		if (y%4 == 0 && y%100 != 0) || y%400 == 0 {
			// Leap year.
			return 29
		}
		return 28
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

type timeType interface {
	int | time.Month | time.Weekday
}

// next returns the position of the first most-significant bit set in field
// after position i. The result, n, is such that n <= limit+1.
func next[T timeType](i, limit T, field uint64) (n T) {
	i++
	mask := ^uint64(0) << i
	next := T(bits.TrailingZeros64(field & mask))
	if next > limit {
		next = limit + 1
	}
	return next
}

// prev returns the position of the first least-significant bit set in field
// before position i. The result, p, is such that p >= limit-1.
func prev[T timeType](i, limit T, field uint64) (p T) {
	i++
	mask := ^(^uint64(0) << i)
	prev := T(bits.Len64(field&mask) - 1)
	if prev < limit {
		prev = limit - 1
	}
	return prev
}
