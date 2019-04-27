package gerber

import (
	"fmt"
	"io"
	"log"
)

// Layer represents a printed circuit board layer.
type Layer struct {
	// Filename is the filename of the Gerber layer.
	Filename string
	// Primitives represents the collection of primitives.
	Primitives []Primitive
	// Apertures represents the apertures used in the layer.
	Apertures []*Aperture

	// apertureMap maps an aperture to its index in the Apertures slice.
	apertureMap map[string]int
	// g is the root Gerber object.
	g   *Gerber
	mbb *MBB // cached minimum bounding box
}

// Add adds primitives to a layer.
// It generates new apertures as necessary.
func (l *Layer) Add(primitives ...Primitive) {
	for _, p := range primitives {
		a := p.Aperture()
		if a == nil {
			continue // use the default layer
		}
		id := a.ID()
		if _, ok := l.apertureMap[id]; ok {
			continue
		}
		l.apertureMap[id] = len(l.Apertures)
		l.Apertures = append(l.Apertures, a)
	}
	l.Primitives = append(l.Primitives, primitives...)
}

// WriteGerber writes a layer to its corresponding Gerber layer file.
func (l *Layer) WriteGerber(w io.Writer) error {
	io.WriteString(w, "%FSLAX36Y36*%\n")
	io.WriteString(w, "%MOMM*%\n")
	io.WriteString(w, "%LPD*%\n")

	io.WriteString(w, "%ADD11C,0.00100*%\n")
	for i, a := range l.Apertures {
		a.WriteGerber(w, 12+i)
	}

	for _, p := range l.Primitives {
		ai := l.apertureMap[p.Aperture().ID()]
		p.WriteGerber(w, 12+ai)
	}

	io.WriteString(w, "M02*\n")
	return nil
}

// MBB returns the minimum bounding box of the layer in millimeters.
func (l *Layer) MBB() MBB {
	if l.mbb != nil {
		return *l.mbb
	}
	for i, p := range l.Primitives {
		v := p.MBB()
		if i == 0 {
			l.mbb = &v
			continue
		}
		l.mbb.Join(&v)
	}
	if l.mbb == nil { // no primitives
		log.Printf("No primivites on layer %v", l.Filename)
		l.mbb = &MBB{}
	}

	return *l.mbb
}

func (g *Gerber) makeLayer(extension string) *Layer {
	layer := &Layer{
		Filename:    g.FilenamePrefix + "." + extension,
		apertureMap: map[string]int{"default": -1},
	}
	g.Layers = append(g.Layers, layer)
	return layer
}

// TopCopper adds a top copper layer to the design
// and returns the layer.
func (g *Gerber) TopCopper() *Layer {
	return g.makeLayer("gtl")
}

// TopSolderMask adds a top solder mask layer to the design
// and returns the layer.
func (g *Gerber) TopSolderMask() *Layer {
	return g.makeLayer("gts")
}

// TopSilkscreen adds a top silkscreen layer to the design
// and returns the layer.
func (g *Gerber) TopSilkscreen() *Layer {
	return g.makeLayer("gto")
}

// BottomCopper adds a bottom copper layer to the design
// and returns the layer.
func (g *Gerber) BottomCopper() *Layer {
	return g.makeLayer("gbl")
}

// BottomSolderMask adds a bottom solder mask layer to the design
// and returns the layer.
func (g *Gerber) BottomSolderMask() *Layer {
	return g.makeLayer("gbs")
}

// BottomSilkscreen adds a bottom silkscreen layer to the design
// and returns the layer.
func (g *Gerber) BottomSilkscreen() *Layer {
	return g.makeLayer("gbo")
}

// LayerN adds a layer-n copper layer to a multi-layer design
// and returns the layer.
func (g *Gerber) LayerN(n int) *Layer {
	return g.makeLayer(fmt.Sprintf("gl%v", n))
}

// Drill adds a drill layer to the design
// and returns the layer.
func (g *Gerber) Drill() *Layer {
	return g.makeLayer("drl")
}

// Outline adds an outline layer to the design
// and returns the layer.
func (g *Gerber) Outline() *Layer {
	return g.makeLayer("gko")
}
