package optimization

import (
	"github.com/clbanning/pseudo"
	"math"
)

type (
	DimacsSolver struct {
		precision   float64
		lowestLabel bool
		fifoBuckets bool
	}
)

func newDimacsEngine(param *EngineParam) UltpitEngine {

	engine := &DimacsSolver{
		precision:   param.Precision,
		lowestLabel: param.LowestLabel,
		fifoBuckets: param.FifoBuckets,
	}

	if math.Abs(engine.precision) < 1e6 {
		engine.precision = 100.0
	}

	return engine
}

func (this *DimacsSolver) computeSolution(ch chan<- string, data []float64, pre *Precedence) (solution []bool, r int) {

	count := len(data)

	solution = make([]bool, count)

	session := this.createSession(data, pre)

	cut := session.Cut()
	for _, n := range cut {
		if n != 1 {
			solution[n-2] = true
		}
	}

	return
}

func (this *DimacsSolver) createSession(data []float64, pre *Precedence) *pseudo.Session {
	session := pseudo.NewSession(pseudo.Context{LowestLabel: this.lowestLabel, FifoBuckets: this.fifoBuckets})
	sessionInitializer := pseudo.NewSessionInitializer(session)

	// source and sink
	numNodes := len(data) + 2
	// source to positive nodes, sink to negative nodes
	numArcs := len(data)

	for i := 0; i < len(data); i++ {
		// Each infinite arc
		if ind := pre.keys[i]; ind != MISSING {
			numArcs += len(pre.defs[ind])
		}
	}

	const SOURCE = 1
	SINK := numNodes

	sessionInitializer.Init(uint(numNodes), uint(numArcs))
	sessionInitializer.SetSource(SOURCE)
	sessionInitializer.SetSink(uint(SINK))

	var from_i, to_i int

	for i := 0; i < len(data); i++ {

		capacity := int(math.Abs(data[i]) * this.precision)

		if data[i] < 0 {
			from_i = i + 2
			to_i = SINK
		} else {
			from_i = SOURCE
			to_i = i + 2
		}

		sessionInitializer.AddArc(uint(from_i), uint(to_i), capacity)
	}

	// Now the infinite ones
	for i := 0; i < len(data); i++ {
		from_i = i + 2 // + 1 for psuedo, +1 for source
		if ind := pre.keys[i]; ind != MISSING {
			for _, off := range pre.defs[ind] {
				sessionInitializer.AddArc(uint(from_i), uint(from_i+off), int(math.MaxUint32))
			}
		}
	}

	sessionInitializer.Complete()
	return session
}
