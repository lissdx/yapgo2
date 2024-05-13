package pipeline

const GenerateInfinitely = 0

type generatorConfig struct {
	timesToGenerate uint
}

//	type sliceGeneratorConfig struct {
//		isDynamicBufferSize bool
//		//ctx            context.Context
//	}
type GenConfigOption interface {
	apply(*generatorConfig)
}

type genConfOptionFn func(*generatorConfig)

func (fn genConfOptionFn) apply(config *generatorConfig) {
	fn(config)
}

func WithTimesToGenerate(timesToGenerate uint) genConfOptionFn {
	return func(config *generatorConfig) {
		config.timesToGenerate = timesToGenerate
	}
}

//
//// WithDynamicBufferSize if true the size of channel
//// wil be calculated on the time of SliceToStreamGeneratorFn
//// creation and will be equal to the size of passed slice
//func WithDynamicBufferSize() GenConfigOption {
//	return genConfOptionFn(func(gConf *sliceGeneratorConfig) {
//		gConf.isDynamicBufferSize = true
//	})
//}
