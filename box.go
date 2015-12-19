// Box creates box plots for the plan9 plot(1) command.
// It reads data sets of the form <name> <number>* from standard input,
// and outputs a series of box plots for plot(1) on standard output.
//
// Example:
// 	echo "linear 1 2 3 4 5 6 exponential 2 4 8 16 32 64" | box -t Title | plot
// shows two box plots,
// one labeled "linear", showing the distribution of the numbers 1 2 3 4 5 6,
// and one labeled "exponential" showing the distribution of 2 4 8 16 32 64.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
)

var title = flag.String("t", "", "plot title")

func main() {
	flag.Parse()
	boxes, err := readBoxes(os.Stdin)
	if err != nil {
		fmt.Println("Read failed: ", err)
		return
	}
	draw(boxes, *title, os.Stdout)
}

type box struct {
	name                 string
	values               []float64
	min, q1, q2, q3, max float64
}

func readBoxes(r io.Reader) ([]box, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	var boxes []box
	if !scanner.Scan() {
		return boxes, nil
	}
	for {
		b, more := readBox(scanner)
		boxes = append(boxes, b)
		if !more {
			break
		}
	}
	return boxes, scanner.Err()
}

// ReadBox reads a box from a word-splitting *bufio.Scanner and returns it.
//
// The current Text() of the scanner is interpreted as the name of the box.
// Following tokens that are parsable by strconv.ParseFloat with 64-bits
// are interpreted as the box data.
// Data is scanned until the the scanner is empty or ParseFloat fails.
//
// The return value more indicates whether the scanner contains more tokens.
// If so, the current Text() of scanner after readBox returns
// is the first token that was not used by the readBox call,
// i.e., the next token for subsequent scanning.
func readBox(scanner *bufio.Scanner) (b box, more bool) {
	b.name = scanner.Text()
	for scanner.Scan() {
		v, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			more = true
			break
		}
		b.values = append(b.values, v)
	}
	if len(b.values) > 0 {
		b.min, b.q1, b.q2, b.q3, b.max = stats5(b.values)
	}
	return b, more
}

// Stats5 returns a five statistic summary of the values.
// The summary includes:
// the minimum value,
// the first quartile,
// the second quartile (a.k.a., the median),
// the third quartile,
// and the maximum value.
// Stats5 sorts the input slice.
func stats5(vs []float64) (min, q1, q2, q3, max float64) {
	sort.Float64s(vs)
	if len(vs) == 1 {
		return vs[0], vs[0], vs[0], vs[0], vs[0]
	}
	min = vs[0]
	q1 = median(vs[:len(vs)/2])
	q2 = median(vs)
	q3 = median(vs[len(vs)/2:])
	max = vs[len(vs)-1]
	return min, q1, q2, q3, max
}

// Median returns the median of a sorted float64 slice.
func median(vs []float64) float64 {
	if len(vs) == 1 {
		return vs[0]
	}
	med := vs[len(vs)/2]
	if len(vs)%2 == 0 {
		med += vs[len(vs)/2-1]
		med /= 2
	}
	return med
}

func draw(boxes []box, title string, w io.Writer) {
	const (
		yPad    = 0.05
		yText   = 0.02
		yBottom = yPad + yText
	)
	yTop := 1.0 - yPad
	if title != "" {
		fmt.Fprintf(w, "m %f %f\nt \"\\C%s\"\n", 0.5, 1.0-yText, title)
		yTop -= yText
	}

	n := float64(len(boxes))
	pad := (1.0 / n) / 3.0
	width := (1.0-(n+1)*pad) / n
	capWidth := width / 4.0

	x := pad
	yMin, yMax := minMax(boxes)
	tr := makeTr(yMin, yMax, yBottom, yTop)
	for _, b := range boxes {
		c := x + width/2.0
		fmt.Fprintf(w, "m %f %f\nt \"\\C%s\"\n", c, yText, b.name)
		bottom, top := tr(b.q1), tr(b.q3)
		fmt.Fprintf(w, "bo %f %f %f %f\n", x, bottom, x+width, top)
		fmt.Fprintf(w, "m %f %f\nt \"\\R%.3g\"\n", x, bottom, b.q1)
		fmt.Fprintf(w, "m %f %f\nt \"\\R%.3g\"\n", x, top, b.q3)
		med := tr(b.q2)
		fmt.Fprintf(w, "li %f %f %f %f\n", x, med, x+width, med)
		fmt.Fprintf(w, "m %f %f\nt \"\\R%.3g\"\n", x, med, b.q2)
		min := tr(b.min)
		fmt.Fprintf(w, "li %f %f %f %f\n", c-capWidth, min, c+capWidth, min)
		fmt.Fprintf(w, "li %f %f %f %f\n", c, bottom, c, min)
		fmt.Fprintf(w, "m %f %f\nt \"\\R%.3g\"\n", c-capWidth, min, b.min)
		max := tr(b.max)
		fmt.Fprintf(w, "li %f %f %f %f\n", c-capWidth, max, c+capWidth, max)
		fmt.Fprintf(w, "li %f %f %f %f\n", c, top, c, max)
		fmt.Fprintf(w, "m %f %f\nt \"\\R%.3g\"\n", c-capWidth, max, b.max)
		x += width + pad
	}
	fmt.Fprintf(w, "cl\n")
}

func minMax(boxes []box) (min, max float64) {
	min, max = math.Inf(1), math.Inf(-1)
	for _, b := range boxes {
		if b.min < min {
			min = b.min
		}
		if b.max > max {
			max = b.max
		}
	}
	return min, max
}

// MakeTr returns a function that applies a linear transform to its value
// such that the range [min0, max0] → [min1, max1].
func makeTr(min0, max0, min1, max1 float64) func(float64) float64 {
	d0 := max0 - min0
	d1 := max1 - min1
	return func(v float64) float64 { return ((v-min0)/d0)*d1 + min1 }
}
